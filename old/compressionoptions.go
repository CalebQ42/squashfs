package squashfs

import (
	"compress/zlib"
	"io"

	"gopkg.in/src-d/go-git.v4/utils/ioutil"
)

const (
	zlibCompression = 1 + iota
	lzmaCompression
	lzoCompression
	xzCompression
	lz4Compression
	zstdCompression
)

//TODO: implement for each type of Options
type CompressionOptions interface {
	Decompress(*io.SectionReader) ([]byte, error)
	DecompressCopy(*io.Reader, *io.Writer) (int, error)
	Compress(*io.SectionReader) ([]byte, error)
	CompressCopy(*io.Reader, *io.Writer) (int, error)
	Reader(io.Reader) (*io.ReadCloser, error)
}

//TODO: Allow creation of options for compression.

type zlibOptionsRaw struct {
	compressionLevel int32
	windowSize       int16
	strategies       int16
}

//ZlibOptions is the options used for zlib compression. Backed by the raw format, with strategies parsed.
type ZlibOptions struct {
	CompressionOptions
	raw                      *zlibOptionsRaw
	DefaultStrategy          bool
	FilteredStrategy         bool
	HuffmanOnlyStrategy      bool
	RunLengthEncodedStrategy bool
	FixedStretegy            bool
}

func NewZlibOptions(raw zlibOptionsRaw) *ZlibOptions {
	//TODO: parse strategies
	return &ZlibOptions{
		raw: &raw,
	}
}

func (zlibOp *ZlibOptions) Decompress(rdr *io.SectionReader) ([]byte, error) {
	zlibRdr, err := zlib.NewReader(rdr)
	defer zlibRdr.Close()
	if err != nil {
		return nil, err
	}
	bytrw := newByteReadWrite(0)
	_, err = io.Copy(bytrw, zlibRdr)
	if err != nil {
		return bytrw.byts, err
	}
	return bytrw.byts, nil
}

func (zlibOp *ZlibOptions) DecompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	zlibRdr, err := zlib.NewReader(*rdr)
	defer zlibRdr.Close()
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(*wrt, zlibRdr)
	return int(n), err
}

func (zlibOp *ZlibOptions) Compress(rdr *io.SectionReader) ([]byte, error) {
	bytWrt := newByteReadWrite(0)
	zlibWrt := zlib.NewWriter(bytWrt) //TODO: allow setting level
	defer zlibWrt.Close()
	_, err := io.Copy(zlibWrt, rdr)
	if err != nil {
		return bytWrt.byts, err
	}
	return bytWrt.byts, nil
}

func (zlibOp *ZlibOptions) CompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	zlibWrt := zlib.NewWriter(*wrt) //TODO: allow setting level
	defer zlibWrt.Close()
	n, err := io.Copy(zlibWrt, *rdr)
	return int(n), err
}

func (zlibOp *ZlibOptions) Reader(rdr io.Reader) (*io.ReadCloser, error) {
	read, err := zlib.NewReader(rdr)
	redClo := ioutil.NewReadCloser(NewByteBufferedReader(read), read)
	return &redClo, err
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
