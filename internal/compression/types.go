package compression

import "io"

//Compressor is a squashfs decompressor interface. Allows for easy compression.
type Compressor interface {
	Compress([]byte) ([]byte, error)
}

//Decompressor is a squashfs decompressor interface. Allows for easy decompression no matter the type of compression.
type Decompressor interface {
	Decompress(io.Reader) ([]byte, error)
}
