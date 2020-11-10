package squashfs

//Superblock is a raw representation of a squashfs
//Descriptions provided by https://dr-emann.github.io/squashfs/
type Superblock struct {
	Magic                uint32 //Magic will be 0x73717368 if it's a legit Squashfs filesystem
	Inodes               uint32 //Inodes is the number of inodes in the inodes table
	MkfsTime             uint32 //MkfsTime is when archive was created
	BlockSize            uint32 //BlockSize is the size of data blocks in bytes
	Fragments            uint32 //Fragments is the number of entries in fragment table
	Compression          uint16 //Compression is what type of compression is used
	BlockLog             uint16 //BlockLog should be log base 2 of BlockSize. If not then the squash might be corrupt
	Flags                uint16 //Flags are the superblock's flags
	IDCount              uint16 //IDCount is the number of IDs in the id lookup table
	Major                uint16 //Major version of squashfs format
	Minor                uint16 //Minor version of squashfs format
	RootInode            uint64 //RootInode is a reference to the root of the squashfs
	BytesUsed            uint64 //BytesUsed is how many bytes the archive is. squashfs archives are often padded to 4KB.
	IDTableOffset        uint64 //IDTableOff is the byte offset of the IDTable
	XattrIDTableOffset   uint64 //XattrIDTableOffset is the byte offset of the xattr id table
	InodeTableOffset     uint64 //InodeTableOffset is the byte offset of the inode table
	DirectoryTableOffset uint64 //DirectoryTableOffset is the byte offset of the directory table
	FragmentTableOffset  uint64 //FragmentTableOffset is the byte offset of the fragment table
	ExportTableOffset    uint64 //ExportTableOffset is the byte offset of the export table
}

//SuperblockFlags is a parsed list of options set in Superblock.Flags
type SuperblockFlags struct {
	UncompressedInodes    bool
	UncompressedData      bool
	Check                 bool //Check is unused in current versions of squashfs
	UncompressedFragments bool
	NoFragments           bool
	AlwaysFragments       bool
	Duplicates            bool //Identical files are stored only once
	Exportable            bool
	UncompressedXattrs    bool
	NoXattrs              bool
	CompressorOptions     bool
	UncompressedIDs       bool
}

//GetFlags returns the Flags parsed into a SuperblockFlags
func (s *Superblock) GetFlags() SuperblockFlags {
	return SuperblockFlags{
		UncompressedInodes:    s.Flags&0x1 == 0x1,
		UncompressedData:      s.Flags&0x2 == 0x2,
		Check:                 s.Flags&0x4 == 0x4,
		UncompressedFragments: s.Flags&0x8 == 0x8,
		NoFragments:           s.Flags&0x10 == 0x10,
		AlwaysFragments:       s.Flags&0x20 == 0x20,
		Duplicates:            s.Flags&0x40 == 0x40,
		Exportable:            s.Flags&0x80 == 0x80,
		UncompressedXattrs:    s.Flags&0x100 == 0x100,
		NoXattrs:              s.Flags&0x200 == 0x200,
		CompressorOptions:     s.Flags&0x400 == 0x400,
		UncompressedIDs:       s.Flags&0x800 == 0x800,
	}
}
