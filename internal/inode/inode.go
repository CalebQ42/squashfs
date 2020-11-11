package inode

type CommonHeader struct {
	InodeType    uint16
	Permissions  uint16
	UID          uint16
	GID          uint16
	ModifiedTime uint32
	Number       uint32
}

type BasicDirectory struct {
	DirectoryIndex    uint32
	HardLinks         uint32
	DirectorySize     uint16
	DirectoryOffset   uint16
	ParentInodeNumber uint32
}

type ExtendedDirectoryInit struct {
	HardLinks         uint32
	DirectorySize     uint32
	DirectoryIndex    uint32
	ParentInodeNumber uint32
	IndexCount        uint16 //one less then directory indexes following structure
	DirectoryOffset   uint16
	XattrIndex        uint32
}

type ExtendedDirectory struct {
	ExtendedDirectoryInit
	//TODO indexes []DirectoryIndex
}

type BasicFile struct {
	BlockStart     uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	Size           uint32
	BlockSizes     []uint32
}

type ExtendedFile struct {
	BlockStart     uint32
	Size           uint32
	Sparse         uint64
	HardLinks      uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	XattrIndex     uint32
	BlockSizes     []uint32
}

type BasicSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

type BasicSymlink struct {
	BasicSymlinkInit
	targetPath []byte //len is TargetPathSize
}

type ExtendedSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

type ExtendedSymlink struct {
	targetPath []byte
	XattrIndex uint32
}

type BasicDevice struct {
	HardLinks uint32
	Device    uint32
}

type ExtendedDevice struct {
	BasicDevice
	XattrIndex uint32
}

type BasicIPC struct {
	HardLink uint32
}

type ExtendedIPC struct {
	BasicIPC
	XattrIndex uint32
}
