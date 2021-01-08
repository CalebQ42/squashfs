package compression

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/ulikunitz/xz"
)

type xzInit struct {
	DictionarySize int32
	Filters        int32
}

//Xz is a Xz decompressor.
type Xz struct {
	DictionarySize int32
	HasFilters     bool
}

//NewXzCompressorWithOptions creates a new Xz compressor/decompressor that reads the compressor options from the given reader.
func NewXzCompressorWithOptions(rdr io.Reader) (*Xz, error) {
	var x Xz
	var init xzInit
	err := binary.Read(rdr, binary.LittleEndian, &init)
	if err != nil {
		return nil, err
	}
	x.DictionarySize = init.DictionarySize
	//TODO: When I can do filters, parse the filters
	if init.Filters != 0 {
		x.HasFilters = true
	}
	return &x, nil
}

//Decompress decompresses all the data from the rdr and returns the uncompressed bytes.
func (x *Xz) Decompress(rdr io.Reader) ([]byte, error) {
	r, err := xz.NewReader(rdr)
	if err != nil {
		return nil, err
	}
	r.DictCap = int(x.DictionarySize)
	err = r.Verify()
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
	w, err := xz.NewWriter(&buf)
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
