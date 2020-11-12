package inode

import (
	"encoding/binary"
	"io"
)

//InodeCommon is the comon header for all inodes
type InodeCommon struct {
	InodeType    uint16
	Permissions  uint16
	UID          uint16
	GID          uint16
	ModifiedTime uint32
	Number       uint32
}

//BasicDirectory is self explainatory
type BasicDirectory struct {
	DirectoryIndex    uint32
	HardLinks         uint32
	DirectorySize     uint16
	DirectoryOffset   uint16
	ParentInodeNumber uint32
}

//ExtendedDirectoryInit is the information that can be directoy decoded
type ExtendedDirectoryInit struct {
	HardLinks         uint32
	DirectorySize     uint32
	DirectoryIndex    uint32
	ParentInodeNumber uint32
	IndexCount        uint16 //one less then directory indexes following structure
	DirectoryOffset   uint16
	XattrIndex        uint32
}

//ExtendedDirectory is a directory with extra info
type ExtendedDirectory struct {
	Init ExtendedDirectoryInit
	//TODO: indexes []DirectoryIndex
}

//NewExtendedDirectory creates a new ExtendedDirectory
func NewExtendedDirectory(rdr *io.Reader) (*ExtendedDirectory, error) {
	var inode ExtendedDirectory
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	//TODO: Read directory indexes
	return &inode, err
}

//BasicFile is self explainatory
type BasicFile struct {
	BlockStart     uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	Size           uint32
	BlockSizes     []uint32
	//TODO: possibly fix BlockSizes
}

//ExtendedFile is a file with additional information
type ExtendedFile struct {
	BlockStart     uint32
	Size           uint32
	Sparse         uint64
	HardLinks      uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	XattrIndex     uint32
	BlockSizes     []uint32
	//TODO: possibly fix BlockSizes
}

type BasicSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

type BasicSymlink struct {
	Init       BasicSymlinkInit
	targetPath []uint8 //len is TargetPathSize
}

func NewBasicSymlink(rdr *io.Reader) (*BasicSymlink, error) {
	var inode BasicSymlink
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	if err != nil {
		return nil, err
	}
	inode.targetPath = make([]uint8, inode.Init.TargetPathSize)
	err = binary.Read(*rdr, binary.LittleEndian, inode.targetPath)
	return &inode, err
}

type ExtendedSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

type ExtendedSymlink struct {
	Init       ExtendedSymlinkInit
	TargetPath []uint8
	XattrIndex uint32
}

func NewExtendedSymlink(rdr *io.Reader) (*ExtendedSymlink, error) {
	var inode ExtendedSymlink
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	if err != nil {
		return &inode, err
	}
	inode.TargetPath = make([]uint8, inode.Init.TargetPathSize)
	err = binary.Read(*rdr, binary.LittleEndian, &inode.XattrIndex)
	return &inode, err
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
