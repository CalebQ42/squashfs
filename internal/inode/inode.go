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

//BasicFileInit is the information that can be directoy decoded
type BasicFileInit struct {
	BlockStart     uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	Size           uint32
	//TODO: possibly fix BlockSizes
}

//BasicFile is self explainatory
type BasicFile struct {
	Init       BasicFileInit
	BlockSizes []uint32
}

//NewBasicFile creates a new BasicFile
func NewBasicFile(rdr *io.Reader, blockSize uint32) (*BasicFile, error) {
	var inode BasicFile
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	if err != nil {
		return &inode, err
	}
	blocks := inode.Init.Size / blockSize
	if inode.Init.Size%blockSize > 0 {
		blocks++
	}
	inode.BlockSizes = make([]uint32, blocks, blocks)
	err = binary.Read(*rdr, binary.LittleEndian, inode.BlockSizes)
	return &inode, err
}

//ExtendedFileInit is the information that can be directly decoded
type ExtendedFileInit struct {
	BlockStart     uint32
	Size           uint32
	Sparse         uint64
	HardLinks      uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	XattrIndex     uint32
	//TODO: possibly fix BlockSizes
}

//ExtendedFile is a file with more information
type ExtendedFile struct {
	Init       ExtendedFileInit
	BlockSizes []uint32
}

//NewExtendedFile creates a new ExtendedFile
func NewExtendedFile(rdr *io.Reader, blockSize uint32) (*ExtendedFile, error) {
	var inode ExtendedFile
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	if err != nil {
		return &inode, err
	}
	blocks := inode.Init.Size / blockSize
	if inode.Init.Size%blockSize > 0 {
		blocks++
	}
	inode.BlockSizes = make([]uint32, blocks, blocks)
	err = binary.Read(*rdr, binary.LittleEndian, inode.BlockSizes)
	return &inode, err
}

//BasicSymlinkInit is all the values that can be directly decoded
type BasicSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

//BasicSymlink is a symlink
type BasicSymlink struct {
	Init       BasicSymlinkInit
	targetPath []uint8 //len is TargetPathSize
}

//NewBasicSymlink creates a new BasicSymlink
func NewBasicSymlink(rdr *io.Reader) (*BasicSymlink, error) {
	var inode BasicSymlink
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	if err != nil {
		return nil, err
	}
	inode.targetPath = make([]uint8, inode.Init.TargetPathSize, inode.Init.TargetPathSize)
	err = binary.Read(*rdr, binary.LittleEndian, inode.targetPath)
	return &inode, err
}

//ExtendedSymlinkInit is all the values that can be directly decoded
type ExtendedSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

//ExtendedSymlink is a symlink with extra information
type ExtendedSymlink struct {
	Init       ExtendedSymlinkInit
	TargetPath []uint8
	XattrIndex uint32
}

//NewExtendedSymlink creates a new ExtendedSymlink
func NewExtendedSymlink(rdr *io.Reader) (*ExtendedSymlink, error) {
	var inode ExtendedSymlink
	err := binary.Read(*rdr, binary.LittleEndian, inode.Init)
	if err != nil {
		return &inode, err
	}
	inode.TargetPath = make([]uint8, inode.Init.TargetPathSize, inode.Init.TargetPathSize)
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
