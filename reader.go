package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/CalebQ42/squashfs/internal/compression"
	"github.com/CalebQ42/squashfs/internal/inode"
)

const (
	magic = 0x73717368
)

var (
	//ErrNoMagic is returned if the magic number in the superblock isn't correct.
	errNoMagic = errors.New("Magic number doesn't match. Either isn't a squashfs or corrupted")
	//ErrIncompatibleCompression is returned if the compression type in the superblock doesn't work.
	errIncompatibleCompression = errors.New("Compression type unsupported")
	//ErrCompressorOptions is returned if compressor options is present. It's not currently supported.
	errCompressorOptions = errors.New("Compressor options is not currently supported")
)

//Reader processes and reads a squashfs archive.
type Reader struct {
	r            io.ReaderAt
	super        superblock
	flags        superblockFlags
	decompressor compression.Decompressor
	fragOffsets  []uint64
}

//NewSquashfsReader returns a new squashfs.Reader from an io.ReaderAt
func NewSquashfsReader(r io.ReaderAt) (*Reader, error) {
	var rdr Reader
	rdr.r = r
	err := binary.Read(io.NewSectionReader(rdr.r, 0, int64(binary.Size(rdr.super))), binary.LittleEndian, &rdr.super)
	if err != nil {
		return nil, err
	}
	if rdr.super.Magic != magic {
		return nil, errNoMagic
	}
	rdr.flags = rdr.super.GetFlags()
	switch rdr.super.CompressionType {
	case gzipCompression:
		rdr.decompressor = &compression.Zlib{}
	case xzCompression:
		rdr.decompressor = &compression.Xz{}
	default:
		return nil, errIncompatibleCompression
	}
	if rdr.flags.CompressorOptions {
		//TODO: parse compressor options
		return nil, errCompressorOptions
	}
	fragBlocks := int(math.Ceil(float64(rdr.super.FragCount) / 512.0))
	if fragBlocks > 0 {
		offset := int64(rdr.super.FragTableStart)
		for i := 0; i < fragBlocks; i++ {
			tmp := make([]byte, 8)
			_, err = r.ReadAt(tmp, offset)
			if err != nil {
				return nil, err
			}
			rdr.fragOffsets = append(rdr.fragOffsets, binary.LittleEndian.Uint64(tmp))
			offset += 8
		}
	}
	return &rdr, nil
}

//GetRootFolder returns a squashfs.File that references the root directory of the squashfs archive.
func (r *Reader) GetRootFolder() (root *File, err error) {
	root = new(File)
	mr, err := r.newMetadataReaderFromInodeRef(r.super.RootInodeRef)
	if err != nil {
		return nil, err
	}
	root.in, err = inode.ProcessInode(mr, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	root.path = "/"
	root.filType = root.in.Type
	root.r = r
	return root, nil
}

//GetAllFiles returns a slice of ALL files and folders contained in the squashfs.
func (r *Reader) GetAllFiles() (fils []*File, err error) {
	root, err := r.GetRootFolder()
	if err != nil {
		return nil, err
	}
	return root.GetChildrenRecursively()
}

//FindFile returns the first file (in the same order as Reader.GetAllFiles) that the given function returns true for. Returns nil if nothing is found.
func (r *Reader) FindFile(query func(*File) bool) *File {
	root, err := r.GetRootFolder()
	if err != nil {
		return nil
	}
	fils, err := root.GetChildren()
	if err != nil {
		return nil
	}
	var childrenDirs []*File
	for _, fil := range fils {
		if query(fil) {
			return fil
		}
		if fil.IsDir() {
			childrenDirs = append(childrenDirs, fil)
		}
	}
	for len(childrenDirs) != 0 {
		var tmp []*File
		for _, dirs := range childrenDirs {
			chil, err := dirs.GetChildren()
			if err != nil {
				return nil
			}
			for _, child := range chil {
				if query(child) {
					return child
				}
				if child.IsDir() {
					tmp = append(tmp, child)
				}
			}
		}
		childrenDirs = tmp
	}
	return nil
}

//FindAll returns all files where the given function returns true.
func (r *Reader) FindAll(query func(*File) bool) (all []*File) {
	root, err := r.GetRootFolder()
	if err != nil {
		return nil
	}
	fils, err := root.GetChildrenRecursively()
	if err != nil {
		return nil
	}
	for _, fil := range fils {
		if query(fil) {
			all = append(all, fil)
		}
	}
	return
}

//GetFileAtPath will return the file at the given path. If the file cannot be found, will return nil.
func (r *Reader) GetFileAtPath(path string) *File {
	dir, err := r.GetRootFolder()
	if err != nil {
		return nil
	}
	return dir.GetFileAtPath(path)
}
