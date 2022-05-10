package data

import (
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/toreader"
)

type FullReader struct {
	r         io.ReaderAt
	d         decompress.Decompressor
	fragRdr   io.Reader
	sizes     []uint32
	blockSize uint32
	start     uint64
}

func NewFullReader(r io.ReaderAt, start uint64, d decompress.Decompressor, blockSizes []uint32, blockSize uint32) *FullReader {
	return &FullReader{
		r:         r,
		start:     start,
		blockSize: blockSize,
		sizes:     blockSizes,
		d:         d,
	}
}

func (r *FullReader) AddFragment(rdr io.Reader) {
	r.fragRdr = rdr
}

type outDat struct {
	err  error
	data []byte
	i    int
}

func (r FullReader) process(index int, offset int64, out chan outDat) {
	var err error
	var dat []byte
	size := realSize(r.sizes[index])
	rdr := io.LimitReader(toreader.NewReader(r.r, offset), int64(size))
	if size == r.sizes[index] {
		rdr, err = r.d.Reader(rdr)
	}
	if err == nil {
		dat, err = io.ReadAll(rdr)
	}
	out <- outDat{
		i:    index,
		err:  err,
		data: dat,
	}
	if clr, ok := rdr.(io.Closer); ok {
		clr.Close()
	}
}

func (r FullReader) WriteTo(w io.Writer) (n int64, err error) {
	out := make(chan outDat, len(r.sizes))
	offset := r.start
	num := len(r.sizes)
	for i := 0; i < num; i++ {
		if i == num-1 && r.fragRdr != nil {
			go func() {
				dat, e := io.ReadAll(r.fragRdr)
				out <- outDat{
					i:    num,
					err:  e,
					data: dat,
				}
				if clr, ok := r.fragRdr.(io.Closer); ok {
					clr.Close()
				}
			}()
			continue
		}
		go r.process(i, int64(offset), out)
		offset += uint64(realSize(r.sizes[i]))
	}
	cache := make(map[int]outDat)
	var tmpN int
	for cur := 0; cur < num; {
		dat := <-out
		if dat.err != nil {
			err = dat.err
			return
		}
		if dat.i != cur {
			cache[dat.i] = dat
			continue
		}
		tmpN, err = w.Write(dat.data)
		n += int64(tmpN)
		if err != nil {
			return
		}
		cur++
		var ok bool
		for {
			dat, ok = cache[cur]
			if !ok {
				break
			}
			tmpN, err = w.Write(dat.data)
			n += int64(tmpN)
			if err != nil {
				return
			}
			cur++
		}
	}
	return
}
