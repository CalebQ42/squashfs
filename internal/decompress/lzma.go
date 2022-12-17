package decompress

import (
	"io"

	"github.com/ulikunitz/xz/lzma"
)

type Lzma struct{}

func (l Lzma) Reader(r io.Reader) (io.ReadCloser, error) {
	rdr, err := lzma.NewReader(r)
	return io.NopCloser(rdr), err
}
