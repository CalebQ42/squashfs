package squashfs

import (
	"encoding/binary"
	"io"
)

//Squashfs is a squashfs backed by a ReadSeeker.
type Squashfs struct {
	rdr   *io.ReadSeeker //underlying reader
	super Superblock
}

//NewSquashfs creates a new Squashfs backed by the given reader
func NewSquashfs(reader io.ReadSeeker) (*Squashfs, error) {
	var superblock Superblock
	err := binary.Read(reader, binary.LittleEndian, &superblock)
	if err != nil {
		return nil, err
	}
	//TODO: check magic
	//TODO: parse more info
	return &Squashfs{
		rdr:   &reader,
		super: superblock,
	}, nil
}

//GetFlags return the SuperblockFlags of the Squashfs
func (s *Squashfs) GetFlags() SuperblockFlags {
	return s.super.GetFlags()
}
