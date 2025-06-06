package inode

import (
	"encoding/binary"
	"io"
	"math"
)

type File struct {
	BlockStart uint32
	FragInd    uint32
	FragOffset uint32
	Size       uint32
	BlockSizes []uint32
}

type eFileInit struct {
	BlockStart uint64
	Size       uint64
	Sparse     uint64
	LinkCount  uint32
	FragInd    uint32
	FragOffset uint32
	XattrInd   uint32
}

type EFile struct {
	eFileInit
	BlockSizes []uint32
}

func ReadFile(r io.Reader, blockSize uint32) (f File, err error) {
	dat := make([]byte, 16)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	f.BlockStart = binary.LittleEndian.Uint32(dat)
	f.FragInd = binary.LittleEndian.Uint32(dat[4:])
	f.FragOffset = binary.LittleEndian.Uint32(dat[8:])
	f.Size = binary.LittleEndian.Uint32(dat[12:])
	toRead := int(math.Floor(float64(f.Size) / float64(blockSize)))
	if f.FragInd == 0xFFFFFFFF && f.Size%blockSize > 0 {
		toRead++
	}
	dat = make([]byte, toRead*4)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	f.BlockSizes = make([]uint32, toRead)
	for i := range toRead {
		f.BlockSizes[i] = binary.LittleEndian.Uint32(dat[i*4:])
	}
	return
}

func ReadEFile(r io.Reader, blockSize uint32) (f EFile, err error) {
	err = binary.Read(r, binary.LittleEndian, &f.eFileInit)
	if err != nil {
		return
	}
	toRead := int(math.Floor(float64(f.Size) / float64(blockSize)))
	if f.FragInd == 0xFFFFFFFF && f.Size%uint64(blockSize) > 0 {
		toRead++
	}
	f.BlockSizes = make([]uint32, toRead)
	err = binary.Read(r, binary.LittleEndian, &f.BlockSizes)
	return
}
