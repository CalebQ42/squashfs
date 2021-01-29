package squashfs

import (
	"io"
	"io/fs"
	"time"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

type DirEntry struct {
	en     *directory.Entry
	parent *FS
	r      *Reader
}

func (r *Reader) newDirEntry(en *directory.Entry, parent *FS) *DirEntry {
	return &DirEntry{
		en:     en,
		parent: parent,
		r:      r,
	}
}

func (d DirEntry) Name() string {
	return d.en.Name
}

func (d DirEntry) IsDir() bool {
	return d.en.Type == inode.DirType
}

func (d DirEntry) Type() fs.FileMode {
	switch d.en.Type {
	case inode.DirType:
		return fs.ModeDir
	case inode.SymType:
		return fs.ModeSymlink
	default:
		return 0
	}
}

func (d DirEntry) Info() (fs.FileInfo, error) {
	in, err := d.r.getInodeFromEntry(d.en)
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		name:   d.en.Name,
		i:      in,
		parent: d.parent,
		r:      d.r,
	}, nil
}

//GetInodeFromEntry returns the inode associated with a given directory.Entry
func (r *Reader) getInodeFromEntry(en *directory.Entry) (*inode.Inode, error) {
	br, err := r.newMetadataReader(int64(r.super.InodeTableStart + uint64(en.InodeOffset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(en.InodeBlockOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	i, err := inode.ProcessInode(br, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	return i, nil
}

type FileInfo struct {
	i      *inode.Inode
	parent *FS
	r      *Reader
	name   string
}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Size() int64 {
	switch f.i.Type {
	case inode.FileType:
		return int64(f.i.Info.(inode.File).Size)
	case inode.ExtFileType:
		return int64(f.i.Info.(inode.ExtFile).Size)
	}
	return 0
}

func (f FileInfo) Mode() fs.FileMode {
	mode := fs.FileMode(f.i.Permissions)
	switch f.i.Type {
	case inode.DirType | inode.ExtDirType:
		return mode | fs.ModeDir
	case inode.ExtDirType:
		return mode | fs.ModeDir
	case inode.SymType:
		return mode | fs.ModeSymlink
	case inode.ExtSymType:
		return mode | fs.ModeSymlink
	}
	return mode
}

func (f FileInfo) ModTime() time.Time {
	return time.Unix(int64(f.i.ModifiedTime), 0)
}

func (f FileInfo) IsDir() bool {
	return f.i.Type == inode.DirType || f.i.Type == inode.ExtDirType
}

func (f FileInfo) Sys() interface{} {
	return &File{
		name:   f.name,
		i:      f.i,
		r:      f.r,
		parent: f.parent,
	}
}
