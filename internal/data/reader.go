package data

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	r         io.Reader
	d         decompress.Decompressor
	frag      io.Reader
	sizes     []uint32
	dat       []byte
	curOffset uint16
	curIndex  uint64
}

func NewReader(r io.Reader, d decompress.Decompressor, sizes []uint32) (*Reader, error) {
	return &Reader{
		r:     r,
		d:     d,
		sizes: sizes,
	}, nil
}

func (r *Reader) AddFrag(fragRdr io.Reader) {
	r.frag = fragRdr
}

func (r *Reader) advance() error {
	r.curOffset = 0
	defer func() { r.curIndex++ }()
	var err error
	if r.curIndex == uint64(len(r.sizes))-1 && r.frag != nil {
		r.dat, err = io.ReadAll(r.frag)
		return err
	} else if r.curIndex >= uint64(len(r.sizes))-1 {
		return io.EOF
	}
	realSize := r.sizes[r.curIndex] &^ 0x8000
	r.dat = make([]byte, realSize)
	err = binary.Read(r.r, binary.LittleEndian, &r.dat)
	if err != nil {
		return err
	}
	if r.sizes[r.curIndex] != realSize {
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
		toRead = len(b) - curRead
		if toRead > len(r.dat)-int(r.curOffset) {
			toRead = len(r.dat) - int(r.curOffset)
		}
		copy(b[curRead:], r.dat[r.curOffset:int(r.curOffset)+toRead])
		r.curOffset += uint16(toRead)
		curRead += toRead
	}
	return curRead, nil
}
