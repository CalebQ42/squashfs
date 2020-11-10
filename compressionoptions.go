package squashfs

//CompressionOptions
type CompressionOptions interface {
	Decompress([]byte) []byte
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

type xzOptionsRaw struct {
	dictionarySize    int32
	executableFilters int32
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
