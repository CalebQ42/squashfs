//go:build no_lzo
package decompress

import (
	"fmt"
)

type Lzo struct{}

func (l Lzo) Decompress(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("lzo is not supported")
}
