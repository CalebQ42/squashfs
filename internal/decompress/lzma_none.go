//go:build no_lzma
package decompress

import (
	"fmt"
)

type Lzma struct{}

func (l Lzma) Decompress(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("lzma is not supported")
}
