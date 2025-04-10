package squashfs

import (
	"io/fs"
	"time"

	"github.com/CalebQ42/squashfs/low/directory"
	"github.com/CalebQ42/squashfs/low/inode"
)

type FileInfo struct {
	name     string
	uid      uint32
	gid      uint32
	size     int64
	target   string
	perm     uint32
	modTime  uint32
	fileType uint16
	inodeNum uint32
}

func (r Reader) newFileInfo(e directory.Entry) (FileInfo, error) {
	b, err := r.Low.BaseFromEntry(e)
	if err != nil {
		return FileInfo{}, err
	}
	uid, err := b.Uid(&r.Low)
	if err != nil {
		return FileInfo{}, err
	}
	gid, err := b.Gid(&r.Low)
	if err != nil {
		return FileInfo{}, err
	}
	return newFileInfo(e.Name, uid, gid, &b.Inode), nil
}

func newFileInfo(name string, uid, gid uint32, i *inode.Inode) FileInfo {
	var size int64
	var target string
	switch i.Type {
	case inode.Fil:
		size = int64(i.Data.(inode.File).Size)
	case inode.EFil:
		size = int64(i.Data.(inode.EFile).Size)
	case inode.Sym:
		target = string(i.Data.(inode.Symlink).Target)
	case inode.ESym:
		target = string(i.Data.(inode.ESymlink).Target)
	}
	return FileInfo{
		name:     name,
		uid:      uid,
		gid:      gid,
		size:     size,
		target:   target,
		perm:     uint32(i.Perm),
		modTime:  i.ModTime,
		fileType: i.Type,
		inodeNum: i.Num,
	}
}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Uid() int {
	return int(f.uid)
}

func (f FileInfo) Gid() int {
	return int(f.gid)
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) SymlinkPath() string {
	return f.target
}

func (f FileInfo) Inode() uint {
	return uint(f.inodeNum)
}

func (f FileInfo) Mode() fs.FileMode {
	switch f.fileType {
	case inode.Dir, inode.EDir:
		return fs.FileMode(f.perm | uint32(fs.ModeDir))
	case inode.Sym, inode.ESym:
		return fs.FileMode(f.perm | uint32(fs.ModeSymlink))
	case inode.Char, inode.EChar, inode.Block, inode.EBlock:
		return fs.FileMode(f.perm | uint32(fs.ModeDevice))
	case inode.Fifo, inode.EFifo:
		return fs.FileMode(f.perm | uint32(fs.ModeNamedPipe))
	case inode.Sock, inode.ESock:
		return fs.FileMode(f.perm | uint32(fs.ModeSocket))
	}
	return fs.FileMode(f.perm)
}

func (f FileInfo) ModTime() time.Time {
	return time.Unix(int64(f.modTime), 0)
}

func (f FileInfo) IsDir() bool {
	return f.fileType == inode.Dir || f.fileType == inode.EDir
}

func (f FileInfo) IsSymlink() bool {
	return f.fileType == inode.Sym || f.fileType == inode.ESym
}

func (f FileInfo) IsDevice() bool {
	return f.fileType == inode.Block || f.fileType == inode.EBlock ||
		f.fileType == inode.Char || f.fileType == inode.EChar
}

func (f FileInfo) IsFifo() bool {
	return f.fileType == inode.Fifo || f.fileType == inode.EFifo
}

func (f FileInfo) IsSocket() bool {
	return f.fileType == inode.Sock || f.fileType == inode.ESock
}

func (f FileInfo) Sys() any {
	return nil
}
