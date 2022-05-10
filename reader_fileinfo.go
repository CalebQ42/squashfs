package squashfs

import (
	"io/fs"
	"time"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

type FileInfo struct {
	e       directory.Entry
	size    int64
	perm    uint32
	modTime uint32
}

func (r Reader) newFileInfo(e directory.Entry) (FileInfo, error) {
	i, err := r.inodeFromDir(e)
	if err != nil {
		return FileInfo{}, err
	}
	return newFileInfo(e, i), nil
}

func newFileInfo(e directory.Entry, i inode.Inode) FileInfo {
	var size int64
	if i.Type == inode.Fil {
		size = int64(i.Data.(inode.File).Size)
	} else if i.Type == inode.EFil {
		size = int64(i.Data.(inode.EFile).Size)
	}
	return FileInfo{
		e:       e,
		size:    size,
		perm:    uint32(i.Perm),
		modTime: i.ModTime,
	}
}

func (f FileInfo) Name() string {
	return f.e.Name
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) Mode() fs.FileMode {
	if f.IsDir() {
		return fs.FileMode(f.perm | uint32(fs.ModeDir))
	}
	return fs.FileMode(f.perm)
}

func (f FileInfo) ModTime() time.Time {
	return time.Unix(int64(f.modTime), 0)
}

func (f FileInfo) IsDir() bool {
	return f.e.Type == inode.Dir
}

func (f FileInfo) Sys() any {
	//TODO
	return nil
}
