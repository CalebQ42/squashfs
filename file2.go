package squashfs

import (
	"errors"
	"io"
	"io/fs"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

type File struct {
	i        *inode.Inode
	parent   *FS
	r        *Reader
	reader   *fileReader
	name     string
	dirsRead int
}

func (f FileInfo) File() (file *File, err error) {
	file = &File{
		name:   f.name,
		r:      f.r,
		parent: f.parent,
		i:      f.i,
	}
	file.reader, err = f.r.newFileReader(f.i)
	return
}

func (r *Reader) newFileFromDirEntry(en *directory.Entry, parent *FS) (file *File, err error) {
	file = &File{
		name:   en.Name,
		r:      r,
		parent: parent,
	}
	file.i, err = r.getInodeFromEntry(en)
	if err != nil {
		return nil, err
	}
	file.reader, err = r.newFileReader(file.i)
	return
}

func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		i:      f.i,
		name:   f.name,
		parent: f.parent,
		r:      f.r,
	}, nil
}

func (f *File) Read(p []byte) (int, error) {
	if f.i.Type == inode.FileType || f.i.Type == inode.ExtFileType {
		if f.reader == nil {
			return 0, fs.ErrClosed
		}
		return f.reader.Read(p)
	}
	return 0, errors.New("Can only read files")
}

func (f *File) WriteTo(w io.Writer) (int64, error) {
	if f.i.Type == inode.FileType || f.i.Type == inode.ExtFileType {
		if f.reader == nil {
			return 0, fs.ErrClosed
		}
		return f.reader.WriteTo(w)
	}
	return 0, errors.New("Can only read files")
}

func (f *File) Close() error {
	f.reader = nil
	return nil
}

func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, errors.New("File is not a directory")
	}
	ffs, err := f.FS()
	if err != nil {
		return nil, err
	}
	var beg, end int
	if n <= 0 {
		beg, end = 0, len(ffs.entries)
	} else {
		beg, end = f.dirsRead, f.dirsRead+n
		if end > len(ffs.entries) {
			end = len(ffs.entries)
			err = io.EOF
		}
	}
	out := make([]fs.DirEntry, end-beg)
	for i, ent := range ffs.entries[beg:end] {
		out[i] = f.r.newDirEntry(ent, ffs)
	}
	return out, err
}

func (f File) FS() (*FS, error) {
	if !f.IsDir() {
		return nil, errors.New("File is not a directory")
	}
	ents, err := f.r.readDirFromInode(f.i)
	if err != nil {
		return nil, err
	}
	return &FS{
		entries: ents,
		parent:  f.parent,
		r:       f.r,
	}, nil
}

func (f File) IsDir() bool {
	return f.i.Type == inode.DirType || f.i.Type == inode.ExtDirType
}

func (f File) Path()

//ReadDirFromInode returns a fully populated Directory from a given Inode.
//If the given inode is not a directory it returns an error.
func (r *Reader) readDirFromInode(i *inode.Inode) ([]*directory.Entry, error) {
	var offset uint32
	var metaOffset uint16
	var size uint32
	switch i.Type {
	case inode.DirType:
		offset = i.Info.(inode.Dir).DirectoryIndex
		metaOffset = i.Info.(inode.Dir).DirectoryOffset
		size = uint32(i.Info.(inode.Dir).DirectorySize)
	case inode.ExtDirType:
		offset = i.Info.(inode.ExtDir).DirectoryIndex
		metaOffset = i.Info.(inode.ExtDir).DirectoryOffset
		size = i.Info.(inode.ExtDir).DirectorySize
	default:
		return nil, errors.New("Not a directory inode")
	}
	br, err := r.newMetadataReader(int64(r.super.DirTableStart + uint64(offset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(metaOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	ents, err := directory.NewDirectory(br, size)
	if err != nil {
		return nil, err
	}
	return ents, nil
}
