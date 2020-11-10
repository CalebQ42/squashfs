package squashfs

import (
	"encoding/binary"
	"io"
)

//Squashfs is a squashfs backed by a ReadSeeker.
type Squashfs struct {
	rdr   *io.ReaderAt //underlying reader
	super Superblock
}

//NewSquashfs creates a new Squashfs backed by the given reader
func NewSquashfs(reader io.ReaderAt) (*Squashfs, error) {
	var superblock Superblock
	err := binary.Read(io.NewSectionReader(reader, 0, int64(binary.Size(superblock))), binary.LittleEndian, &superblock)
	if err != nil {
		return nil, err
	}
	return &Squashfs{
		rdr:   &reader,
		super: superblock,
	}, nil
}

//GetFlags return the SuperblockFlags of the Squashfs
func (s *Squashfs) GetFlags() SuperblockFlags {
	return s.super.GetFlags()
}
