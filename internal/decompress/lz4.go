package decompress

import (
	"bytes"
	"io"
	"sync"

	"github.com/pierrec/lz4/v4"
)

type Lz4 struct {
	pool sync.Pool
}

func NewLz4() *Lz4 {
	return &Lz4{
		pool: sync.Pool{
			New: func() any {
				return lz4.NewReader(nil)
			},
		},
	}
}

func (l *Lz4) Decompress(data []byte) ([]byte, error) {
	rdr := l.pool.Get().(*lz4.Reader)
	defer l.pool.Put(rdr)
	rdr.Reset(bytes.NewReader(data))
	return io.ReadAll(rdr)
}
