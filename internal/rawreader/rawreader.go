package rawreader

import (
	"bytes"
	"errors"
	"io"
)

func ConvertReader(r io.Reader) (RawReader, error) {
	if rr, ok := r.(RawReader); ok {
		return rr, nil
	}
	if rs, is := r.(io.ReadSeeker); is {
		return &fromReadSeeker{
			ReadSeeker: rs,
		}, nil
	}
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	return &fromReader{
		data: buf.Bytes(),
	}, nil
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
	data []byte
	off  int
}

func (r *fromReader) ReadAt(p []byte, off int64) (n int, err error) {
	n = len(p)
	if int(off)+len(p) > len(r.data) {
		n = len(r.data) - int(off)
		err = io.EOF
	}
	if n < 0 {
		n = 0
	}
	for i := 0; i < n; i++ {
		p[i] = r.data[int(off)+i]
	}
	return
}

func (r *fromReader) Seek(off int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekEnd:
		r.off = len(r.data) - int(off)
		if r.off < 0 {
			r.off = 0
			err = io.EOF
		}
	case io.SeekCurrent:
		r.off += int(off)
	case io.SeekStart:
		r.off = int(off)
	}
	if r.off > len(r.data) {
		r.off = len(r.data)
		return int64(r.off), io.EOF
	}
	return int64(r.off), err
}

func (r *fromReader) Read(p []byte) (n int, err error) {
	n = len(p)
	if r.off+len(p) > len(r.data) {
		n = len(r.data) - r.off
		err = io.EOF
	}
	if n < 0 {
		n = 0
	}
	for i := 0; i < n; i++ {
		p[i] = r.data[r.off+i]
	}
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
