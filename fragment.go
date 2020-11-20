package squashfs

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

type FragmentEntryRaw struct {
	Start uint64
	Size  uint32
}

type FragmentEntry struct {
	start      uint64
	size       uint32
	compressed bool
}

//NewFragmentEntry reads a fragment entry from the given io.Reader.
func (r *Reader) NewFragmentEntry(rdr *io.Reader) (*FragmentEntry, error) {
	var entry FragmentEntry
	var raw FragmentEntryRaw
	err := binary.Read(*rdr, binary.LittleEndian, &raw)
	if err != nil {
		return nil, err
	}
	entry.start = raw.Start
	entry.compressed = raw.Size&0x1000000 == 0x1000000
	entry.size = raw.Size &^ 0x1000000
	return &entry, nil
}

//GetFragmentFromInode returns the fragment data for a given inode
func (r *Reader) GetFragmentFromInode(in *inode.Inode) ([]byte, error) {
	if in.Type != inode.BasicFileType {
		return nil, errors.New("Only basic file is supported right now")
	}
	bf := in.Info.(inode.BasicFile)
	var size uint32
	if bf.Init.BlockStart == 0 {
		size = bf.Init.Size
	} else {
		size = bf.BlockSizes
	}
}
