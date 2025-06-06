package metadata

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	r         io.Reader
	d         decompress.Decompressor
	dat       []byte
	curOffset uint16
}

func NewReader(r io.Reader, d decompress.Decompressor) Reader {
	return Reader{
		r: r,
		d: d,
	}
}

func (r *Reader) advance() error {
	r.curOffset = 0
	dat := make([]byte, 2)
	_, err := r.r.Read(dat)
	if err != nil {
		return err
	}
	size := binary.LittleEndian.Uint16(dat)
	realSize := size &^ 0x8000
	r.dat = make([]byte, realSize)
	_, err = r.r.Read(r.dat)
	if err != nil {
		return err
	}
	if size != realSize {
		return nil
	}
	r.dat, err = r.d.Decompress(r.dat)
	return err
}

func (r *Reader) Read(b []byte) (int, error) {
	curRead := 0
	var toRead int
	for curRead < len(b) {
		if r.curOffset >= uint16(len(r.dat)) {
			if err := r.advance(); err != nil {
				return curRead, err
			}
		}
		toRead = min(len(b)-curRead, len(r.dat)-int(r.curOffset))
		copy(b[curRead:], r.dat[r.curOffset:int(r.curOffset)+toRead])
		r.curOffset += uint16(toRead)
		curRead += toRead
	}
	return curRead, nil
}

func (r *Reader) Close() error {
	r.dat = nil
	return nil
}
