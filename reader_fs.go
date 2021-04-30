package squashfs

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/squashfs/internal/directory"
)

//FS is a fs.FS representation of a squashfs directory.
//Implements fs.GlobFS, fs.ReadDirFS, fs.ReadFileFS, fs.StatFS, and fs.SubFS
type FS struct {
	r       *Reader
	parent  *FS
	name    string
	entries []*directory.Entry
}

//Open opens the file at name. Returns a squashfs.File.
func (f FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = path.Clean(strings.TrimPrefix(name, "/"))
	split := strings.Split(name, "/")
	if split[0] == ".." {
		if f.parent == nil {
			//This should only happen on the root FS
			return nil, &fs.PathError{
				Op:   "open",
				Path: name,
				//TODO: make error clearer
				Err: errors.New("trying to get file outside of squashfs"),
			}
		}
		return f.parent.Open(strings.Join(split[1:], "/"))
	}
	for i := 0; i < len(f.entries); i++ {
		if match, _ := path.Match(split[0], f.entries[i].Name); match {
			if len(split) == 1 {
				return f.r.newFileFromDirEntry(f.entries[i], &f)
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					pathErr.Op = "open"
					pathErr.Path = name
					return nil, err
				}
				return nil, &fs.PathError{
					Op:   "open",
					Path: name,
					Err:  err,
				}
			}
			fil, err := sub.Open(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "open"
					pathErr.Path = name
					return nil, err
				}
				return nil, &fs.PathError{
					Op:   "open",
					Path: name,
					Err:  err,
				}
			}
			return fil, nil
		}
	}
	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

//Glob returns the name of the files at the given pattern.
//All paths are relative to the FS.
func (f FS) Glob(pattern string) (out []string, err error) {
	if !fs.ValidPath(pattern) {
		return nil, &fs.PathError{
			Op:   "glob",
			Path: pattern,
			Err:  fs.ErrInvalid,
		}
	}
	pattern = path.Clean(strings.TrimPrefix(pattern, "/"))
	split := strings.Split(pattern, "/")
	if split[0] == ".." {
		if f.parent == nil {
			//This should only happen on the root FS
			return nil, &fs.PathError{
				Op:   "readdir",
				Path: pattern,
				//TODO: make error clearer
				Err: errors.New("trying to get file outside of squashfs"),
			}
		}
		return f.parent.Glob(strings.Join(split[1:], "/"))
	}
	for i := 0; i < len(f.entries); i++ {
		if match, _ := path.Match(split[0], f.entries[i].Name); match {
			if len(split) == 1 {
				out = append(out, f.entries[i].Name)
				continue
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "glob"
					pathErr.Path = pattern
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "glob",
					Path: pattern,
					Err:  err,
				}
			}
			subGlob, err := sub.(FS).Glob(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "glob"
					pathErr.Path = pattern
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "glob",
					Path: pattern,
					Err:  err,
				}
			}
			for i := 0; i < len(subGlob); i++ {
				subGlob[i] = f.name + "/" + subGlob[i]
			}
			out = append(out, subGlob...)
		}
	}
	return
}

//ReadDir returns all the DirEntry returns all DirEntry's for the directory at name.
//If name is not a directory, returns an error.
func (f FS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = path.Clean(strings.TrimPrefix(name, "/"))
	split := strings.Split(name, "/")
	if split[0] == ".." {
		if f.parent == nil {
			//This should only happen on the root FS
			return nil, &fs.PathError{
				Op:   "readdir",
				Path: name,
				//TODO: make error clearer
				Err: errors.New("trying to get file outside of squashfs"),
			}
		}
		return f.parent.ReadDir(strings.Join(split[1:], "/"))
	}
	for i := 0; i < len(f.entries); i++ {
		if match, _ := path.Match(split[0], f.entries[i].Name); match {
			if len(split) == 1 {
				in, err := f.r.getInodeFromEntry(f.entries[i])
				if err != nil {
					return nil, &fs.PathError{
						Op:   "readdir",
						Path: name,
						Err:  err,
					}
				}
				ents, err := f.r.readDirFromInode(in)
				if err != nil {
					return nil, &fs.PathError{
						Op:   "readdir",
						Path: name,
						Err:  err,
					}
				}
				out := make([]fs.DirEntry, len(f.entries))
				for i, ent := range ents {
					out[i] = &DirEntry{
						en:     ent,
						parent: &f,
						r:      f.r,
					}
				}
				return out, nil
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "readir"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "readdir",
					Path: name,
					Err:  err,
				}
			}
			redDir, err := sub.(FS).ReadDir(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "readdir"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "readdir",
					Path: name,
					Err:  err,
				}
			}
			return redDir, nil
		}
	}
	return nil, &fs.PathError{
		Op:   "readdir",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

//ReadFile returns the data (in []byte) for the file at name.
func (f FS) ReadFile(name string) ([]byte, error) {
	fil, err := f.Open(name)
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok {
			pathErr.Op = "readfile"
			pathErr.Path = name
			return nil, pathErr
		}
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, fil)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: name,
			Err:  err,
		}
	}
	return buf.Bytes(), nil
}

//Stat returns the fs.FileInfo for the file at name.
func (f FS) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = path.Clean(strings.TrimPrefix(name, "/"))
	split := strings.Split(name, "/")
	if split[0] == ".." {
		if f.parent == nil {
			//This should only happen on the root FS
			return nil, &fs.PathError{
				Op:   "stat",
				Path: name,
				//TODO: make error clearer
				Err: errors.New("trying to get file outside of squashfs"),
			}
		}
		return f.parent.Stat(strings.Join(split[1:], "/"))
	}
	for i := 0; i < len(f.entries); i++ {
		if match, _ := path.Match(split[0], f.entries[i].Name); match {
			if len(split) == 1 {
				in, err := f.r.getInodeFromEntry(f.entries[i])
				if err != nil {
					return nil, &fs.PathError{
						Op:   "stat",
						Path: name,
						Err:  err,
					}
				}
				return FileInfo{
					i:      in,
					parent: &f,
					r:      f.r,
					name:   f.entries[i].Name,
				}, nil
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "stat"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "stat",
					Path: name,
					Err:  err,
				}
			}
			stat, err := sub.(FS).Stat(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "stat"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "stat",
					Path: name,
					Err:  err,
				}
			}
			return stat, nil
		}
	}
	return nil, &fs.PathError{
		Op:   "stat",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

//Sub returns the FS at dir
func (f FS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, &fs.PathError{
			Op:   "sub",
			Path: dir,
			Err:  fs.ErrInvalid,
		}
	}
	dir = path.Clean(strings.TrimPrefix(dir, "/"))
	split := strings.Split(dir, "/")
	if split[0] == ".." {
		if f.parent == nil {
			//This should only happen on the root FS
			return nil, &fs.PathError{
				Op:   "sub",
				Path: dir,
				//TODO: make error clearer
				Err: errors.New("trying to get file outside of squashfs"),
			}
		}
		return f.parent.Sub(strings.Join(split[1:], "/"))
	}
	for i := 0; i < len(f.entries); i++ {
		if match, _ := path.Match(split[0], f.entries[i].Name); match {
			if len(split) == 1 {
				in, err := f.r.getInodeFromEntry(f.entries[i])
				if err != nil {
					return nil, &fs.PathError{
						Op:   "sub",
						Path: dir,
						Err:  err,
					}
				}
				ents, err := f.r.readDirFromInode(in)
				if err != nil {
					return nil, &fs.PathError{
						Op:   "sub",
						Path: dir,
						Err:  err,
					}
				}
				return &FS{
					r:       f.r,
					parent:  &f,
					name:    f.entries[i].Name,
					entries: ents,
				}, nil
			}
			sub, err := f.Sub(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "sub"
					pathErr.Path = dir
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "sub",
					Path: dir,
					Err:  err,
				}
			}
			return sub, nil
		}
	}
	return nil, &fs.PathError{
		Op:   "sub",
		Path: dir,
		Err:  fs.ErrNotExist,
	}
}

func (f FS) path() string {
	if f.name == "/" {
		return f.name
	} else if f.parent.name == "/" {
		return f.name
	}
	return f.parent.path() + "/" + f.name
}

//ExtractTo extracts the File to the given folder with the default options.
//It extracts the directory's contents to the folder.
func (f FS) ExtractTo(folder string) error {
	return f.ExtractWithOptions(folder, DefaultOptions())
}

//ExtractSymlink extracts the File to the folder with the DereferenceSymlink option.
//It extracts the directory's contents to the folder.
func (f FS) ExtractSymlink(folder string) error {
	return f.ExtractWithOptions(folder, ExtractionOptions{
		DereferenceSymlink: true,
		FolderPerm:         fs.ModePerm,
	})
}

//ExtractWithOptions extracts the File to the given folder with the given ExtrationOptions.
//It extracts the directory's contents to the folder.
func (f FS) ExtractWithOptions(folder string, op ExtractionOptions) error {
	op.notBase = true
	folder = path.Clean(folder)
	err := os.MkdirAll(folder, op.FolderPerm)
	if err != nil {
		return err
	}
	errChan := make(chan error)
	for i := 0; i < len(f.entries); i++ {
		go func(ent *DirEntry) {
			fil, goErr := ent.File()
			if goErr != nil {
				errChan <- goErr
				return
			}
			errChan <- fil.ExtractWithOptions(folder, op)
			fil.Close()
		}(&DirEntry{
			en:     f.entries[i],
			parent: &f,
			r:      f.r,
		})
	}
	for i := 0; i < len(f.entries); i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}
	return nil
}
