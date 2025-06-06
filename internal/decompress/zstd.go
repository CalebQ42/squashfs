package decompress

import (
	"github.com/klauspost/compress/zstd"
)

type Zstd struct {
	rdr *zstd.Decoder
}

func NewZstd() Zstd {
	rdr, _ := zstd.NewReader(nil, zstd.WithDecoderLowmem(true))
	return Zstd{
		rdr: rdr,
	}
}

func (z Zstd) Decompress(data []byte) ([]byte, error) {
	return z.rdr.DecodeAll(data, nil)
}
