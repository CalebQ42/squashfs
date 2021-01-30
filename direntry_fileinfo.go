package squashfs

import (
	"io"
	"io/fs"
	"time"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

//DirEntry is a child of a directory.
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

//Name returns the DirEntry's name
func (d DirEntry) Name() string {
	return d.en.Name
}

//IsDir Yep.
func (d DirEntry) IsDir() bool {
	return d.en.Type == inode.DirType
}

//Type returns the type bits of fs.FileMode of the DirEntry.
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

//Info returns the fs.FileInfo for the given DirEntry.
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

//FileInfo is a fs.FileInfo for a file.
type FileInfo struct {
	i      *inode.Inode
	parent *FS
	r      *Reader
	name   string
}

//Name is the file's name.
func (f FileInfo) Name() string {
	return f.name
}

//Size is the file's size if it's a regular file. Otherwise, returns 0.
func (f FileInfo) Size() int64 {
	switch f.i.Type {
	case inode.FileType:
		return int64(f.i.Info.(inode.File).Size)
	case inode.ExtFileType:
		return int64(f.i.Info.(inode.ExtFile).Size)
	}
	return 0
}

//Mode returns the fs.FileMode bits of the file.
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

//ModTime is the last time the file was modified.
func (f FileInfo) ModTime() time.Time {
	return time.Unix(int64(f.i.ModifiedTime), 0)
}

//IsDir yep.
func (f FileInfo) IsDir() bool {
	return f.i.Type == inode.DirType || f.i.Type == inode.ExtDirType
}

//Sys returns the File for the FileInfo. If something goes wrong, nil is returned.
func (f FileInfo) Sys() interface{} {
	fil, err := f.File()
	if err != nil {
		return nil
	}
	return fil
}
