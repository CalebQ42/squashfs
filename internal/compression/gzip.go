package compression

import (
	"bytes"
	"compress/zlib"
	"io"
)

//Gzip is a decompressor for gzip type compression. Uses zlib for compression and decompression
type Gzip struct{}

//Decompress reads the entirety of the given reader and returns it uncompressed as a byte slice.
func (g *Gzip) Decompress(r io.Reader) ([]byte, error) {
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
func (g *Gzip) Compress(data []byte) ([]byte, error) {
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
