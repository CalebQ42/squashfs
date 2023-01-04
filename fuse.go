package squashfs

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/CalebQ42/fuse"
	"github.com/CalebQ42/fuse/fs"
	"github.com/CalebQ42/squashfs/internal/inode"
)

// Mounts the archive to the given mountpoint using fuse3.
// Blocks until the arhive is unmounted.
// Hightly suggested to run in a goroutine.
// Will take a moment before MountWait and Unmount will work correctly.
func (r *Reader) Mount(mountpoint string) (err error) {
	if r.con != nil {
		return errors.New("squashfs archive already mounted")
	}
	r.con, err = fuse.Mount(mountpoint, fuse.ReadOnly())
	if err != nil {
		return
	}
	err = fs.Serve(r.con, &squashFuse{r: r})
	return
}

// Blocks until the mount ends.
func (r *Reader) MountWait() {
	if r.con != nil {
		<-r.con.Ready
	}
}

// Unmounts the archive.
func (r *Reader) Unmount() error {
	if r.con != nil {
		defer func() { r.con = nil }()
		return r.con.Close()
	}
	return errors.New("squashfs archive is not mounted")
}

type squashFuse struct {
	r *Reader
}

func (s *squashFuse) Root() (fs.Node, error) {
	return &fileNode{File: s.r.FS.File}, nil
}

type fileNode struct {
	*File
}

func (f *fileNode) Attr(ctx context.Context, attr *fuse.Attr) error {
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

func (f *fileNode) Id() uint64 {
	return uint64(f.i.Num)
}

func (f *fileNode) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return f.SymlinkPath(), nil
}

func (f *fileNode) Lookup(ctx context.Context, name string) (fs.Node, error) {
	asFS, err := f.FS()
	if err != nil {
		return nil, fuse.ENOTDIR
	}
	ret, err := asFS.OpenFile(name)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return &fileNode{File: ret}, nil
}

func (f *fileNode) ReadAll(ctx context.Context) ([]byte, error) {
	if f.IsRegular() {
		var buf bytes.Buffer
		_, err := f.WriteTo(&buf)
		return buf.Bytes(), err
	}
	return nil, fuse.ENODATA
}

func (f *fileNode) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if f.IsRegular() {
		buf := make([]byte, req.Size)
		n, err := f.File.ReadAt(buf, req.Offset)
		if err == io.EOF {
			resp.Data = buf[:n]
		}
		return nil
	}
	return fuse.ENODATA
}

func (f *fileNode) ReadDirAll(ctx context.Context) (out []fuse.Dirent, err error) {
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
