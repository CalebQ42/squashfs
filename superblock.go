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

//SuperblockFlags is the series of flags describing how a squashfs archive is packed.
type SuperblockFlags struct {
	//If true, inodes are stored uncompressed.
	UncompressedInodes bool
	//If true, data is stored uncompressed.
	UncompressedData bool
	check            bool
	//If true, fragments are stored uncompressed.
	UncompressedFragments bool
	//If true, ALL data is stored in sequential data blocks instead of utilizing fragments.
	NoFragments bool
	//If true, the last block of data will always be stored as a fragment if it's less then the block size.
	AlwaysFragment bool
	//If true, duplicate files are only stored once. (Currently unsupported)
	RemoveDuplicates bool
	//If true, the export table is populated. (Currently unsupported)
	Exportable bool
	//If true, the xattr table is uncompressed. (Currently unsupported)
	UncompressedXattr bool
	//If true, the xattr table is not populated. (Currently unsupported)
	NoXattr           bool
	compressorOptions bool
	//If true, the UID/GID table is stored uncompressed.
	UncompressedIDs bool
}

//DefaultFlags are the default SuperblockFlags that are used.
var DefaultFlags = SuperblockFlags{
	RemoveDuplicates: true,
	Exportable:       true,
}

//GetFlags returns a SuperblockFlags for a given superblock.
func (s *superblock) GetFlags() SuperblockFlags {
	return SuperblockFlags{
		UncompressedInodes:    s.Flags&0x1 == 0x1,
		UncompressedData:      s.Flags&0x2 == 0x2,
		check:                 s.Flags&0x4 == 0x4,
		UncompressedFragments: s.Flags&0x8 == 0x8,
		NoFragments:           s.Flags&0x10 == 0x10,
		AlwaysFragment:        s.Flags&0x20 == 0x20,
		RemoveDuplicates:      s.Flags&0x40 == 0x40,
		Exportable:            s.Flags&0x80 == 0x80,
		UncompressedXattr:     s.Flags&0x100 == 0x100,
		NoXattr:               s.Flags&0x200 == 0x200,
		compressorOptions:     s.Flags&0x400 == 0x400,
		UncompressedIDs:       s.Flags&0x800 == 0x800,
	}
}

//ToUint returns the uint16 representation of the given SuperblockFlags
func (s *SuperblockFlags) ToUint() uint16 {
	var out uint16
	if s.UncompressedInodes {
		out = out | 0x1
	}
	if s.UncompressedData {
		out = out | 0x2
	}
	if s.check {
		out = out | 0x4
	}
	if s.UncompressedFragments {
		out = out | 0x8
	}
	if s.NoFragments {
		out = out | 0x10
	}
	if s.AlwaysFragment {
		out = out | 0x20
	}
	if s.RemoveDuplicates {
		out = out | 0x40
	}
	if s.Exportable {
		out = out | 0x80
	}
	if s.UncompressedXattr {
		out = out | 0x100
	}
	if s.NoXattr {
		out = out | 0x200
	}
	if s.compressorOptions {
		out = out | 0x400
	}
	if s.UncompressedIDs {
		out = out | 0x800
	}
	return out
}
