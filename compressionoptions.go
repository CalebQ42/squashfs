package squashfs

import (
	"compress/gzip"
	"io"

	"github.com/CalebQ42/GoSquashfs/bytereadwrite"
)

const (
	gzipCompression = 1 + iota
	lzmaCompression
	lzoCompression
	xzCompression
	lz4Compression
	zstdCompression
)

//TODO: implement for each type of Options
type CompressionOptions interface {
	Decompress(io.Reader) ([]byte, error)
	DecompressCopy(*io.Reader, *io.Writer) (int, error)
	Compress(*io.Reader) ([]byte, error)
	CompressCopy(*io.Reader, *io.Writer) (int, error)
}

//TODO: Allow creation of options for compression.

type gzipOptionsRaw struct {
	CompressionLevel int32
	WindowSize       int16
	Strategies       int16
}

//GzipOptions is the options used for gzip compression. Backed by the raw format, with strategies parsed.
type GzipOptions struct {
	CompressionOptions
	raw                      *gzipOptionsRaw
	DefaultStrategy          bool
	FilteredStrategy         bool
	HuffmanOnlyStrategy      bool
	RunLengthEncodedStrategy bool
	FixedStretegy            bool
}

func NewGzipOptions(raw gzipOptionsRaw) *GzipOptions {
	//TODO: parse strategies
	return &GzipOptions{
		raw: &raw,
	}
}

func (gzipOp *GzipOptions) Decompress(rdr io.Reader) ([]byte, error) {
	gzipRdr, err := gzip.NewReader(rdr)
	// defer gzipRdr.Close()
	if err != nil {
		return nil, err
	}
	bytrw := bytereadwrite.NewByteReaderWriter()
	_, err = io.Copy(bytrw, gzipRdr)
	if err != nil {
		return nil, err
	}
	return bytrw.GetBytes(), nil
}

func (gzipOp *GzipOptions) DecompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	gzipRdr, err := gzip.NewReader(*rdr)
	defer gzipRdr.Close()
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(*wrt, gzipRdr)
	return int(n), err
}

func (gzipOp *GzipOptions) Compress(rdr *io.Reader) ([]byte, error) {
	bytWrt := bytereadwrite.NewByteReaderWriter()
	gzipWrt := gzip.NewWriter(bytWrt) //TODO: allow setting level
	defer gzipWrt.Close()
	_, err := io.Copy(gzipWrt, *rdr)
	if err != nil {
		return bytWrt.GetBytes(), err
	}
	return bytWrt.GetBytes(), nil
}

func (gzipOp *GzipOptions) CompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	gzipWrt := gzip.NewWriter(*wrt) //TODO: allow setting level
	defer gzipWrt.Close()
	n, err := io.Copy(gzipWrt, *rdr)
	return int(n), err
}

type xzOptionsRaw struct {
	DictionarySize    int32
	ExecutableFilters int32
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
		Execx86:      raw.ExecutableFilters&0x1 == 0x1,
		ExecPower:    raw.ExecutableFilters&0x2 == 0x2,
		Execa64:      raw.ExecutableFilters&0x4 == 0x4,
		ExecArm:      raw.ExecutableFilters&0x8 == 0x8,
		ExecArmThumb: raw.ExecutableFilters&0x10 == 0x10,
		ExecSparc:    raw.ExecutableFilters&0x20 == 0x20,
	}
}

type lz4OptionsRaw struct {
	Version int32
	Flags   int32
}

//ZstdOptions is the options set for zstdOptions
type ZstdOptions struct {
	CompressionLevel int32 //CompressionLevel should be between 1 and 22
}

type lzoOptionsRaw struct {
	Algorithm        int32
	CompressionLevel int32
}
