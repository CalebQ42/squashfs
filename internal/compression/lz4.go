package compression

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pierrec/lz4/v4"
)

//Lz4 is a Lz4 Compressor/Decompressor
type Lz4 struct {
	HC bool
}

//NewLz4CompressorWithOptions creates a new lz4 compressor/decompressor with options read from the given reader.
func NewLz4CompressorWithOptions(r io.Reader) (*Lz4, error) {
	var lz4 Lz4
	var init struct {
		Version int32
		Flags   int32
	}
	err := binary.Read(r, binary.LittleEndian, &init)
	if err != nil {
		return nil, err
	}
	lz4.HC = init.Flags == 1
	return &lz4, nil
}

//Decompress decompresses all data from r and returns the uncompressed bytes
func (l *Lz4) Decompress(r io.Reader) ([]byte, error) {
	rdr := lz4.NewReader(r)
	var buf bytes.Buffer
	_, err := io.Copy(&buf, rdr)
	return buf.Bytes(), err
}
