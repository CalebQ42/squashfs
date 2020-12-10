package compression

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

type gzipInit struct {
	CompressionLevel int32
	WindowSize       int16
	Strategies       int16
}

//Gzip is a decompressor for gzip type compression. Uses zlib for compression and decompression
type Gzip struct {
	CompressionLevel int32
	HasCustomWindow  bool
	HasStrategies    bool
}

//NewGzipCompressorWithOptions creates a new gzip compressor/decompressor with options read from the given reader.
func NewGzipCompressorWithOptions(r io.Reader) (*Gzip, error) {
	var gzip Gzip
	var init gzipInit
	err := binary.Read(r, binary.LittleEndian, &init)
	if err != nil {
		return nil, err
	}
	gzip.CompressionLevel = init.CompressionLevel
	//TODO: proper support for window size and strategies
	gzip.HasCustomWindow = init.WindowSize != 15
	gzip.HasStrategies = init.Strategies != 0 && init.Strategies != 1
	return &gzip, nil
}

//Decompress reads the entirety of the given reader and returns it uncompressed as a byte slice.
func (g *Gzip) Decompress(r io.Reader) ([]byte, error) {
	rdr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	var data bytes.Buffer
	_, err = io.Copy(&data, rdr)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

//Compress compresses the given data (as a byte array) and returns the compressed data.
func (g *Gzip) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	wrt := zlib.NewWriter(&buf)
	defer wrt.Close()
	_, err := wrt.Write(data)
	if err != nil {
		return nil, err
	}
	err = wrt.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
