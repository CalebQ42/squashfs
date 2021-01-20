package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"

	"github.com/CalebQ42/squashfs/internal/compression"
	"github.com/CalebQ42/squashfs/internal/inode"
)

const (
	magic uint32 = 0x73717368
)

var (
	//ErrNoMagic is returned if the magic number in the superblock isn't correct.
	errNoMagic = errors.New("Magic number doesn't match. Either isn't a squashfs or corrupted")
	//ErrIncompatibleCompression is returned if the compression type in the superblock doesn't work.
	errIncompatibleCompression = errors.New("Compression type unsupported")
	//ErrCompressorOptions is returned if compressor options is present. It's not currently supported.
	errCompressorOptions = errors.New("Compressor options is not currently supported")
	//ErrOptions is returned when compression options that I haven't tested is set. When this is returned, the Reader is also returned.
	ErrOptions = errors.New("Possibly incompatible compressor options")
)

//TODO: implement fs.FS, possibly more FS types for compatibility. Most of this work will mostly be handed off to root anyway so this shouldn't be too difficult.

//Reader processes and reads a squashfs archive.
type Reader struct {
	r            io.ReaderAt
	decompressor compression.Decompressor
	root         *File
	fragOffsets  []uint64
	idTable      []uint32
	super        superblock
	flags        SuperblockFlags
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
	if rdr.super.BlockLog != uint16(math.Log2(float64(rdr.super.BlockSize))) {
		return nil, errors.New("BlockSize and BlockLog doesn't match. The archive is probably corrupt")
	}
	hasUnsupportedOptions := false
	rdr.flags = rdr.super.GetFlags()
	if rdr.flags.compressorOptions {
		switch rdr.super.CompressionType {
		case GzipCompression:
			var gzip *compression.Gzip
			gzip, err = compression.NewGzipCompressorWithOptions(io.NewSectionReader(rdr.r, int64(binary.Size(rdr.super)), 8))
			if err != nil {
				return nil, err
			}
			if gzip.HasCustomWindow || gzip.HasStrategies {
				hasUnsupportedOptions = true
			}
			rdr.decompressor = gzip
		case XzCompression:
			var xz *compression.Xz
			xz, err = compression.NewXzCompressorWithOptions(io.NewSectionReader(rdr.r, int64(binary.Size(rdr.super)), 8))
			if err != nil {
				return nil, err
			}
			if xz.HasFilters {
				return nil, errors.New("XZ compression options has filters. These are not yet supported")
			}
			rdr.decompressor = xz
		case Lz4Compression:
			var lz4 *compression.Lz4
			lz4, err = compression.NewLz4CompressorWithOptions(io.NewSectionReader(rdr.r, int64(binary.Size(rdr.super)), 8))
			if err != nil {
				return nil, err
			}
			rdr.decompressor = lz4
		case ZstdCompression:
			var zstd *compression.Zstd
			zstd, err = compression.NewZstdCompressorWithOptions(io.NewSectionReader(rdr.r, int64(binary.Size(rdr.super)), 4))
			if err != nil {
				return nil, err
			}
			rdr.decompressor = zstd
		default:
			return nil, errIncompatibleCompression
		}
	} else {
		switch rdr.super.CompressionType {
		case GzipCompression:
			rdr.decompressor = &compression.Gzip{}
		case LzmaCompression:
			rdr.decompressor = &compression.Lzma{}
		case XzCompression:
			rdr.decompressor = &compression.Xz{}
		case Lz4Compression:
			rdr.decompressor = &compression.Lz4{}
		case ZstdCompression:
			rdr.decompressor = &compression.Zstd{}
		default:
			//TODO: all compression types.
			return nil, errIncompatibleCompression
		}
	}
	fragBlocks := int(math.Ceil(float64(rdr.super.FragCount) / 512))
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
	unread := rdr.super.IDCount
	blockOffsets := make([]uint64, int(math.Ceil(float64(rdr.super.IDCount)/2048)))
	for i := range blockOffsets {
		secRdr := io.NewSectionReader(r, int64(rdr.super.IDTableStart)+(8*int64(i)), 8)
		err = binary.Read(secRdr, binary.LittleEndian, &blockOffsets[i])
		if err != nil {
			return nil, err
		}
		idRdr, err := rdr.newMetadataReader(int64(blockOffsets[i]))
		if err != nil {
			return nil, err
		}
		read := uint16(math.Min(float64(unread), 2048))
		for i := uint16(0); i < read; i++ {
			var tmp uint32
			err = binary.Read(idRdr, binary.LittleEndian, &tmp)
			if err != nil {
				return nil, err
			}
			rdr.idTable = append(rdr.idTable, tmp)
		}
		unread -= read
	}
	if hasUnsupportedOptions {
		return &rdr, ErrOptions
	}
	return &rdr, nil
}

//ModTime is the last time the file was modified/created.
func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.super.CreationTime), 0)
}

//ExtractTo tries to extract ALL files to the given path. This is the same as getting the root folder and extracting that.
func (r *Reader) ExtractTo(path string) []error {
	if r.root == nil {
		_, err := r.GetRootFolder()
		if err != nil {
			return []error{err}
		}
	}
	return r.root.ExtractTo(path)
}

//GetRootFolder returns a squashfs.File that references the root directory of the squashfs archive.
func (r *Reader) GetRootFolder() (*File, error) {
	if r.root != nil {
		return r.root, nil
	}
	mr, err := r.newMetadataReaderFromInodeRef(r.super.RootInodeRef)
	if err != nil {
		return nil, err
	}
	var root File
	root.in, err = inode.ProcessInode(mr, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	root.dir = "/"
	root.filType = root.in.Type
	root.r = r
	r.root = &root
	return r.root, nil
}

//GetAllFiles returns a slice of ALL files and folders contained in the squashfs.
func (r *Reader) GetAllFiles() (fils []*File, err error) {
	if r.root == nil {
		_, err := r.GetRootFolder()
		if err != nil {
			return nil, err
		}
	}
	return r.root.GetChildrenRecursively()
}

//FindFile returns the first file (in the same order as Reader.GetAllFiles) that the given function returns true for. Returns nil if nothing is found.
func (r *Reader) FindFile(query func(*File) bool) *File {
	if r.root == nil {
		_, err := r.GetRootFolder()
		if err != nil {
			return nil
		}
	}
	fils, err := r.root.GetChildren()
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
	if r.root == nil {
		_, err := r.GetRootFolder()
		if err != nil {
			return nil
		}
	}
	fils, err := r.root.GetChildrenRecursively()
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
func (r *Reader) GetFileAtPath(filepath string) *File {
	if r.root == nil {
		_, err := r.GetRootFolder()
		if err != nil {
			return nil
		}
	}
	return r.root.GetFileAtPath(filepath)
}
