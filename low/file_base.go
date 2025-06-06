package squashfslow

import (
	"errors"

	"github.com/CalebQ42/squashfs/internal/metadata"
	"github.com/CalebQ42/squashfs/internal/toreader"
	"github.com/CalebQ42/squashfs/low/data"
	"github.com/CalebQ42/squashfs/low/directory"
	"github.com/CalebQ42/squashfs/low/inode"
)

type FileBase struct {
	Inode inode.Inode
	Name  string
}

func (r Reader) BaseFromInode(i inode.Inode, name string) FileBase {
	return FileBase{Inode: i, Name: name}
}

func (r Reader) BaseFromEntry(e directory.Entry) (FileBase, error) {
	in, err := r.InodeFromEntry(e)
	if err != nil {
		return FileBase{}, err
	}
	return FileBase{Inode: in, Name: e.Name}, nil
}

func (r Reader) BaseFromRef(ref uint64, name string) (FileBase, error) {
	in, err := r.InodeFromRef(ref)
	if err != nil {
		return FileBase{}, err
	}
	return FileBase{Inode: in, Name: name}, nil
}

func (b FileBase) Uid(r *Reader) (uint32, error) {
	return r.Id(b.Inode.UidInd)
}

func (b FileBase) Gid(r *Reader) (uint32, error) {
	return r.Id(b.Inode.GidInd)
}

func (b FileBase) IsDir() bool {
	return b.Inode.Type == inode.Dir || b.Inode.Type == inode.EDir
}

func (b FileBase) ToDir(r Reader) (Directory, error) {
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
		return Directory{}, errors.New("not a directory")
	}
	dirRdr := metadata.NewReader(toreader.NewReader(r.r, int64(r.Superblock.DirTableStart)+int64(blockStart)), r.d)
	defer dirRdr.Close()
	_, err := dirRdr.Read(make([]byte, offset))
	if err != nil {
		return Directory{}, err
	}
	entries, err := directory.ReadDirectory(&dirRdr, size)
	if err != nil {
		return Directory{}, err
	}
	return Directory{
		FileBase: b,
		Entries:  entries,
	}, nil
}

func (b FileBase) IsRegular() bool {
	return b.Inode.Type == inode.Fil || b.Inode.Type == inode.EFil
}

// Returns a regular file's readers. They are linked, so the data.Reader calls to the data.FullReader.
// Aka: closing the FullReader breaks the Reader
func (b FileBase) GetRegFileReaders(r Reader) (data.Reader, data.FullReader, error) {
	if !b.IsRegular() {
		return data.Reader{}, data.FullReader{}, errors.New("not a regular file")
	}
	var blockStart uint64
	var fragIndex uint32
	var fragOffset uint32
	var sizes []uint32
	var fileSize uint64
	if b.Inode.Type == inode.Fil {
		blockStart = uint64(b.Inode.Data.(inode.File).BlockStart)
		fragIndex = b.Inode.Data.(inode.File).FragInd
		fragOffset = b.Inode.Data.(inode.File).FragOffset
		sizes = b.Inode.Data.(inode.File).BlockSizes
		fileSize = uint64(b.Inode.Data.(inode.File).Size)
	} else {
		blockStart = b.Inode.Data.(inode.EFile).BlockStart
		fragIndex = b.Inode.Data.(inode.EFile).FragInd
		fragOffset = b.Inode.Data.(inode.EFile).FragOffset
		sizes = b.Inode.Data.(inode.EFile).BlockSizes
		fileSize = b.Inode.Data.(inode.EFile).Size
	}
	outFull := data.NewFullReader(r.r, r.d, r.Superblock.BlockSize, fileSize, blockStart, sizes)
	if fragIndex != 0xFFFFFFFF {
		ent, err := r.fragEntry(fragIndex)
		if err != nil {
			return data.Reader{}, data.FullReader{}, err
		}
		outFull.AddFragData(ent.Start, ent.Size, fragOffset)
	}
	outRdr, err := data.NewReader(&outFull)
	if err != nil {
		return data.Reader{}, data.FullReader{}, err
	}
	return outRdr, outFull, nil
}

func (b FileBase) GetFullReader(r *Reader) (data.FullReader, error) {
	if !b.IsRegular() {
		return data.FullReader{}, errors.New("not a regular file")
	}
	var blockStart uint64
	var fragIndex uint32
	var fragOffset uint32
	var sizes []uint32
	var fileSize uint64
	if b.Inode.Type == inode.Fil {
		blockStart = uint64(b.Inode.Data.(inode.File).BlockStart)
		fragIndex = b.Inode.Data.(inode.File).FragInd
		fragOffset = b.Inode.Data.(inode.File).FragOffset
		sizes = b.Inode.Data.(inode.File).BlockSizes
		fileSize = uint64(b.Inode.Data.(inode.File).Size)
	} else {
		blockStart = b.Inode.Data.(inode.EFile).BlockStart
		fragIndex = b.Inode.Data.(inode.EFile).FragInd
		fragOffset = b.Inode.Data.(inode.EFile).FragOffset
		sizes = b.Inode.Data.(inode.EFile).BlockSizes
		fileSize = b.Inode.Data.(inode.EFile).Size
	}
	outFull := data.NewFullReader(r.r, r.d, r.Superblock.BlockSize, fileSize, blockStart, sizes)
	if fragIndex != 0xFFFFFFFF {
		ent, err := r.fragEntry(fragIndex)
		if err != nil {
			return data.FullReader{}, err
		}
		outFull.AddFragData(ent.Start, ent.Size, fragOffset)
	}
	return outFull, nil
}
