package inode

import (
	"encoding/binary"
	"io"
)

//The different types of inodes as defined by inodetype
const (
	DirType = iota + 1
	FileType
	SymType
	BlockDevType
	CharDevType
	FifoType
	SocketType
	ExtDirType
	ExtFileType
	ExtSymType
	ExtBlockDeviceType
	ExtCharDeviceType
	ExtFifoType
	ExtSocketType
)

//Header is the common header for all inodes
type Header struct {
	Type         uint16
	Permissions  uint16
	UID          uint16
	GID          uint16
	ModifiedTime uint32
	Number       uint32
}

//Dir is self explainatory
type Dir struct {
	DirectoryIndex    uint32
	HardLinks         uint32
	DirectorySize     uint16
	DirectoryOffset   uint16
	ParentInodeNumber uint32
}

//ExtDirInit is the information that can be directoy decoded
type ExtDirInit struct {
	HardLinks         uint32
	DirectorySize     uint32
	DirectoryIndex    uint32
	ParentInodeNumber uint32
	IndexCount        uint16 //one less then directory indexes following structure
	DirectoryOffset   uint16
	XattrIndex        uint32
}

//ExtDir is a directory with extra info
type ExtDir struct {
	Indexes []DirIndex
	ExtDirInit
}

//NewExtendedDirectory creates a new ExtendedDirectory
func NewExtendedDirectory(rdr io.Reader) (ExtDir, error) {
	var inode ExtDir
	err := binary.Read(rdr, binary.LittleEndian, &inode.ExtDirInit)
	if err != nil {
		return inode, err
	}
	for i := uint16(0); i < inode.IndexCount; i++ {
		var tmp DirIndex
		tmp, err = NewDirectoryIndex(rdr)
		if err != nil {
			return inode, err
		}
		inode.Indexes = append(inode.Indexes, tmp)
	}
	return inode, err
}

//DirIndexInit holds the values that can be easily decoded
type DirIndexInit struct {
	Offset         uint32
	DirTableOffset uint32
	NameSize       uint32
}

//DirIndex is a quick lookup provided by an ExtendedDirectory
type DirIndex struct {
	Name string
	DirIndexInit
}

//NewDirectoryIndex return a new DirectoryIndex
func NewDirectoryIndex(rdr io.Reader) (DirIndex, error) {
	var index DirIndex
	err := binary.Read(rdr, binary.LittleEndian, &index.DirIndexInit)
	if err != nil {
		return index, err
	}
	tmp := make([]byte, index.NameSize+1, index.NameSize+1)
	err = binary.Read(rdr, binary.LittleEndian, &tmp)
	if err != nil {
		return index, err
	}
	index.Name = string(tmp)
	return index, nil
}

//FileInit is the information that can be directly decoded
type FileInit struct {
	BlockStart     uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	Size           uint32
}

//File is self explainatory
type File struct {
	BlockSizes []uint32
	Fragmented bool
	FileInit
}

//NewFile creates a new File
func NewFile(rdr io.Reader, blockSize uint32) (File, error) {
	var inode File
	err := binary.Read(rdr, binary.LittleEndian, &inode.FileInit)
	if err != nil {
		return inode, err
	}
	inode.Fragmented = inode.FragmentIndex != 0xFFFFFFFF
	blocks := inode.Size / blockSize
	if inode.Size%blockSize > 0 {
		blocks++
	}
	inode.BlockSizes = make([]uint32, blocks, blocks)
	err = binary.Read(rdr, binary.LittleEndian, &inode.BlockSizes)
	return inode, err
}

//ExtFileInit is the information that can be directly decoded
type ExtFileInit struct {
	BlockStart     uint64
	Size           uint64
	Sparse         uint64
	HardLinks      uint32
	FragmentIndex  uint32
	FragmentOffset uint32
	XattrIndex     uint32
}

//ExtFile is a file with more information
type ExtFile struct {
	BlockSizes []uint32
	Fragmented bool
	ExtFileInit
}

//NewExtendedFile creates a new ExtendedFile
func NewExtendedFile(rdr io.Reader, blockSize uint32) (ExtFile, error) {
	var inode ExtFile
	err := binary.Read(rdr, binary.LittleEndian, &inode.ExtFileInit)
	if err != nil {
		return inode, err
	}
	inode.Fragmented = inode.FragmentIndex != 0xFFFFFFFF
	blocks := inode.Size / uint64(blockSize)
	if inode.Size%uint64(blockSize) > 0 {
		blocks++
	}
	inode.BlockSizes = make([]uint32, blocks, blocks)
	err = binary.Read(rdr, binary.LittleEndian, &inode.BlockSizes)
	return inode, err
}

//SymInit is all the values that can be directly decoded
type SymInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

//Sym is a symlink
type Sym struct {
	Path       string
	targetPath []byte //len is TargetPathSize
	SymInit
}

//NewSymlink creates a new Symlink
func NewSymlink(rdr io.Reader) (Sym, error) {
	var inode Sym
	err := binary.Read(rdr, binary.LittleEndian, &inode.SymInit)
	if err != nil {
		return inode, err
	}
	inode.targetPath = make([]byte, inode.TargetPathSize, inode.TargetPathSize)
	err = binary.Read(rdr, binary.LittleEndian, &inode.targetPath)
	if err != nil {
		return inode, err
	}
	inode.Path = string(inode.targetPath)
	return inode, err
}

//ExtSymInit is all the values that can be directly decoded
type ExtSymInit struct {
	HardLinks      uint32
	TargetPathSize uint32
}

//ExtSym is a symlink with extra information
type ExtSym struct {
	Path       string
	targetPath []uint8
	ExtSymInit
	XattrIndex uint32
}

//NewExtendedSymlink creates a new ExtendedSymlink
func NewExtendedSymlink(rdr io.Reader) (ExtSym, error) {
	var inode ExtSym
	err := binary.Read(rdr, binary.LittleEndian, &inode.ExtSymInit)
	if err != nil {
		return inode, err
	}
	inode.targetPath = make([]uint8, inode.TargetPathSize, inode.TargetPathSize)
	err = binary.Read(rdr, binary.LittleEndian, &inode.targetPath)
	if err != nil {
		return inode, err
	}
	inode.Path = string(inode.targetPath)
	err = binary.Read(rdr, binary.LittleEndian, &inode.XattrIndex)
	return inode, err
}

//Device is a device
type Device struct {
	HardLinks uint32
	Device    uint32
}

//ExtDevice is a device with more info
type ExtDevice struct {
	Device
	XattrIndex uint32
}

//IPC is a Fifo or Socket device
type IPC struct {
	HardLink uint32
}

//ExtIPC is a IPC device with extra info
type ExtIPC struct {
	IPC
	XattrIndex uint32
}
