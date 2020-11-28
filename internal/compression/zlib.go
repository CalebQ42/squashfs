package compression

import (
	"bytes"
	"compress/zlib"
	"io"
)

//Zlib is a decompressor for gzip type compression
type Zlib struct{}

//Decompress reads the entirety of the given reader and returns it uncompressed as a byte slice.
func (z *Zlib) Decompress(r io.Reader) ([]byte, error) {
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

//Compress compresses the given data (as a byte array) and returns the compressed data.
func (z *Zlib) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	wrt := zlib.NewWriter(&buf)
	defer wrt.Close()
	_, err := wrt.Write(data)
	if err != nil {
		return nil, err
	}
	err = wrt.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
