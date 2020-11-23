package inode

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	BasicDirectoryType = iota + 1
	BasicFileType
	BasicSymlinkType
	BasicBlockDeviceType
	BasicCharDeviceType
	BasicFifoType
	BasicSocketType
	ExtDirType
	ExtFileType
	ExtSymlinkType
	ExtBlockDeviceType
	ExtCharDeviceType
	ExtFifoType
	ExtSocketType
)

//Header is the common header for all inodes
type Header struct {
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
	Init    ExtendedDirectoryInit
	Indexes []DirectoryIndex
}

//NewExtendedDirectory creates a new ExtendedDirectory
func NewExtendedDirectory(rdr io.Reader) (ExtendedDirectory, error) {
	var inode ExtendedDirectory
	err := binary.Read(rdr, binary.LittleEndian, &inode.Init)
	if err != nil {
		return inode, err
	}
	if inode.Init.IndexCount > 0 {
		inode.Indexes = make([]DirectoryIndex, inode.Init.IndexCount)
		for i := uint16(0); i < inode.Init.IndexCount; i++ {
			inode.Indexes[i], err = NewDirectoryIndex(rdr)
			if err != nil {
				fmt.Println("Error while reading Directory Index ", i)
				return inode, err
			}
		}
	}
	return inode, err
}

//DirectoryIndexInit holds the values that can be easily decoded
type DirectoryIndexInit struct {
	Offset         uint32
	DirTableOffset uint32
	NameSize       uint32
}

//DirectoryIndex is a quick lookup provided by an ExtendedDirectory
type DirectoryIndex struct {
	Init DirectoryIndexInit
	Name []byte
}

//NewDirectoryIndex return a new DirectoryIndex
func NewDirectoryIndex(rdr io.Reader) (DirectoryIndex, error) {
	var index DirectoryIndex
	err := binary.Read(rdr, binary.LittleEndian, &index.Init)
	if err != nil {
		return index, err
	}
	index.Name = make([]byte, index.Init.NameSize, index.Init.NameSize)
	err = binary.Read(rdr, binary.LittleEndian, &index.Name)
	return index, err
}

//BasicFileInit is the information that can be directoy decoded
type BasicFileInit struct {
	BlockStart     uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	Size           uint32
}

//BasicFile is self explainatory
type BasicFile struct {
	Init       BasicFileInit
	BlockSizes []uint32
	Fragmented bool
}

//NewBasicFile creates a new BasicFile
func NewBasicFile(rdr io.Reader, blockSize uint32) (BasicFile, error) {
	var inode BasicFile
	err := binary.Read(rdr, binary.LittleEndian, &inode.Init)
	if err != nil {
		return inode, err
	}
	inode.Fragmented = inode.Init.FragmentIndex != 0xFFFFFFFF
	blocks := inode.Init.Size / blockSize
	if inode.Init.Size%blockSize > 0 {
		blocks++
	}
	inode.BlockSizes = make([]uint32, blocks, blocks)
	err = binary.Read(rdr, binary.LittleEndian, &inode.BlockSizes)
	return inode, err
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
}

//ExtendedFile is a file with more information
type ExtendedFile struct {
	Init       ExtendedFileInit
	BlockSizes []uint32
	Fragmented bool
}

//NewExtendedFile creates a new ExtendedFile
func NewExtendedFile(rdr io.Reader, blockSize uint32) (ExtendedFile, error) {
	var inode ExtendedFile
	err := binary.Read(rdr, binary.LittleEndian, &inode.Init)
	if err != nil {
		return inode, err
	}
	inode.Fragmented = inode.Init.FragmentIndex != 0xFFFFFFFF
	blocks := inode.Init.Size / blockSize
	if inode.Init.Size%blockSize > 0 {
		blocks++
	}
	inode.BlockSizes = make([]uint32, blocks, blocks)
	err = binary.Read(rdr, binary.LittleEndian, &inode.BlockSizes)
	return inode, err
}

//BasicSymlinkInit is all the values that can be directly decoded
type BasicSymlinkInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

//BasicSymlink is a symlink
type BasicSymlink struct {
	Init       BasicSymlinkInit
	targetPath []byte //len is TargetPathSize
}

//NewBasicSymlink creates a new BasicSymlink
func NewBasicSymlink(rdr io.Reader) (BasicSymlink, error) {
	var inode BasicSymlink
	err := binary.Read(rdr, binary.LittleEndian, &inode.Init)
	if err != nil {
		return inode, err
	}
	inode.targetPath = make([]byte, inode.Init.TargetPathSize, inode.Init.TargetPathSize)
	err = binary.Read(rdr, binary.LittleEndian, &inode.targetPath)
	return inode, err
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
func NewExtendedSymlink(rdr io.Reader) (ExtendedSymlink, error) {
	var inode ExtendedSymlink
	err := binary.Read(rdr, binary.LittleEndian, &inode.Init)
	if err != nil {
		return inode, err
	}
	inode.TargetPath = make([]uint8, inode.Init.TargetPathSize, inode.Init.TargetPathSize)
	err = binary.Read(rdr, binary.LittleEndian, &inode.XattrIndex)
	return inode, err
}

//BasicDevice is a device
type BasicDevice struct {
	HardLinks uint32
	Device    uint32
}

//ExtendedDevice is a device with more info
type ExtendedDevice struct {
	BasicDevice
	XattrIndex uint32
}

//BasicIPC is a Fifo or Socket device
type BasicIPC struct {
	HardLink uint32
}

//ExtendedIPC is a IPC device with extra info
type ExtendedIPC struct {
	BasicIPC
	XattrIndex uint32
}
