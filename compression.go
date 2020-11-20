package squashfs

import (
	"bytes"
	"compress/zlib"
	"io"
)

type Decompressor interface {
	Decompress(io.Reader) ([]byte, error)
}

type ZlibDecompressor struct{}

func (z *ZlibDecompressor) Decompress(r io.Reader) ([]byte, error) {
	rdr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	var data bytes.Buffer
	_, err = io.Copy(&data, rdr)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}
