package decompress

import "io"

type Decompressor interface {
	Reader(src io.Reader) (io.ReadCloser, error)
}
