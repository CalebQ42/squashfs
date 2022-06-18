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
}

func NewReader(r io.Reader, d decompress.Decompressor) (*Reader, error) {
	var out Reader
	out.d = d
	out.master = r
	return &out, out.Advance()
}

func (r *Reader) Advance() error {

	//For some reason things get closed improperly and causes issues.
	//NO IDEA HOW THIS IS HAPPENING.

	// if clr, ok := r.cur.(io.Closer); ok {
	// 	clr.Close()
	// 	r.cur = nil
	// }
	var size uint16
	err := binary.Read(r.master, binary.LittleEndian, &size)
	if err != nil {
		return err
	}
	comp := size&0x8000 != 0x8000
	size &^= 0x8000
	r.cur = io.LimitReader(r.master, int64(size))
	if comp {
		r.cur, err = r.d.Reader(r.cur)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Reader) Read(p []byte) (n int, err error) {
	n, err = r.cur.Read(p)
	if err == io.EOF {
		err = r.Advance()
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
