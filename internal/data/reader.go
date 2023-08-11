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
	comRdr     io.Reader
	blockSizes []uint32
	blockSize  uint32
	resetable  bool
	fileSize   uint64
}

func NewReader(r io.Reader, d decompress.Decompressor, blockSizes []uint32, blockSize uint32, fileSize uint64) *Reader {
	return &Reader{
		d:          d,
		master:     r,
		blockSizes: blockSizes,
		blockSize:  blockSize,
		resetable:  true,
		fileSize:   fileSize,
	}
}

func (r *Reader) AddFragment(rdr io.Reader) {
	r.fragRdr = rdr
	r.blockSizes = append(r.blockSizes, 0)
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
			outSize := r.blockSize
			if r.fileSize < uint64(r.blockSize) {
				outSize = uint32(r.fileSize)
			}
			r.cur = bytes.NewReader(make([]byte, outSize))
		} else {
			r.cur = io.LimitReader(r.master, int64(size))
			if size == r.blockSizes[0] {
				if rs, ok := r.d.(decompress.Resetable); ok {
					if r.comRdr == nil {
						r.cur, err = r.d.Reader(r.cur)
						if err != nil {
							return
						}
					} else {
						err = rs.Reset(r.comRdr, r.cur)
						r.cur = r.comRdr
					}
				} else {
					r.cur, err = r.d.Reader(r.cur)
				}
			}
		}
	}
	r.blockSizes = r.blockSizes[1:]
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.cur == nil {
		err = r.advance()
		if err != nil {
			return
		}
	}
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
