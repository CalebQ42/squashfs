package inode

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

const (
	//The inode type from inode.Common.InodeType

	BasicDirectoryType = iota + 1
	BasicFileType
	BasicSymlinkType
	BasicBlockDeviceType
	BasicCharDeviceType
	BasicFifoType
	BasicSocketType
	ExtendedDirectoryType
	ExtendedFileType
	ExtendedSymlinkType
	ExtendedBlockDeviceType
	ExtendedCharDeviceType
	ExtendedFifoType
	ExtendedSocketType
)

//ProcessInode processes the next inode in the given reader
func ProcessInode(rdr *io.Reader, blockSize uint32) (*Common, interface{}, error) {
	var inodeHeader Common
	err := binary.Read(*rdr, binary.LittleEndian, &inodeHeader)
	if err != nil {
		return nil, nil, err
	}
	switch inodeHeader.InodeType {
	case BasicDirectoryType:
		var inode BasicDirectory
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, &inode, err
	case BasicFileType:
		inode, err := NewBasicFile(rdr, blockSize)
		return &inodeHeader, inode, err
	case BasicSymlinkType:
		inode, err := NewBasicSymlink(rdr)
		return &inodeHeader, inode, err
	case BasicBlockDeviceType:
		var inode BasicDevice
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case BasicCharDeviceType:
		var inode BasicDevice
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case BasicFifoType:
		var inode BasicIPC
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case BasicSocketType:
		var inode BasicIPC
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case ExtendedDirectoryType:
		inode, err := NewExtendedDirectory(rdr)
		return &inodeHeader, inode, err
	case ExtendedFileType:
		inode, err := NewExtendedFile(rdr, blockSize)
		return &inodeHeader, inode, err
	case ExtendedSymlinkType:
		inode, err := NewExtendedSymlink(rdr)
		return &inodeHeader, inode, err
	case ExtendedBlockDeviceType:
		var inode ExtendedDevice
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case ExtendedCharDeviceType:
		var inode ExtendedDevice
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case ExtendedFifoType:
		var inode ExtendedIPC
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case ExtendedSocketType:
		var inode ExtendedIPC
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	//TODO: implement ALL cases
	default:
		return nil, nil, errors.New("Inode type is unrecognized: " + strconv.FormatInt(int64(inodeHeader.InodeType), 2))
	}
}
