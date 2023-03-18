package squashfs

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
)

// Mounts the archive to the given mountpoint using fuse2. Non-blocking.
// If Unmount does not get called, the mount point must be unmounted using umount before the directory can be used again.
func (r *Reader) MountFuse2(mountpoint string) (err error) {
	if r.con != nil {
		return errors.New("squashfs archive already mounted")
	}
	r.con2, err = fuse.Mount(mountpoint, fuse.ReadOnly())
	if err != nil {
		return
	}
	<-r.con2.Ready
	r.mount2Done = make(chan struct{})
	go func() {
		fs.Serve(r.con2, squashFuse2{r: r})
		close(r.mount2Done)
	}()
	return
}

// Blocks until the mount ends.
// Fuse2 version.
func (r *Reader) MountWaitFuse2() {
	if r.mount2Done != nil {
		<-r.mount2Done
	}
}

// Unmounts the archive.
// Fuse2 version.
func (r *Reader) UnmountFuse2() error {
	if r.con != nil {
		defer func() { r.con = nil }()
		return r.con.Close()
	}
	return errors.New("squashfs archive is not mounted")
}

type squashFuse2 struct {
	r *Reader
}

func (s squashFuse2) Root() (fs.Node, error) {
	return fileNode2{File: s.r.FS.File}, nil
}

type fileNode2 struct {
	*File
}

func (f fileNode2) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Blocks = f.r.s.Size / 512
	if f.r.s.Size%512 > 0 {
		attr.Blocks++
	}
	attr.Gid = f.r.ids[f.i.GidInd]
	attr.Inode = uint64(f.i.Num)
	attr.Mode = f.i.Mode()
	attr.Nlink = f.i.LinkCount()
	attr.Size = f.i.Size()
	attr.Uid = f.r.ids[f.i.UidInd]
	return nil
}

func (f fileNode2) Id() uint64 {
	return uint64(f.i.Num)
}

func (f fileNode2) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return f.SymlinkPath(), nil
}

func (f fileNode2) Lookup(ctx context.Context, name string) (fs.Node, error) {
	asFS, err := f.FS()
	if err != nil {
		return nil, fuse.ENOTDIR
	}
	ret, err := asFS.OpenFile(name)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return fileNode2{File: ret}, nil
}

func (f fileNode2) ReadAll(ctx context.Context) ([]byte, error) {
	if f.IsRegular() {
		var buf bytes.Buffer
		_, err := f.WriteTo(&buf)
		return buf.Bytes(), err
	}
	return nil, ENODATA
}

func (f fileNode2) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if f.IsRegular() {
		buf := make([]byte, req.Size)
		n, err := f.File.ReadAt(buf, req.Offset)
		if err == io.EOF {
			resp.Data = buf[:n]
		}
		return nil
	}
	return ENODATA
}

func (f fileNode2) ReadDirAll(ctx context.Context) (out []fuse.Dirent, err error) {
	asFS, err := f.FS()
	if err != nil {
		return nil, fuse.ENOTDIR
	}
	var t fuse.DirentType
	for i := range asFS.e {
		switch asFS.e[i].Type {
		case inode.Fil:
			t = fuse.DT_File
		case inode.Dir:
			t = fuse.DT_Dir
		case inode.Block:
			t = fuse.DT_Block
		case inode.Sym:
			t = fuse.DT_Link
		case inode.Char:
			t = fuse.DT_Char
		case inode.Fifo:
			t = fuse.DT_FIFO
		case inode.Sock:
			t = fuse.DT_Socket
		default:
			t = fuse.DT_Unknown
		}
		out = append(out, fuse.Dirent{
			Inode: uint64(asFS.e[i].Num),
			Type:  t,
			Name:  asFS.e[i].Name,
		})
	}
	return
}
