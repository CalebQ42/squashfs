//go:build !no_gpl

package decompress

import (
	"bytes"

	"github.com/rasky/go-lzo"
)

type Lzo struct{}

func NewLzo() (Lzo, error) {
	return Lzo{}, nil
}

func (l Lzo) Decompress(data []byte) ([]byte, error) {
	return lzo.Decompress1X(bytes.NewReader(data), len(data), 0)
}
