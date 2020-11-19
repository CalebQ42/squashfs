package squashfs

import (
	"errors"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/directory"
	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

//ProcessInodeRef processes an inode reference and returns two values
//
//The first value is the inode table offset. AKA, it's where the metadata block of the inode STARTS relative to the inode table.
//
//The second value is the offset of the inode, INSIDE of the metadata.
func processInodeRef(inodeRef uint64) (tableOffset uint64, metaOffset uint64) {
	tableOffset = inodeRef >> 16
	metaOffset = inodeRef &^ 0xFFFFFFFF0000
	return
}

func (r *Reader) ReadDirFromInode(i inode.Inode) (*directory.Directory, error) {
	var offset uint32
	var metaOffset uint16
	var size uint16
	switch i.Type {
	case inode.BasicDirectoryType:
		offset = i.Info.(inode.BasicDirectory).DirectoryIndex
		metaOffset = i.Info.(inode.BasicDirectory).DirectoryOffset
		size = i.Info.(inode.BasicDirectory).DirectorySize
	case inode.ExtDirType:
		offset = i.Info.(inode.ExtendedDirectory).Init.DirectoryIndex
		metaOffset = i.Info.(inode.ExtendedDirectory).Init.DirectoryOffset
		size = uint16(i.Info.(inode.ExtendedDirectory).Init.DirectorySize)
	default:
		return nil, errors.New("Not a directory inode")
	}
	br, err := r.NewBlockReader(int64(r.super.DirTableStart + uint64(offset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(metaOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	dir, err := directory.NewDirectory(br, size)
	if err != nil {
		return dir, err
	}
	return dir, nil
}

func (r *Reader) GetInodeFromEntry(en *directory.Entry) (*inode.Inode, error) {
	br, err := r.NewBlockReader(int64(r.super.InodeTableStart + uint64(en.Header.InodeOffset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(en.Init.Offset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	i, err := inode.ProcessInode(br, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	return &i, nil
}
