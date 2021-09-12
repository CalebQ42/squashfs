package compression

import (
	"encoding/binary"
	"io"

	lzo "github.com/rasky/go-lzo"
)

type Lzo struct {
	Algorithm int32
	Level     int32
}

func NewLzoCompressorWithOptions(rdr io.Reader) (*Lzo, error) {
	var lz Lzo
	err := binary.Read(rdr, binary.LittleEndian, &lz)
	if err != nil {
		return nil, err
	}
	return &lz, nil
}

func (l Lzo) Decompress(rdr io.Reader) ([]byte, error) {
	byt, err := lzo.Decompress1X(rdr, 0, 0)
	if err != nil {
		return nil, err
	}
	return byt, nil
}
