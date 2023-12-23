package squashfs

import (
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/data"
	"github.com/CalebQ42/squashfs/internal/metadata"
	"github.com/CalebQ42/squashfs/internal/toreader"
	"github.com/CalebQ42/squashfs/squashfs/directory"
	"github.com/CalebQ42/squashfs/squashfs/inode"
)

type Base struct {
	Inode *inode.Inode
	Name  string
}

func (r *Reader) baseFromInode(i *inode.Inode, name string) *Base {
	return &Base{Inode: i, Name: name}
}

func (r *Reader) baseFromEntry(e directory.Entry) (*Base, error) {
	rdr := metadata.NewReader(toreader.NewReader(r.r, int64(r.sup.InodeTableStart)+int64(e.BlockStart)), r.d)
	rdr.Read(make([]byte, e.Offset))
	in, err := inode.Read(rdr, r.sup.BlockSize)
	if err != nil {
		return nil, err
	}
	return &Base{Inode: in, Name: e.Name}, nil
}

func (b *Base) Uid(r *Reader) (uint32, error) {
	return r.id(b.Inode.UidInd)
}

func (b *Base) Gid(r *Reader) (uint32, error) {
	return r.id(b.Inode.GidInd)
}

func (b *Base) IsDir() bool {
	return b.Inode.Type == inode.Dir || b.Inode.Type == inode.EDir
}

func (b *Base) ToDir(r *Reader) (*Directory, error) {
	return r.directoryFromInode(b.Inode, b.Name)
}

func (b *Base) IsRegular() bool {
	return b.Inode.Type == inode.Fil || b.Inode.Type == inode.EFil
}

func (b *Base) GetRegFileReaders(r *Reader) (io.ReadCloser, error) {
	if !b.IsRegular() {
		return nil, errors.New("not a regular file")
	}
	var blockStart uint64
	var fragIndex uint32
	var fragOffset uint32
	var sizes []uint32
	if b.Inode.Type == inode.Fil {
		blockStart = uint64(b.Inode.Data.(inode.File).BlockStart)
		fragIndex = b.Inode.Data.(inode.File).FragInd
		fragOffset = b.Inode.Data.(inode.File).FragOffset
		sizes = b.Inode.Data.(inode.File).BlockSizes
	} else {
		blockStart = b.Inode.Data.(inode.EFile).BlockStart
		fragIndex = b.Inode.Data.(inode.EFile).FragInd
		fragOffset = b.Inode.Data.(inode.EFile).FragOffset
		sizes = b.Inode.Data.(inode.EFile).BlockSizes
	}
	var frag *data.Reader
	if fragIndex != 0xFFFFFFFF {
		ent, err := r.fragEntry(fragIndex)
		if err != nil {
			return nil, err
		}
		frag = data.NewReader(toreader.NewReader(r.r, int64(ent.start)), r.d, []uint32{ent.size})
		frag.Read(make([]byte, fragOffset))
	}
	out := data.NewReader(toreader.NewReader(r.r, int64(blockStart)), r.d, sizes)
	if frag != nil {
		out.AddFrag(out)
	}
	//TODO: implement and add full reader
	return out, nil
}
