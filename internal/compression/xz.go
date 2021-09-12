package compression

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/therootcompany/xz"

	wrtXz "github.com/ulikunitz/xz"
)

type Xz struct {
	DictionarySize int32
	Filters        int32
}

//NewXzCompressorWithOptions creates a new Xz compressor/decompressor that reads the compressor options from the given reader.
func NewXzCompressorWithOptions(rdr io.Reader) (*Xz, error) {
	var x Xz
	err := binary.Read(rdr, binary.LittleEndian, &x)
	if err != nil {
		return nil, err
	}
	return &x, nil
}

//Decompress decompresses all the data from the rdr and returns the uncompressed bytes.
func (x *Xz) Decompress(rdr io.Reader) ([]byte, error) {
	r, err := xz.NewReader(rdr, 0)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//Compress implements compression.Compress
func (x *Xz) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := wrtXz.NewWriter(&buf)
	if err != nil {
		return nil, err
	}
	w.DictCap = int(x.DictionarySize)
	err = w.Verify()
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
