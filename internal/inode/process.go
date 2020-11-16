package inode

import (
	"encoding/binary"
	"io"
)

//Inode holds an inode. Header is the header that's common for all inodes.
//
//Info holds the actual Inode. Due to each inode type being a different type, it's store as an interface{}
type Inode struct {
	Header Header
	Type   int         //Type the inode type defined in the header. Here so it's easy to access
	Info   interface{} //Info is the parsed specific data. It's type is defined by Type.
}

//ProcessInode tries to read an inode from the BlockReader
func ProcessInode(br io.Reader, blockSize uint32) (Inode, error) {
	var head Header
	err := binary.Read(br, binary.LittleEndian, &head)
	if err != nil {
		return Inode{}, err
	}
	var info interface{}
	switch head.InodeType {
	case BasicDirectoryType:
		var inode BasicDirectory
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case BasicFileType:
		inode, err := NewBasicFile(br, blockSize)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case BasicSymlinkType:
		inode, err := NewBasicSymlink(br)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case BasicBlockDeviceType:
		var inode BasicDevice
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case BasicCharDeviceType:
		var inode BasicDevice
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case BasicFifoType:
		var inode BasicIPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case BasicSocketType:
		var inode BasicIPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtDirType:
		inode, err := NewExtendedDirectory(br)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtFileType:
		inode, err := NewExtendedFile(br, blockSize)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtSymlinkType:
		inode, err := NewExtendedSymlink(br)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtBlockDeviceType:
		var inode ExtendedDevice
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtCharDeviceType:
		var inode ExtendedDevice
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtFifoType:
		var inode ExtendedIPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	case ExtSocketType:
		var inode ExtendedIPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return Inode{}, err
		}
		info = inode
	}
	return Inode{
		Type:   int(head.InodeType),
		Header: head,
		Info:   info,
	}, nil
}
