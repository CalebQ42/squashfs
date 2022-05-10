package decompress

import (
	"io"

	"github.com/klauspost/compress/zlib"
)

type GZip struct{}

func (g GZip) Reader(src io.Reader) (io.ReadCloser, error) {
	return zlib.NewReader(src)
}
