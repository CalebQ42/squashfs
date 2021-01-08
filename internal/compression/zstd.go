package compression

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/klauspost/compress/zstd"
)

//Zstd is a zstd compressor/decompressor
type Zstd struct {
	CompressionLevel int32
}

//NewZstdCompressorWithOptions creates a new Zstd with options read from the given reader
func NewZstdCompressorWithOptions(r io.Reader) (*Zstd, error) {
	var zstd Zstd
	err := binary.Read(r, binary.LittleEndian, &zstd)
	if err != nil {
		return nil, err
	}
	return &zstd, nil
}

//Decompress decompresses all data from the reader and returns the uncompressed data
func (z *Zstd) Decompress(r io.Reader) ([]byte, error) {
	rdr, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, rdr)
	return buf.Bytes(), err
}

//Compress impelements compression.Compress
func (z *Zstd) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.EncoderLevel(z.CompressionLevel)))
	if err != nil {
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}
