package squashfs

import (
	"io/fs"
	"time"

	"github.com/CalebQ42/squashfs/low/directory"
	"github.com/CalebQ42/squashfs/low/inode"
)

type fileInfo struct {
	name     string
	size     int64
	perm     uint32
	modTime  uint32
	fileType uint16
}

func (r Reader) newFileInfo(e directory.Entry) (fileInfo, error) {
	i, err := r.Low.InodeFromEntry(e)
	if err != nil {
		return fileInfo{}, err
	}
	return newFileInfo(e.Name, &i), nil
}

func newFileInfo(name string, i *inode.Inode) fileInfo {
	var size int64
	switch i.Type {
	case inode.Fil:
		size = int64(i.Data.(inode.File).Size)
	case inode.EFil:
		size = int64(i.Data.(inode.EFile).Size)
	}
	return fileInfo{
		name:     name,
		size:     size,
		perm:     uint32(i.Perm),
		modTime:  i.ModTime,
		fileType: i.Type,
	}
}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) Mode() fs.FileMode {
	if f.IsDir() {
		return fs.FileMode(f.perm | uint32(fs.ModeDir))
	}
	if f.IsSymlink() {
		return fs.FileMode(f.perm | uint32(fs.ModeSymlink))
	}
	return fs.FileMode(f.perm)
}

func (f fileInfo) ModTime() time.Time {
	return time.Unix(int64(f.modTime), 0)
}

func (f fileInfo) IsDir() bool {
	return f.fileType == inode.Dir || f.fileType == inode.EDir
}

func (f fileInfo) IsSymlink() bool {
	return f.fileType == inode.Sym || f.fileType == inode.ESym
}

func (f fileInfo) Sys() any {
	return nil
}
