package decompress

import (
	"sync"

	"github.com/klauspost/compress/zstd"
)

type Zstd struct {
	pool sync.Pool
}

func NewZstd() *Zstd {
	return &Zstd{
		pool: sync.Pool{
			New: func() any {
				rdr, _ := zstd.NewReader(nil, zstd.WithDecoderLowmem(true), zstd.WithDecoderConcurrency(1))
				return rdr
			},
		},
	}
}

func (z *Zstd) Decompress(data []byte) ([]byte, error) {
	rdr := z.pool.Get().(*zstd.Decoder)
	defer func() {
		rdr.Reset(nil)
		z.pool.Put(rdr)
	}()
	return rdr.DecodeAll(data, nil)
}
