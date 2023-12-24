package squashfs

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/CalebQ42/squashfs/squashfs"
	"github.com/CalebQ42/squashfs/squashfs/data"
	"github.com/CalebQ42/squashfs/squashfs/inode"
)

// File represents a file inside a squashfs archive.
type File struct {
	b        *squashfs.Base
	full     *data.FullReader
	rdr      *data.Reader
	parent   *FS
	r        *Reader
	dirsRead int
}

func (f *File) FS() (*FS, error) {
	if !f.IsDir() {
		return nil, errors.New("not a directory")
	}
	d, err := f.b.ToDir(f.r.r)
	if err != nil {
		return nil, err
	}
	return &FS{d: d, parent: f.parent, r: f.r}, nil
}

// Closes the underlying readers.
// Further calls to Read and WriteTo will re-create the readers.
// Never returns an error.
func (f *File) Close() error {
	if f.rdr != nil {
		return f.rdr.Close()
	}
	f.rdr = nil
	f.full = nil
	return nil
}

// Returns the file the symlink points to.
// If the file isn't a symlink, or points to a file outside the archive, returns nil.
func (f *File) GetSymlinkFile() fs.File {
	if !f.IsSymlink() {
		return nil
	}
	if filepath.IsAbs(f.SymlinkPath()) {
		return nil
	}
	fil, err := f.parent.Open(f.SymlinkPath())
	if err != nil {
		return nil
	}
	return fil
}

// Returns whether the file is a directory.
func (f *File) IsDir() bool {
	return f.b.IsDir()
}

// Returns whether the file is a regular file.
func (f *File) IsRegular() bool {
	return f.b.IsRegular()
}

// Returns whether the file is a symlink.
func (f *File) IsSymlink() bool {
	return f.b.Inode.Type == inode.Sym || f.b.Inode.Type == inode.ESym
}

func (f *File) Mode() fs.FileMode {
	return f.b.Inode.Mode()
}

// Read reads the data from the file. Only works if file is a normal file.
func (f *File) Read(b []byte) (int, error) {
	if !f.IsRegular() {
		return 0, errors.New("file is not a regular file")
	}
	if f.rdr == nil {
		err := f.initializeReaders()
		if err != nil {
			return 0, err
		}
	}
	return f.rdr.Read(b)
}

// ReadDir returns n fs.DirEntry's that's contained in the File (if it's a directory).
// If n <= 0 all fs.DirEntry's are returned.
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, errors.New("file is not a directory")
	}
	d, err := f.b.ToDir(f.r.r)
	if err != nil {
		return nil, err
	}
	start, end := 0, len(d.Entries)
	if n > 0 {
		start, end = f.dirsRead, f.dirsRead+n
		if end > len(d.Entries) {
			end = len(d.Entries)
			err = io.EOF
		}
	}
	var out []fs.DirEntry
	var fi fileInfo
	for _, e := range d.Entries[start:end] {
		fi, err = f.r.newFileInfo(e)
		if err != nil {
			f.dirsRead += len(out)
			return out, err
		}
		out = append(out, fs.FileInfoToDirEntry(fi))
	}
	f.dirsRead += len(out)
	return out, err
}

// Returns the file's fs.FileInfo
func (f *File) Stat() (fs.FileInfo, error) {
	return newFileInfo(f.b.Name, f.b.Inode), nil
}

// SymlinkPath returns the symlink's target path. Is the File isn't a symlink, returns an empty string.
func (f *File) SymlinkPath() string {
	switch f.b.Inode.Type {
	case inode.Sym:
		return string(f.b.Inode.Data.(inode.Symlink).Target)
	case inode.ESym:
		return string(f.b.Inode.Data.(inode.ESymlink).Target)
	}
	return ""
}

// Writes all data from the file to the given writer in a multi-threaded manner.
// The underlying reader is separate
func (f *File) WriteTo(w io.Writer) (int64, error) {
	if !f.IsRegular() {
		return 0, errors.New("file is not a regular file")
	}
	if f.full == nil {
		err := f.initializeReaders()
		if err != nil {
			return 0, err
		}
	}
	return f.full.WriteTo(w)
}

func (f *File) initializeReaders() error {
	var err error
	f.rdr, f.full, err = f.b.GetRegFileReaders(f.r.r)
	return err
}

// Extract the file to the given folder. If the file is a folder, the folder's contents will be extracted to the folder.
// Uses default extraction options.
func (f *File) Extract(folder string) error {
	return f.ExtractWithOptions(folder, DefaultOptions())
}

// Extract the file to the given folder. If the file is a folder, the folder's contents will be extracted to the folder.
// Allows setting various extraction options via ExtractionOptions.
func (f *File) ExtractWithOptions(folder string, op *ExtractionOptions) error {
	//TODO
}
