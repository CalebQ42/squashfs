package compression

import (
	"bytes"
	"io"

	"github.com/ulikunitz/xz"
)

//Xz is a decompressor for xz type compression
type Xz struct{}

//Decompress reads the entirety of the given reader and returns it uncompressed as a byte slice.
func (z *Xz) Decompress(r io.Reader) ([]byte, error) {
	rdr, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	err = rdr.Verify()
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
func (z *Xz) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	wrt, err := xz.NewWriter(&buf)
	if err != nil {
		return nil, err
	}
	defer wrt.Close()
	_, err = wrt.Write(data)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
