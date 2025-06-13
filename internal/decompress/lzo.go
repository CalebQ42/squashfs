package decompress

import (
	"github.com/anchore/go-lzo"
)

type Lzo struct{}

func NewLzo() (Lzo, error) {
	return Lzo{}, nil
}

func (l Lzo) Decompress(data []byte) ([]byte, error) {
	var dest []byte
	_, err := lzo.Decompress(data, dest)
	return dest, err
}
