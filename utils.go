package squashfs

import (
	"errors"
	"io"
	"strings"

	"github.com/CalebQ42/GoSquashfs/internal/directory"
	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

var (
	//ErrNotFound means that the given path is NOT present in the archive
	ErrNotFound = errors.New("Path not found")
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

//ReadDirFromInode returns a fully populated directory.Directory from a given inode.Inode.
//If the given inode is not a directory it returns an error.
func (r *Reader) ReadDirFromInode(i *inode.Inode) (*directory.Directory, error) {
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
	br, err := r.NewMetadataReader(int64(r.super.DirTableStart + uint64(offset)))
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

//GetInodeFromEntry returns the inode associated with a given directory.Entry
func (r *Reader) GetInodeFromEntry(en *directory.Entry) (*inode.Inode, error) {
	br, err := r.NewMetadataReader(int64(r.super.InodeTableStart + uint64(en.Header.InodeOffset)))
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
	return i, nil
}

//GetInodeFromPath returns the inode at the given path, relative to root.
//The given path can start or without "/".
func (r *Reader) GetInodeFromPath(path string) (*inode.Inode, error) {
	path = strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/")
	pathDirs := strings.Split(path, "/")
	rdr, err := r.NewMetadataReaderFromInodeRef(r.super.RootInodeRef)
	if err != nil {
		return nil, err
	}
	curInodeDir, err := inode.ProcessInode(rdr, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return curInodeDir, nil
	}
	for depth := 0; depth < len(pathDirs); depth++ {
		if curInodeDir.Type != inode.BasicDirectoryType && curInodeDir.Type != inode.ExtDirType {
			return nil, ErrNotFound
		}
		dir, err := r.ReadDirFromInode(curInodeDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range dir.Entries {
			if entry.Name == pathDirs[depth] {
				if depth == len(pathDirs)-1 {
					in, err := r.GetInodeFromEntry(&entry)
					if err != nil {
						return nil, err
					}
					return in, nil
				}
				curInodeDir, err = r.GetInodeFromEntry(&entry)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return nil, ErrNotFound
}
