package decompress

import (
	"bytes"
	"io"
	"sync"

	"github.com/klauspost/compress/zlib"
)

type Zlib struct {
	pool sync.Pool
}

func NewZlib() *Zlib {
	return &Zlib{}
}

func (z *Zlib) Decompress(data []byte) ([]byte, error) {
	rdr := z.pool.Get()
	defer z.pool.Put(rdr)
	var err error
	if rdr == nil {
		rdr, err = zlib.NewReader(bytes.NewReader(data))
	} else {
		err = rdr.(zlib.Resetter).Reset(bytes.NewReader(data), nil)
	}
	if err != nil {
		return nil, err
	}
	return io.ReadAll(rdr.(io.ReadCloser))
}
