package decompress

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type Zstd struct {
	writeToReader *zstd.Decoder
}

func (z Zstd) Reader(src io.Reader) (io.ReadCloser, error) {
	r, err := zstd.NewReader(src)
	return r.IOReadCloser(), err
}

func (z Zstd) Reset(old, src io.Reader) error {
	return old.(*zstd.Decoder).Reset(src)
}

func (z Zstd) Decode(in []byte, outSize int) ([]byte, error) {
	if z.writeToReader == nil {
		z.writeToReader, _ = zstd.NewReader(nil)
	}
	return z.writeToReader.DecodeAll(in, make([]byte, outSize))
}
