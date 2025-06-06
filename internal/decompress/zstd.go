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
	dat, err := z.rdr.DecodeAll(data, nil)
	if err != nil {
		return nil, err
	}
	return dat, err
}
