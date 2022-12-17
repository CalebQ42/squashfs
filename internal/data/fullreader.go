package data

import (
	"io"
	"sync"

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

func (r *FullReader) AddFragment(rdr func() (io.Reader, error), size uint32) {
	r.fragRdr = rdr
	r.sizes = append(r.sizes, size)
}

type outDat struct {
	err  error
	data []byte
	i    int
}

func (r FullReader) process(index int, offset int64, od *outDat, out chan *outDat) {
	defer func() {
		out <- od
	}()
	od.i = index
	size := realSize(r.sizes[index])
	if size == 0 {
		od.err = nil
		od.data = make([]byte, r.blockSize)
		return
	}
	if size == r.sizes[index] {
		if dec, ok := r.d.(decompress.Decoder); ok {
			dat := make([]byte, size)
			_, od.err = r.r.ReadAt(dat, offset)
			if od.err != nil {
				return
			}
			od.data, od.err = dec.Decode(dat, int(r.blockSize))
			return
		}
		var rdr io.ReadCloser
		rdr, od.err = r.d.Reader(io.LimitReader(toreader.NewReader(r.r, offset), int64(size)))
		if od.err != nil {
			return
		}
		od.data = make([]byte, r.blockSize)
		var read int
		read, od.err = rdr.Read(od.data)
		od.data = od.data[:read]
		rdr.Close()
	} else {
		od.data = make([]byte, size)
		_, od.err = r.r.ReadAt(od.data, offset)
	}
}

func (r FullReader) ReadAt(p []byte, off int64) (n int, err error) {
	pol := &sync.Pool{
		New: func() any {
			return new(outDat)
		},
	}
	out := make(chan *outDat, len(r.sizes))
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
		od := pol.Get().(*outDat)
		if i == num-1 && r.fragRdr != nil {
			go func() {
				defer func() {
					out <- od
				}()
				rdr, e := r.fragRdr()
				if err != nil {
					od.i = num - 1
					od.err = e
					return
				}
				od.data = make([]byte, r.sizes[num-1])
				_, e = rdr.Read(od.data)
				od.i = num - 1
				od.err = e
				if clr, ok := rdr.(io.Closer); ok {
					clr.Close()
				}
			}()
			continue
		}
		go r.process(i, int64(offset), od, out)
		offset += uint64(realSize(r.sizes[i]))
	}
	cur := start
	cache := make(map[int]outDat)
	for dat := range out {
		if dat.err != nil {
			err = dat.err
			pol.Put(dat)
			return
		}
		if dat.i != int(cur) {
			cache[dat.i] = *dat
			pol.Put(dat)
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
		pol.Put(dat)
		var ok bool
		var curDat outDat
		for {
			curDat, ok = cache[int(cur)]
			if !ok {
				break
			}
			for i := range curDat.data {
				p[n+i] = curDat.data[i]
			}
			n += len(curDat.data)
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
	pol := &sync.Pool{
		New: func() any {
			return new(outDat)
		},
	}
	out := make(chan *outDat, len(r.sizes))
	offset := r.start
	num := len(r.sizes)
	for i := 0; i < num; i++ {
		od := pol.Get().(*outDat)
		if i == num-1 && r.fragRdr != nil {
			go func() {
				defer func() {
					out <- od
				}()
				rdr, e := r.fragRdr()
				if err != nil {
					od.i = num - 1
					od.err = e
					return
				}
				buf := make([]byte, r.sizes[num-1])
				_, e = rdr.Read(buf)
				od.i = num - 1
				od.err = e
				od.data = buf
				if clr, ok := rdr.(io.Closer); ok {
					clr.Close()
				}
			}()
			continue
		}
		go r.process(i, int64(offset), od, out)
		offset += uint64(realSize(r.sizes[i]))
	}
	wt, ok := w.(io.WriterAt)
	if !ok {
		var cur int
		cache := make(map[int]outDat)
		var tmpN int
		var dat *outDat
		for cur < len(r.sizes) {
			dat = <-out
			defer pol.Put(dat)
			if dat.err != nil {
				err = dat.err
				return
			}
			if dat.i != cur {
				cache[dat.i] = *dat
				continue
			}
			tmpN, err = w.Write(dat.data)
			n += int64(tmpN)
			if err != nil {
				return
			}
			cur++
			var ok bool
			var curDat outDat
			for {
				curDat, ok = cache[cur]
				if !ok {
					break
				}
				tmpN, err = w.Write(curDat.data)
				n += int64(tmpN)
				if err != nil {
					return
				}
				cur++
			}
		}
	} else {
		var done int
		var dat *outDat
		for done < len(r.sizes) {
			dat = <-out
			defer pol.Put(dat)
			if dat.err != nil {
				err = dat.err
				return
			}
			_, err = wt.WriteAt(dat.data, int64(dat.i*int(r.blockSize)))
			if err != nil {
				return
			}
			done++
		}
	}
	return
}
