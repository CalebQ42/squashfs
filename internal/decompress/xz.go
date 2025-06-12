package decompress

import (
	"bytes"
	"io"
	"sync"

	"github.com/mikelolasagasti/xz"
)

type Xz struct {
	pool sync.Pool
}

func NewXz() *Xz {
	return &Xz{
		pool: sync.Pool{
			New: func() any {
				rdr, _ := xz.NewReader(nil, 0)
				return rdr
			},
		},
	}
}

func (x *Xz) Decompress(data []byte) ([]byte, error) {
	rdr := x.pool.Get().(*xz.Reader)
	defer x.pool.Put(rdr)
	err := rdr.Reset(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(rdr)
}
