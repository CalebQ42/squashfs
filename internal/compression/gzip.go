package compression

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/klauspost/compress/zlib"
)

type gzipInit struct {
	CompressionLevel int32
	WindowSize       int16
	Strategies       int16
}

//Gzip is a decompressor for gzip type compression. Uses zlib for compression and decompression
type Gzip struct {
	wrt *zlib.Writer
	gzipInit
	HasCustomWindow bool
	HasStrategies   bool
}

//NewGzipCompressorWithOptions creates a new gzip compressor/decompressor with options read from the given reader.
func NewGzipCompressorWithOptions(r io.Reader) (*Gzip, error) {
	var gzip Gzip
	err := binary.Read(r, binary.LittleEndian, &gzip.gzipInit)
	if err != nil {
		return nil, err
	}
	//TODO: proper support for window size and strategies
	gzip.HasCustomWindow = gzip.WindowSize != 15
	gzip.HasStrategies = gzip.Strategies != 0 && gzip.Strategies != 1
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
	var err error
	if g.wrt == nil {
		if g.CompressionLevel == 0 {
			g.wrt = zlib.NewWriter(&buf)
		} else {
			g.wrt, err = zlib.NewWriterLevel(&buf, int(g.CompressionLevel))
			if err != nil {
				return nil, err
			}
		}
	}
	wrt, err := zlib.NewWriterLevel(&buf, int(g.CompressionLevel))
	if err != nil {
		return nil, err
	}
	_, err = wrt.Write(data)
	if err != nil {
		return nil, err
	}
	wrt.Close()
	return buf.Bytes(), nil
}
