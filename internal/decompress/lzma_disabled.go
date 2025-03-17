//go:build no_obsolete

package decompress

import (
	"errors"
)

type Lzma struct{}

func NewLzma() (Lzma, error) {
	return Lzma{}, errors.New("lzma compression is disable in this build with no_obsolete")
}

func (l Lzma) Decompress(data []byte) ([]byte, error) {
	return nil, errors.New("lzma compression is disable in this build with no_obsolete")
}
