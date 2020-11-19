package squashfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/directory"
	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

const (
	magic = 0x73717368
)

type Reader struct {
	r            io.ReaderAt
	super        Superblock
	flags        SuperblockFlags
	decompressor Decompressor
	dirs         []*directory.Directory
}

func NewSquashfsReader(r io.ReaderAt) (*Reader, error) {
	var rdr Reader
	rdr.r = r
	err := binary.Read(io.NewSectionReader(rdr.r, 0, int64(binary.Size(rdr.super))), binary.LittleEndian, &rdr.super)
	if err != nil {
		return nil, err
	}
	if rdr.super.Magic != magic {
		return nil, errors.New("doesn't have magic number, probably isn't a squashfs")
	}
	rdr.flags = rdr.super.GetFlags()
	switch rdr.super.CompressionType {
	case gzipCompression:
		rdr.decompressor = &ZlibDecompressor{}
	default:
		return nil, errors.New("Unsupported compression type")
	}
	if rdr.flags.CompressorOptions {
		//TODO: parse compressor options
		fmt.Println("Compressor options is NOT currently supported")
		return nil, errors.New("Has compressor options")
	}
	return &rdr, nil
}

func (r *Reader) readDir(i *inode.Inode) (paths []string, err error) {
	dir, err := r.ReadDirFromInode(*i)
	if err != nil {
		return
	}
	for _, entry := range dir.Entries {
		if entry.Init.Type == inode.BasicDirectoryType {
			paths = append(paths)
			i, err = r.GetInodeFromEntry(&entry)
			if err != nil {
				return
			}
			var subPaths []string
			subPaths, err = r.readDir(i)
			if err != nil {
				return
			}
			for pathI := range subPaths {
				subPaths[pathI] = entry.Name + "/" + subPaths[pathI]
			}
			paths = append(paths, entry.Name+"/")
			paths = append(paths, subPaths...)
		} else {
			paths = append(paths, entry.Name)
		}
	}
	return
}

func (r *Reader) readDirTable() error {
	inoderdr, err := r.NewBlockReaderFromInodeRef(r.super.RootInodeRef)
	if err != nil {
		return err
	}
	i, err := inode.ProcessInode(inoderdr, r.super.BlockSize)
	if err != nil {
		return err
	}
	paths, err := r.readDir(&i)
	if err != nil {
		return err
	}
	for _, path := range paths {
		fmt.Println(path)
	}
	return nil
}
