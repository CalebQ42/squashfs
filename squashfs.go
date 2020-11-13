package squashfs

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/directory"
	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

type Squashfs struct {
	r           Reader
	super       Superblock
	flags       SuperblockFlags
	compression CompressionOptions
}

func NewSquashfs(rdr io.ReaderAt) (*Squashfs, error) {
	var squash Squashfs
	squash.r = NewReader(rdr)
	err := binary.Read(&squash.r, binary.LittleEndian, &squash.super)
	if err != nil {
		return nil, err
	}
	squash.flags = squash.super.getFlags()
	switch squash.super.Compression {
	case gzipCompression:
		var raw gzipOptionsRaw
		err := binary.Read(&squash.r, binary.LittleEndian, &raw)
		if err != nil {
			return nil, err
		}
		squash.compression = NewGzipOptions(raw)
	default:
		fmt.Println("Other compression options are not currently supported")
		return nil, err
	}
	return &squash, nil
}

func (s *Squashfs) printDirTable() error {
	offset, metaOffset := inode.ProcessInodeRef(s.super.RootInode)
	br, err := s.NewBlockReader(int64(offset))
	if err != nil {
		return err
	}
	fmt.Println(offset, metaOffset)
	br.dataOffset = int64(metaOffset)
	_, inodeType, err := inode.ProcessInode(br, s.super.BlockSize)
	if err != nil {
		return err
	}
	rootDir := inodeType.(*inode.BasicDirectory)
	fmt.Println(*rootDir)
	br, err = s.NewBlockReader(int64(s.super.DirectoryTableOffset) + int64(rootDir.DirectoryIndex))
	if err != nil {
		return err
	}
	br.dataOffset = int64(rootDir.DirectoryOffset)
	dir, err := directory.NewDirectory(br)
	if err != nil {
		return err
	}
	for _, entry := range dir.Entries {
		fmt.Println(entry.Name)
	}
	return nil
}

//GetFlags returns the SuperblockFlags from the Superblock
func (s *Squashfs) GetFlags() SuperblockFlags {
	return s.flags
}
