package data

import (
	"errors"
	"io"
	"runtime"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type FullReader struct {
	fileSize       uint64
	blockSize      uint32
	goroutineLimit uint16
	rdr            io.ReaderAt
	decomp         decompress.Decompressor
	sizes          []uint32
	blockOffsets   []uint64
	fragDat        []byte
}

func NewFullReader(rdr io.ReaderAt, decomp decompress.Decompressor, blockSize uint32, size uint64, start uint64, sizes []uint32) FullReader {
	out := FullReader{
		fileSize:  size,
		blockSize: blockSize,
		rdr:       rdr,
		decomp:    decomp,
		sizes:     sizes,
	}
	out.blockOffsets = make([]uint64, len(sizes))
	curOffset := start
	for i := range sizes {
		out.blockOffsets[i] = curOffset
		curOffset += uint64(sizes[i]) &^ (1 << 24)
	}
	return out
}

func (f *FullReader) Close() error {
	f.fragDat = nil
	f.sizes = nil
	f.blockOffsets = nil
	return nil
}

func (f *FullReader) AddFragData(blockStart uint64, blockSize uint32, offset uint32) error {
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

func (f *FullReader) SetGoroutineLimit(limit uint16) {
	f.goroutineLimit = limit
}

// The number of blocks, including the fragment block if present
func (f FullReader) BlockNum() uint32 {
	out := len(f.sizes)
	if f.fragDat != nil {
		out++
	}
	return uint32(out)
}

// Returns the data block at the given index
func (f FullReader) Block(i uint32) ([]byte, error) {
	if i == uint32(len(f.sizes)) && f.fragDat != nil {
		return f.fragDat, nil
	}
	if i >= uint32(len(f.sizes)) {
		return nil, errors.New("invalid block index")
	}
	realSize := f.sizes[i] &^ (1 << 24)
	if realSize == 0 {
		if i == uint32(len(f.sizes)-1) && f.fragDat == nil {
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

type blockResults struct {
	idx   uint32
	block []byte
	err   error
}

func (f FullReader) WriteTo(w io.Writer) (wrote int64, err error) {
	routineLimit := f.goroutineLimit
	if routineLimit == 0 {
		routineLimit = uint16(runtime.NumCPU() / 2)
	}
	dispatchChan := make(chan struct{}, routineLimit)
	for range int(routineLimit) {
		dispatchChan <- struct{}{}
	}
	resChan := make(chan blockResults, routineLimit)
	var results map[uint32]blockResults
	if _, is := w.(io.WriterAt); !is {
		results = make(map[uint32]blockResults)
	}
	for i := range f.BlockNum() {
		go func(idx uint32) {
			_, open := <-dispatchChan
			if !open {
				resChan <- blockResults{}
				return
			}
			block, err := f.Block(idx)
			resChan <- blockResults{
				idx:   idx,
				block: block,
				err:   err,
			}
			dispatchChan <- struct{}{}
		}(i)
	}
	out := int64(0)
	errOut := make([]error, 0)
	for i := uint32(0); i < f.BlockNum(); {
		res := <-resChan
		if res.err != nil {
			close(dispatchChan)
			errOut = append(errOut, res.err)
		}
		if len(errOut) > 0 {
			i++
			continue
		}
		if wa, is := w.(io.WriterAt); is {
			_, err := wa.WriteAt(res.block, int64(res.idx)*int64(f.blockSize))
			if err != nil {
				errOut = append(errOut, err)
			} else {
				out = max(out, int64(res.idx)*int64(f.blockSize)+int64(len(res.block)))
			}
			i++
			continue
		}
		var err error
		if res.idx == i {
			_, err = w.Write(res.block)
			if err != nil {
				errOut = append(errOut, err)
			} else {
				out = max(out, int64(res.idx)*int64(f.blockSize)+int64(len(res.block)))
			}
			i++
		} else {
			results[res.idx] = res
		}
		var has bool
		for {
			res, has = results[i]
			if has {
				_, err = w.Write(res.block)
				if err != nil {
					errOut = append(errOut, err)
				} else {
					out = max(out, int64(res.idx)*int64(f.blockSize)+int64(len(res.block)))
				}
				i++
			} else {
				break
			}
		}
	}
	if len(errOut) > 0 {
		return out, errors.Join(errOut...)
	}
	return out, nil
}
