package squashfs

import (
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/metadata"
	"github.com/CalebQ42/squashfs/internal/toreader"
	"github.com/CalebQ42/squashfs/squashfs/data"
	"github.com/CalebQ42/squashfs/squashfs/directory"
	"github.com/CalebQ42/squashfs/squashfs/inode"
)

type Base struct {
	Inode *inode.Inode
	Name  string
}

func (r *Reader) BaseFromInode(i *inode.Inode, name string) *Base {
	return &Base{Inode: i, Name: name}
}

func (r *Reader) BaseFromEntry(e directory.Entry) (*Base, error) {
	rdr := metadata.NewReader(toreader.NewReader(r.r, int64(r.Superblock.InodeTableStart)+int64(e.BlockStart)), r.d)
	defer rdr.Close()
	rdr.Read(make([]byte, e.Offset))
	in, err := inode.Read(rdr, r.Superblock.BlockSize)
	if err != nil {
		return nil, err
	}
	return &Base{Inode: in, Name: e.Name}, nil
}

func (r *Reader) BaseFromRef(ref uint64, name string) (*Base, error) {
	in, err := r.inodeFromRef(ref)
	if err != nil {
		return nil, err
	}
	return &Base{Inode: in, Name: name}, nil
}

func (b *Base) Uid(r *Reader) (uint32, error) {
	return r.Id(b.Inode.UidInd)
}

func (b *Base) Gid(r *Reader) (uint32, error) {
	return r.Id(b.Inode.GidInd)
}

func (b *Base) IsDir() bool {
	return b.Inode.Type == inode.Dir || b.Inode.Type == inode.EDir
}

func (b *Base) ToDir(r *Reader) (*Directory, error) {
	var blockStart uint32
	var size uint32
	var offset uint16
	switch b.Inode.Type {
	case inode.Dir:
		blockStart = b.Inode.Data.(inode.Directory).BlockStart
		size = uint32(b.Inode.Data.(inode.Directory).Size)
		offset = b.Inode.Data.(inode.Directory).Offset
	case inode.EDir:
		blockStart = b.Inode.Data.(inode.EDirectory).BlockStart
		size = b.Inode.Data.(inode.EDirectory).Size
		offset = b.Inode.Data.(inode.EDirectory).Offset
	default:
		return nil, errors.New("not a directory")
	}
	dirRdr := metadata.NewReader(toreader.NewReader(r.r, int64(r.Superblock.DirTableStart)+int64(blockStart)), r.d)
	defer dirRdr.Close()
	_, err := dirRdr.Read(make([]byte, offset))
	if err != nil {
		return nil, err
	}
	entries, err := directory.ReadDirectory(dirRdr, size)
	if err != nil {
		return nil, err
	}
	return &Directory{
		Base:    *b,
		Entries: entries,
	}, nil
}

func (b *Base) IsRegular() bool {
	return b.Inode.Type == inode.Fil || b.Inode.Type == inode.EFil
}

func (b *Base) GetRegFileReaders(r *Reader) (*data.Reader, *data.FullReader, error) {
	if !b.IsRegular() {
		return nil, nil, errors.New("not a regular file")
	}
	var blockStart uint64
	var fragIndex uint32
	var fragOffset uint32
	var fragSize uint64
	var sizes []uint32
	if b.Inode.Type == inode.Fil {
		blockStart = uint64(b.Inode.Data.(inode.File).BlockStart)
		fragIndex = b.Inode.Data.(inode.File).FragInd
		fragOffset = b.Inode.Data.(inode.File).FragOffset
		sizes = b.Inode.Data.(inode.File).BlockSizes
		fragSize = uint64(b.Inode.Data.(inode.File).Size % r.Superblock.BlockSize)
	} else {
		blockStart = b.Inode.Data.(inode.EFile).BlockStart
		fragIndex = b.Inode.Data.(inode.EFile).FragInd
		fragOffset = b.Inode.Data.(inode.EFile).FragOffset
		sizes = b.Inode.Data.(inode.EFile).BlockSizes
		fragSize = b.Inode.Data.(inode.EFile).Size % uint64(r.Superblock.BlockSize)
	}
	frag := func() (io.Reader, error) {
		ent, err := r.fragEntry(fragIndex)
		if err != nil {
			return nil, err
		}
		frag := data.NewReader(toreader.NewReader(r.r, int64(ent.Start)), r.d, []uint32{ent.Size}, uint64(r.Superblock.BlockSize), r.Superblock.BlockSize)
		frag.Read(make([]byte, fragOffset))
		return io.LimitReader(frag, int64(fragSize)), nil
	}
	outRdr := data.NewReader(toreader.NewReader(r.r, int64(blockStart)), r.d, sizes, fragSize, r.Superblock.BlockSize)
	if fragIndex != 0xffffffff {
		f, err := frag()
		if err != nil {
			return nil, nil, err
		}
		outRdr.AddFrag(f)
	}
	outFull := data.NewFullReader(r.r, int64(blockStart), r.d, sizes, fragSize, r.Superblock.BlockSize)
	if fragIndex != 0xffffffff {
		outFull.AddFrag(frag)
	}
	return outRdr, outFull, nil
}
