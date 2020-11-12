package inode

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

const (
	basicDirectory = iota + 1
	basicFile
	basicSymlink
	basicBlockDevice
	basicCharDevice
	basicFifo
	basicSocket
	extendedDirectory
	extendedFile
	extendedSymlink
	extendedBlockDevice
	extendedCharDevice
	extendedFifo
	extendedSocket
)

func ProcessInode(rdr *io.Reader) (*InodeCommon, interface{}, error) {
	var inodeHeader InodeCommon
	err := binary.Read(*rdr, binary.LittleEndian, &inodeHeader)
	if err != nil {
		return nil, nil, err
	}
	switch inodeHeader.InodeType {
	case basicDirectory:
		var inode BasicDirectory
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case basicFile:
		var inode BasicFile
		err = binary.Read(*rdr, binary.LittleEndian, &inode)
		return &inodeHeader, inode, err
	case basicSymlink:
		inode, err := NewBasicSymlink(rdr)
		return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	// case basicFile:
	// 	var inode BasicFile
	// 	err = binary.Read(*rdr, binary.LittleEndian, &inode)
	// 	return &inodeHeader, inode, err
	//TODO: implement ALL cases
	default:
		return nil, nil, errors.New("Inode type is unrecognized: " + strconv.FormatInt(int64(inodeHeader.InodeType), 2))
	}
}
