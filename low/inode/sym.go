package inode

import (
	"encoding/binary"
	"io"
)

type Symlink struct {
	LinkCount  uint32
	TargetSize uint32
	Target     []byte
}

func ReadSym(r io.Reader) (s Symlink, err error) {
	dat := make([]byte, 8)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	s.LinkCount = binary.LittleEndian.Uint32(dat)
	s.TargetSize = binary.LittleEndian.Uint32(dat[4:])
	s.Target = make([]byte, s.TargetSize)
	_, err = r.Read(s.Target)
	return
}

type ESymlink struct {
	LinkCount  uint32
	TargetSize uint32
	Target     []byte
	XattrInd   uint32
}

func ReadESym(r io.Reader) (s ESymlink, err error) {
	dat := make([]byte, 8)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	s.LinkCount = binary.LittleEndian.Uint32(dat)
	s.TargetSize = binary.LittleEndian.Uint32(dat[4:])
	s.Target = make([]byte, s.TargetSize)
	_, err = r.Read(s.Target)
	if err != nil {
		return
	}
	dat = make([]byte, 4)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	s.XattrInd = binary.LittleEndian.Uint32(dat)
	return
}
