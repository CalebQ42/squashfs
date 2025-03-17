//go:build no_gpl

package decompress

import "errors"

// The types of compression supported by squashfs
const (
	ZlibCompression = uint16(iota + 1)
	LZMACompression
	LZOCompression
	XZCompression
	LZ4Compression
	ZSTDCompression
)

func GetDecompressor(compType uint16) (Decompressor, error) {
	switch compType {
	case ZlibCompression:
		return Zlib{}, nil
	case LZMACompression:
		return Lzma{}, nil
	case LZOCompression:
		return nil, errors.New("lzo compression is disable in this build with no_gpl")
	case XZCompression:
		return Xz{}, nil
	case LZ4Compression:
		return Lz4{}, nil
	case ZSTDCompression:
		return &Zstd{}, nil
	default:
		return nil, errors.New("invalid compression type. possible corrupted archive")
	}
}
