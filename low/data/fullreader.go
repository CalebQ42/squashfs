package data

import (
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type FullReader struct {
	fileSize     uint64
	blockSize    uint32
	rdr          io.ReaderAt
	decomp       decompress.Decompressor
	sizes        []uint32
	blockOffsets []uint64
	fragDat      []byte
}

func NewFullReader(rdr io.ReaderAt, decomp decompress.Decompressor, size uint64, start uint64, blockSizes []uint32) FullReader {
	out := FullReader{
		fileSize: size,
		rdr:      rdr,
		decomp:   decomp,
		sizes:    blockSizes,
	}
	out.blockOffsets = make([]uint64, len(blockSizes))
	curOffset := start
	for i := range blockSizes {
		out.blockOffsets[i] = curOffset
		curOffset += uint64(blockSizes[i]) &^ (1 << 24)
	}
	return out
}

func (f *FullReader) AddFragData(blockStart uint64, offset uint32, blockSize uint32) error {
	realSize := blockSize &^ (1 << 24)
	dat := make([]byte, realSize)
	_, err := f.rdr.ReadAt(dat, int64(blockStart))
	if err != nil {
		return err
	}
	if blockSize == realSize {
		dat, err = f.decomp.Decompress(dat)
		if err != nil {
			return err
		}
	}
	f.fragDat = dat[offset : offset+uint32(f.fileSize%uint64(f.blockSize))]
	return nil
}

// Returns the data block at the given index
func (f FullReader) Block(i int) ([]byte, error) {
	if i == len(f.sizes) && f.fragDat != nil {
		return f.fragDat, nil
	}
	if i >= len(f.sizes) {
		return nil, errors.New("invalid block index")
	}
	realSize := f.sizes[i] &^ (1 << 24)
	if realSize == 0 {
		if i == len(f.sizes)-1 && f.fragDat == nil {
			return make([]byte, f.fileSize%uint64(f.blockSize)), nil
		}
		return make([]byte, f.blockSize), nil
	}
	dat := make([]byte, realSize)
	_, err := f.rdr.ReadAt(dat, int64(f.blockOffsets[i]))
	if err != nil {
		return nil, err
	}
	if realSize == f.sizes[i] {
		return f.decomp.Decompress(dat)
	}
	return dat, nil
}

func (f FullReader) WriteTo(w io.Writer) (int64, error) {

}
