package metadata

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	master io.Reader
	cur    io.Reader
	d      decompress.Decompressor
	comRdr io.Reader
}

func NewReader(master io.Reader, d decompress.Decompressor) *Reader {
	return &Reader{
		master: master,
		d:      d,
	}
}

func realSize(siz uint16) uint16 {
	return siz &^ 0x8000
}

func (r *Reader) advance() (err error) {
	if _, ok := r.d.(decompress.Resetable); !ok {
		if clr, ok := r.cur.(io.Closer); ok {
			clr.Close()
		}
	}
	var raw uint16
	err = binary.Read(r.master, binary.LittleEndian, &raw)
	if err != nil {
		return
	}
	size := realSize(raw)
	r.cur = io.LimitReader(r.master, int64(size))
	if size == raw {
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
		for i := 0; i < tmpN; i++ {
			p[n+i] = tmp[i]
		}
		n += tmpN
	}
	return
}
