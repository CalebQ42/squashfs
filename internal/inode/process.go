package inode

import (
	"encoding/binary"
	"io"
)

//Inode holds an inode. Header is the header that's common for all inodes.
//
//Info holds the actual Inode. Due to each inode type being a different type, it's store as an interface{}
type Inode struct {
	Info interface{} //Info is the parsed specific data. It's type is defined by Type.
	Type int         //Type the inode type defined in the header. Here so it's easy to access
	Header
}

//ProcessInode tries to read an inode from the BlockReader
func ProcessInode(br io.Reader, blockSize uint32) (*Inode, error) {
	var head Header
	err := binary.Read(br, binary.LittleEndian, &head)
	if err != nil {
		return nil, err
	}
	var info interface{}
	switch head.InodeType {
	case DirType:
		var inode Dir
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case FileType:
		var inode File
		inode, err = NewFile(br, blockSize)
		if err != nil {
			return nil, err
		}
		info = inode
	case SymType:
		var inode Sym
		inode, err = NewSymlink(br)
		if err != nil {
			return nil, err
		}
		info = inode
	case BlockDevType:
		var inode Device
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case CharDevType:
		var inode Device
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case FifoType:
		var inode IPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case SocketType:
		var inode IPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtDirType:
		var inode ExtDir
		inode, err = NewExtendedDirectory(br)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtFileType:
		var inode ExtFile
		inode, err = NewExtendedFile(br, blockSize)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtSymType:
		var inode ExtSym
		inode, err = NewExtendedSymlink(br)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtBlockDeviceType:
		var inode ExtDevice
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtCharDeviceType:
		var inode ExtDevice
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtFifoType:
		var inode ExtIPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	case ExtSocketType:
		var inode ExtIPC
		err = binary.Read(br, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		info = inode
	}
	return &Inode{
		Type:   int(head.InodeType),
		Header: head,
		Info:   info,
	}, nil
}
