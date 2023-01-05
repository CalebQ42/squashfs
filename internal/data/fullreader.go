package data

import (
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/toreader"
)

type FullReader struct {
	r         io.ReaderAt
	d         decompress.Decompressor
	fragRdr   func() (io.Reader, error)
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

func (r *FullReader) AddFragment(rdr func() (io.Reader, error)) {
	r.fragRdr = rdr
	r.sizes = append(r.sizes, 0)
}

type outDat struct {
	err  error
	data []byte
	i    int
}

func (r FullReader) process(index int, offset int64, out chan outDat) {
	var err error
	var dat []byte
	var rdr io.ReadCloser
	size := realSize(r.sizes[index])
	if size == 0 {
		out <- outDat{
			i:    index,
			err:  nil,
			data: make([]byte, r.blockSize),
		}
		return
	}
	// rdr := io.LimitReader(toreader.NewReader(r.r, offset), int64(size))
	if size == r.sizes[index] {
		if dec, ok := r.d.(decompress.Decoder); ok {
			dat = make([]byte, size)
			_, err = r.r.ReadAt(dat, offset)
			if err == nil {
				dat, err = dec.Decode(dat)
			}
		} else {
			rdr, err = r.d.Reader(io.LimitReader(toreader.NewReader(r.r, offset), int64(size)))
			if err == nil {
				dat, err = io.ReadAll(rdr)
			}
		}
	} else {
		dat = make([]byte, size)
		_, err = r.r.ReadAt(dat, offset)
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

func (r FullReader) ReadAt(p []byte, off int64) (n int, err error) {
	out := make(chan outDat, len(r.sizes))
	offset := r.start
	num := len(r.sizes)
	start := off / int64(r.blockSize)
	end := len(p) / int(r.blockSize)
	if end%int(r.blockSize) > 0 {
		end++
	}
	if end > len(r.sizes) {
		if r.fragRdr != nil {
			end = len(r.sizes)
		} else {
			end = len(r.sizes) + 1
		}
	}
	for i := 0; i < num; i++ {
		if i < int(start) || i > end {
			offset += uint64(realSize(r.sizes[i]))
			continue
		}
		if i == num-1 && r.fragRdr != nil {
			go func() {
				rdr, e := r.fragRdr()
				if err != nil {
					out <- outDat{
						i:   num - 1,
						err: e,
					}
					return
				}
				dat, e := io.ReadAll(rdr)
				out <- outDat{
					i:    num - 1,
					err:  e,
					data: dat,
				}
				if clr, ok := rdr.(io.Closer); ok {
					clr.Close()
				}
			}()
			continue
		}
		go r.process(i, int64(offset), out)
		offset += uint64(realSize(r.sizes[i]))
	}
	cache := make(map[int]outDat)
	for cur := start; cur < int64(end); {
		dat := <-out
		if dat.err != nil {
			err = dat.err
			return
		}
		if dat.i != int(cur) {
			cache[dat.i] = dat
			continue
		}
		if cur == start {
			dat.data = dat.data[off%int64(r.blockSize):]
		}
		for i := range dat.data {
			p[n+i] = dat.data[i]
		}
		n += len(dat.data)
		cur++
		var ok bool
		for {
			dat, ok = cache[int(cur)]
			if !ok {
				break
			}
			for i := range dat.data {
				p[n+i] = dat.data[i]
			}
			n += len(dat.data)
			cur++
			delete(cache, int(cur))
		}
	}
	if n < len(p) {
		err = io.EOF
	}
	return
}

func (r FullReader) WriteTo(w io.Writer) (n int64, err error) {
	out := make(chan outDat, len(r.sizes))
	offset := r.start
	num := len(r.sizes)
	for i := 0; i < num; i++ {
		if i == num-1 && r.fragRdr != nil {
			go func() {
				rdr, e := r.fragRdr()
				if err != nil {
					out <- outDat{
						i:   num - 1,
						err: e,
					}
					return
				}
				dat, e := io.ReadAll(rdr)
				out <- outDat{
					i:    num - 1,
					err:  e,
					data: dat,
				}
				if clr, ok := rdr.(io.Closer); ok {
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
