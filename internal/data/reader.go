package data

import (
	"bytes"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	master     io.Reader
	cur        io.Reader
	fragRdr    io.Reader
	d          decompress.Decompressor
	blockSizes []uint32
	blockSize  uint32
}

func NewReader(r io.Reader, d decompress.Decompressor, blockSizes []uint32, blockSize uint32) (*Reader, error) {
	var out Reader
	out.d = d
	out.master = r
	out.blockSizes = blockSizes
	out.blockSize = blockSize
	err := out.advance()
	return &out, err
}

func (r *Reader) AddFragment(rdr io.Reader) {
	r.fragRdr = rdr
}

func realSize(siz uint32) uint32 {
	return siz &^ (1 << 24)
}

func (r *Reader) advance() (err error) {
	if clr, ok := r.cur.(io.Closer); ok {
		clr.Close()
	}
	if len(r.blockSizes) == 0 {
		return io.EOF
	}
	if len(r.blockSizes) == 1 && r.fragRdr != nil {
		r.cur = r.fragRdr
	} else {
		size := realSize(r.blockSizes[0])
		if size == 0 {
			r.cur = bytes.NewReader(make([]byte, r.blockSize))
		} else {
			r.cur = io.LimitReader(r.master, int64(size))
			if size == r.blockSizes[0] {
				r.cur, err = r.d.Reader(r.cur)
			}
		}
	}
	r.blockSizes = r.blockSizes[1:]
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.cur.Read(p)
	if err == io.EOF {
		err = r.advance()
		if err != nil {
			return
		}
		var tmpN int
		tmp := make([]byte, len(p)-n)
		tmpN, err = r.Read(tmp)
		for i := range tmp {
			p[n+i] = tmp[i]
		}
		n += tmpN
	}
	return
}
