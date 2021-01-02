package squashfs

//The types of compression supported by squashfs.
const (
	GzipCompression = 1 + iota
	LzmaCompression
	LzoCompression
	XzCompression
	Lz4Compression
	ZstdCompression
)

//Superblock contains important information about a squashfs file. Located at the very front of the archive.
type superblock struct {
	Magic            uint32
	InodeCount       uint32
	CreationTime     uint32
	BlockSize        uint32
	FragCount        uint32
	CompressionType  uint16
	BlockLog         uint16
	Flags            uint16
	IDCount          uint16
	MajorVersion     uint16
	MinorVersion     uint16
	RootInodeRef     uint64
	BytesUsed        uint64
	IDTableStart     uint64
	XattrTableStart  uint64
	InodeTableStart  uint64
	DirTableStart    uint64
	FragTableStart   uint64
	ExportTableStart uint64
}

//SuperblockFlags is the parsed version of Superblock.Flags
type superblockFlags struct {
	UncompressedInodes    bool
	UncompressedData      bool
	Check                 bool
	UncompressedFragments bool
	NoFragments           bool
	AlwaysFragments       bool
	Duplicates            bool
	Exportable            bool
	UncompressedXattr     bool
	NoXattr               bool
	CompressorOptions     bool
	UncompressedIDs       bool
}

//GetFlags returns a SuperblockFlags for a given superblock.
func (s *superblock) GetFlags() superblockFlags {
	return superblockFlags{
		UncompressedInodes:    s.Flags&0x1 == 0x1,
		UncompressedData:      s.Flags&0x2 == 0x2,
		Check:                 s.Flags&0x4 == 0x4,
		UncompressedFragments: s.Flags&0x8 == 0x8,
		NoFragments:           s.Flags&0x10 == 0x10,
		AlwaysFragments:       s.Flags&0x20 == 0x20,
		Duplicates:            s.Flags&0x40 == 0x40,
		Exportable:            s.Flags&0x80 == 0x80,
		UncompressedXattr:     s.Flags&0x100 == 0x100,
		NoXattr:               s.Flags&0x200 == 0x200,
		CompressorOptions:     s.Flags&0x400 == 0x400,
		UncompressedIDs:       s.Flags&0x800 == 0x800,
	}
}
