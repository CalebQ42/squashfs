package decompress

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type Zstd struct{}

func (z Zstd) Reader(src io.Reader) (io.ReadCloser, error) {
	r, err := zstd.NewReader(src)
	return r.IOReadCloser(), err
}
