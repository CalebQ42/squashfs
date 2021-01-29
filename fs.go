package squashfs

import (
	"io/fs"
	"path"
	"strings"

	"github.com/CalebQ42/squashfs/internal/directory"
)

type FS struct {
	r       *Reader
	parent  *FS
	entries []*directory.Entry
}

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

func (f FS) Glob(pattern string) ([]string, error) {
	return nil, nil
}

func (f FS) ReadDir(name string) ([]DirEntry, error) {
	return nil, nil
}

func (f FS) ReadFile(name string) ([]byte, error) {
	return nil, nil
}

func (f FS) Stat(name string) ([]byte, error) {
	return nil, nil
}

func (f FS) Sub(dir string) (fs.FS, error) {
	return nil, nil
}
