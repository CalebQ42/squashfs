package decompress

import (
	"bytes"
	"io"

	"github.com/klauspost/compress/zstd"
)

type Zstd struct{}

func (z Zstd) Reader(src io.Reader) (io.ReadCloser, error) {
	r, err := zstd.NewReader(src)
	return r.IOReadCloser(), err
}

type ZstdDecodeAll struct {
	rdr *zstd.Decoder
}

func (z *ZstdDecodeAll) Reader(src io.Reader) (io.ReadCloser, error) {
	if z.rdr == nil {
		z.rdr, _ = zstd.NewReader(nil)
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	out, err := z.rdr.DecodeAll(data, nil)
	return io.NopCloser(bytes.NewReader(out)), err
}
