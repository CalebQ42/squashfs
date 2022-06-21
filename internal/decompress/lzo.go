package decompress

import (
	"bytes"
	"io"

	"github.com/rasky/go-lzo"
)

type Lzo struct{}

func (l Lzo) Reader(r io.Reader) (io.ReadCloser, error) {
	cache, err := lzo.Decompress1X(r, 0, 0)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(cache)), nil
}

func (l Lzo) Resetable() bool { return false }

func (l Lzo) Reset(old, src io.Reader) error { return ErrNotResetable }
