package squashfs

import (
	"compress/zlib"
	"io"
)

const (
	zlibCompression = 1 + iota
	lzmaCompression
	lzoCompression
	xzCompression
	lz4Compression
	zstdCompression
)

//TODO: implement decompress for each type of Options
type CompressionOptions interface {
	Decompress(*io.SectionReader, int) ([]byte, error)
	DecompressCopy(*io.Reader, *io.Writer) (int, error)
	Compress(*io.SectionReader, int) ([]byte, error)
	CompressCopy(*io.Reader, *io.Writer) (int, error)
}

//TODO: Allow creation of options for compression.

type gzipOptionsRaw struct {
	compressionLevel int32
	windowSize       int16
	strategies       int16
}

//GzipOptions is the options used for gzip compression. Backed by the raw format, with strategies parsed.
type GzipOptions struct {
	CompressionOptions       //TODO: remove
	raw                      *gzipOptionsRaw
	DefaultStrategy          bool
	FilteredStrategy         bool
	HuffmanOnlyStrategy      bool
	RunLengthEncodedStrategy bool
	FixedStretegy            bool
}

func NewGzipOptions(raw gzipOptionsRaw) GzipOptions {
	//TODO: parse strategies
	return GzipOptions{
		raw: &raw,
	}
}

func (gzipOp *GzipOptions) Decompress(rdr *io.SectionReader, blockSize int) ([]byte, error) {
	zlibRdr, err := zlib.NewReader(rdr)
	defer zlibRdr.Close()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0)
	var tmp []byte
	read := blockSize
	for read == blockSize {
		tmp = make([]byte, blockSize)
		read, err = zlibRdr.Read(tmp)
		if err != io.EOF {
			return nil, err
		}
		if read < blockSize {
			tmp = tmp[:read]
		}
		out = append(out, tmp...)
	}
	return out, nil
}

func (gzipOp *GzipOptions) DecompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	zlibRdr, err := zlib.NewReader(*rdr)
	defer zlibRdr.Close()
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(*wrt, zlibRdr)
	return int(n), err
}

func (gzipOp *GzipOptions) Compress(rdr *io.SectionReader, blockSize int) ([]byte, error) {

}

func (gzipOp *GzipOptions) CompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	zlibWrt, err := zlib.NewWriter(*wrt) //TODO: allow setting level
	defer zlibWrt.Close()
	if err != nil {
		return 0, err
	}

}

type xzOptionsRaw struct {
	dictionarySize    int32
	executableFilters int32
}

type XzOptions struct {
	CompressionOptions //TODO: Remove
	raw                *xzOptionsRaw
	Execx86            bool
	ExecPower          bool
	Execa64            bool
	ExecArm            bool
	ExecArmThumb       bool
	ExecSparc          bool
}

func NewXzOption(raw xzOptionsRaw) XzOptions {
	return XzOptions{
		raw:          &raw,
		Execx86:      raw.executableFilters&0x1 == 0x1,
		ExecPower:    raw.executableFilters&0x2 == 0x2,
		Execa64:      raw.executableFilters&0x4 == 0x4,
		ExecArm:      raw.executableFilters&0x8 == 0x8,
		ExecArmThumb: raw.executableFilters&0x10 == 0x10,
		ExecSparc:    raw.executableFilters&0x20 == 0x20,
	}
}

type lz4OptionsRaw struct {
	version int32
	flags   int32
}

//ZstdOptions is the options set for zstdOptions
type ZstdOptions struct {
	CompressionLevel int32 //CompressionLevel should be between 1 and 22
}

type lzoOptionsRaw struct {
	algorithm        int32
	compressionLevel int32
}
