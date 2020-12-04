package compression

import (
	"bytes"
	"io"

	"github.com/ulikunitz/xz/lzma"
)

//Lzma is a lzma decompressor
type Lzma struct{}

//Decompress decompresses all the data in the given reader and returns the uncompressed bytes.
func (l *Lzma) Decompress(rdr io.Reader) ([]byte, error) {
	r, err := lzma.NewReader(rdr)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
