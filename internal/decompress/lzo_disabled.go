//go:build no_gpl

package decompress

import "errors"

type Lzo struct{}

func NewLzo() (Lzo, error) {
	return Lzo{}, errors.New("lzo compression is disable in this build with no_gpl")
}

func (l Lzo) Decompress(data []byte) ([]byte, error) {
	return nil, errors.New("lzo compression is disable in this build with no_gpl")
}
