package rawreader

import (
	"errors"
	"io"
)

func ConvertReader(r io.Reader) RawReader {
	if rr, ok := r.(RawReader); ok {
		return rr
	}
	if rs, is := r.(io.ReadSeeker); is {
		return &fromReadSeeker{
			ReadSeeker: rs,
		}
	}
	return &fromReader{
		rdr:   r,
		cache: make([]byte, 0),
	}
}

func ConvertReaderAt(r io.ReaderAt) RawReader {
	if rr, ok := r.(RawReader); ok {
		return rr
	}
	return &fromReaderAt{
		ReaderAt: r,
	}
}

//TODO: Add way to discard data from fromReader
//RawReader implements the needed interfaces for reading a squashfs archive.
type RawReader interface {
	io.ReadSeeker
	io.ReaderAt
}

type fromReader struct {
	rdr   io.Reader
	cache []byte
	off   int
}

func (r *fromReader) increaseCache(len int) error {
	newCache := make([]byte, len)
	_, err := r.rdr.Read(newCache)
	if err != nil {
		return err
	}
	r.cache = append(r.cache, newCache...)
	return nil
}

func (r *fromReader) ReadAt(p []byte, off int64) (n int, err error) {
	if int(off)+len(p) > len(r.cache) {
		r.increaseCache((int(off) + len(p)) - len(r.cache))
	}
	for i := int64(0); i < int64(len(p)); i++ {
		p[i] = r.cache[off+i]
	}
	return
}

func (r *fromReader) Seek(off int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekEnd:
		return 0, errors.New("cannot SeekEnd RawReader")
	case io.SeekCurrent:
		r.off += int(off)
	case io.SeekStart:
		r.off = int(off)
	}
	if r.off > len(r.cache) {
		err = r.increaseCache(len(r.cache) - r.off)
		if err != nil {
			r.off = len(r.cache)
		}
	}
	return int64(r.off), err
}

func (r *fromReader) Read(p []byte) (n int, err error) {
	if len(p)+r.off > len(r.cache) {
		err = r.increaseCache((len(p) + r.off) - len(r.cache))
		if err != nil {
			return
		}
	}
	for i := 0; i < len(p); i++ {
		p[i] = r.cache[r.off+i]
	}
	r.off += len(p)
	return
}

type fromReadSeeker struct {
	io.ReadSeeker
}

func (r *fromReadSeeker) ReadAt(p []byte, off int64) (n int, err error) {
	tmp, _ := r.Seek(0, io.SeekCurrent)
	defer r.Seek(tmp, io.SeekStart)
	_, err = r.Seek(off, io.SeekStart)
	if err != nil {
		return
	}
	return r.Read(p)
}

type fromReaderAt struct {
	io.ReaderAt

	off int
}

func (r *fromReaderAt) Read(p []byte) (n int, err error) {
	n, err = r.ReadAt(p, int64(r.off))
	r.off += n
	return
}

func (r *fromReaderAt) Seek(off int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekEnd:
		return 0, errors.New("cannot SeekEnd RawReader")
	case io.SeekCurrent:
		r.off += int(off)
	case io.SeekStart:
		r.off = int(off)
	}
	return int64(r.off), nil
}
