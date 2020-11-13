package squashfs

import (
	"compress/gzip"
	"io"

	"gopkg.in/src-d/go-git.v4/utils/ioutil"
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
	Decompress(*io.SectionReader, int) ([]byte, error)
	DecompressCopy(*io.Reader, *io.Writer) (int, error)
	Compress(*io.SectionReader, int) ([]byte, error)
	CompressCopy(*io.Reader, *io.Writer) (int, error)
	Reader(io.Reader) (*io.ReadCloser, error)
}

//TODO: Allow creation of options for compression.

type gzipOptionsRaw struct {
	compressionLevel int32
	windowSize       int16
	strategies       int16
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

type byteWriterReader struct {
	byts   []byte
	offset int
}

func newByteReadWrite(length int) *byteWriterReader {
	return &byteWriterReader{
		byts:   make([]byte, length),
		offset: 0,
	}
}

func newByteReadWriteFromBytes(byts []byte) *byteWriterReader {
	return &byteWriterReader{
		byts:   byts,
		offset: 0,
	}
}

func (bwr *byteWriterReader) getBytes() []byte {
	return bwr.byts
}

//Read reads the bytes.
func (bwr *byteWriterReader) Read(byt []byte) (int, error) {
	if len(bwr.byts) < bwr.offset+len(byt) {
		bytesWritten := len(bwr.byts) - bwr.offset
		for i := 0; i < bytesWritten; i++ {
			byt[i] = bwr.byts[i+bwr.offset]
		}
		return bytesWritten, io.EOF
	}
	for i := 0; i < len(byt); i++ {
		byt[i] = bwr.byts[bwr.offset+i]
	}
	bwr.offset += len(byt)
	return len(byt), nil
}

//Write writes to the bytes. WILL expand to accept the incoming bytes.
func (bwr *byteWriterReader) Write(byts []byte) (int, error) {
	bwr.byts = append(bwr.byts, byts...)
	return len(byts), nil
}

func NewGzipOptions(raw gzipOptionsRaw) *GzipOptions {
	//TODO: parse strategies
	return &GzipOptions{
		raw: &raw,
	}
}

func (gzipOp *GzipOptions) Decompress(rdr *io.SectionReader, blockSize int) ([]byte, error) {
	gzipRdr, err := gzip.NewReader(rdr)
	defer gzipRdr.Close()
	if err != nil {
		return nil, err
	}
	bytrw := newByteReadWrite(0)
	_, err = io.Copy(bytrw, gzipRdr)
	if err != nil {
		return bytrw.byts, err
	}
	return bytrw.byts, nil
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

func (gzipOp *GzipOptions) Compress(rdr *io.SectionReader, blockSize int) ([]byte, error) {
	bytWrt := newByteReadWrite(0)
	gzipWrt := gzip.NewWriter(bytWrt) //TODO: allow setting level
	defer gzipWrt.Close()
	_, err := io.Copy(gzipWrt, rdr)
	if err != nil {
		return bytWrt.byts, err
	}
	return bytWrt.byts, nil
}

func (gzipOp *GzipOptions) CompressCopy(rdr *io.Reader, wrt *io.Writer) (int, error) {
	gzipWrt := gzip.NewWriter(*wrt) //TODO: allow setting level
	defer gzipWrt.Close()
	n, err := io.Copy(gzipWrt, *rdr)
	return int(n), err
}

func (gzipOp *GzipOptions) Reader(rdr io.Reader) (*io.ReadCloser, error) {
	read, err := gzip.NewReader(rdr)
	redClo := ioutil.NewReadCloser(read, read)
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
