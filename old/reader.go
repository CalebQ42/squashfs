package squashfs

import (
	"errors"
	"io"
)

//TODO: possible custom reader because I'm havng some issuse...

//Reader is a reader which implements Reader, ReaderAt, and Seeker, all with an accesible offset (for reasons)
type Reader struct {
	rdr    io.ReaderAt
	offset int64
}

//NewReader creates a squashfs.Reader from a io.ReaderAt
func NewReader(baseReader io.ReaderAt) Reader {
	return Reader{
		rdr:    baseReader,
		offset: 0,
	}
}

//Read reads len(byt) into byt. Advances the internal offset
func (r *Reader) Read(byt []byte) (int, error) {
	n, err := r.rdr.ReadAt(byt, r.offset)
	r.offset += int64(n)
	return n, err
}

//ReadAt wraps the internal io.ReadAt's function. DOES NOT advance the internal offset for Read function.
//Returns how many bytes were read.
func (r *Reader) ReadAt(byt []byte, offset int64) (int, error) {
	return r.rdr.ReadAt(byt, offset)
}

//ReadAtFromOffset is the same as ReadAt, but the given offset is offset by the internal offset. DOES NOT advance the internal offset.
//Returns how many bytes were read.
func (r *Reader) ReadAtFromOffset(byt []byte, offset int64) (int, error) {
	offset += r.offset
	return r.rdr.ReadAt(byt, offset)
}

//Seek advances the internal offset. SeekEnd DOES NOT work
//Might not be necessary, but here just in case
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		n, err := r.Read(make([]byte, offset))
		return int64(n), err
	case io.SeekStart:
		r.offset = 0
		n, err := r.Read(make([]byte, offset))
		return int64(n), err
	case io.SeekEnd:
		return 0, errors.New("SeekEnd is NOT supported")
	}
	return 0, errors.New("incorrect whence")
}
