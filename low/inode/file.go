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

func ReadFile(r io.Reader, blockSize uint32) (f File, err error) {
	dat := make([]byte, 16)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	f.BlockStart = binary.LittleEndian.Uint32(dat[0:4])
	f.FragInd = binary.LittleEndian.Uint32(dat[4:8])
	f.FragOffset = binary.LittleEndian.Uint32(dat[8:12])
	f.Size = binary.LittleEndian.Uint32(dat[12:16])
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

type EFile struct {
	BlockStart uint64
	Size       uint64
	Sparse     uint64
	LinkCount  uint32
	FragInd    uint32
	FragOffset uint32
	XattrInd   uint32
	BlockSizes []uint32
}

func ReadEFile(r io.Reader, blockSize uint32) (f EFile, err error) {
	dat := make([]byte, 40)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	f.BlockStart = binary.LittleEndian.Uint64(dat[0:8])
	f.Size = binary.LittleEndian.Uint64(dat[8:16])
	f.Sparse = binary.LittleEndian.Uint64(dat[16:24])
	f.LinkCount = binary.LittleEndian.Uint32(dat[24:28])
	f.FragInd = binary.LittleEndian.Uint32(dat[28:32])
	f.FragOffset = binary.LittleEndian.Uint32(dat[32:36])
	f.XattrInd = binary.LittleEndian.Uint32(dat[36:40])
	toRead := f.Size / uint64(blockSize)
	if f.FragInd == 0xFFFFFFFF && f.Size%uint64(blockSize) > 0 {
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
