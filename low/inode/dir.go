package inode

import (
	"encoding/binary"
	"io"
)

type Directory struct {
	BlockStart uint32
	LinkCount  uint32
	Size       uint16
	Offset     uint16
	ParentNum  uint32
}

func ReadDir(r io.Reader) (d Directory, err error) {
	dat := make([]byte, 16)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	d.BlockStart = binary.LittleEndian.Uint32(dat)
	d.LinkCount = binary.LittleEndian.Uint32(dat[4:])
	d.Size = binary.LittleEndian.Uint16(dat[8:])
	d.Offset = binary.LittleEndian.Uint16(dat[10:])
	d.ParentNum = binary.LittleEndian.Uint32(dat[12:])
	return
}

type EDirectory struct {
	LinkCount  uint32
	Size       uint32
	BlockStart uint32
	ParentNum  uint32
	IndCount   uint16
	Offset     uint16
	XattrInd   uint32
	Indexes    []DirectoryIndex
}

type DirectoryIndex struct {
	Ind      uint32
	Start    uint32
	NameSize uint32
	Name     []byte
}

func ReadEDir(r io.Reader) (d EDirectory, err error) {
	dat := make([]byte, 24)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	d.LinkCount = binary.LittleEndian.Uint32(dat)
	d.Size = binary.LittleEndian.Uint32(dat[4:])
	d.BlockStart = binary.LittleEndian.Uint32(dat[8:])
	d.ParentNum = binary.LittleEndian.Uint32(dat[12:])
	d.IndCount = binary.LittleEndian.Uint16(dat[16:])
	d.Offset = binary.LittleEndian.Uint16(dat[18:])
	d.XattrInd = binary.LittleEndian.Uint32(dat[20:])
	d.Indexes = make([]DirectoryIndex, d.IndCount)
	for i := range d.IndCount {
		dat = make([]byte, 12)
		_, err = r.Read(dat)
		if err != nil {
			return
		}
		d.Indexes[i].Ind = binary.LittleEndian.Uint32(dat)
		d.Indexes[i].Start = binary.LittleEndian.Uint32(dat[4:])
		d.Indexes[i].NameSize = binary.LittleEndian.Uint32(dat[8:])
		d.Indexes[i].Name = make([]byte, d.Indexes[i].NameSize+1)
		_, err = r.Read(d.Indexes[i].Name)
		if err != nil {
			return
		}
	}
	return
}
