package squashfs

import (
	"bytes"
	"compress/zlib"
	"io"
)

//Decompressor is a squashfs decompressor interface. Allows for easy decompression no matter the type of compression.
type decompressor interface {
	Decompress(io.Reader) ([]byte, error)
}

//ZlibDecompressor is a decompressor for gzip type compression
type zlibDecompressor struct{}

//Decompress reads the entirety of the given reader and returns it uncompressed as a byte slice.
func (z *zlibDecompressor) Decompress(r io.Reader) ([]byte, error) {
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
