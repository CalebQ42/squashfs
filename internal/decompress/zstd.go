package decompress

import (
	"github.com/klauspost/compress/zstd"
)

type Zstd struct{}

func (z Zstd) Decompress(data []byte) ([]byte, error) {
	rdr, err := zstd.NewReader(nil, zstd.WithDecoderLowmem(true), zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	defer rdr.Close()
	return rdr.DecodeAll(data, nil)
}
