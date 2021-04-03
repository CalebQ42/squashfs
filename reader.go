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
	errNoMagic = errors.New("magic number doesn't match. Either isn't a squashfs or corrupted")
	//ErrIncompatibleCompression is returned if the compression type in the superblock doesn't work.
	errIncompatibleCompression = errors.New("compression type unsupported")
	//ErrOptions is returned when compression options that I haven't tested is set. When this is returned, the Reader is also returned.
	ErrOptions = errors.New("possibly incompatible compressor options")
)

//TODO: implement fs.FS, possibly more FS types for compatibility. Most of this work will mostly be handed off to root anyway so this shouldn't be too difficult.

//Reader processes and reads a squashfs archive.
type Reader struct {
	FS
	r            *io.SectionReader
	decompressor compression.Decompressor
	fragOffsets  []uint64
	idTable      []uint32
	super        superblock
	flags        SuperblockFlags
}

//NewSquashfsReader returns a new squashfs.Reader from an io.ReaderAt
func NewSquashfsReader(r io.ReaderAt) (*Reader, error) {
	var rdr Reader
	err := binary.Read(io.NewSectionReader(r, 0, int64(binary.Size(rdr.super))), binary.LittleEndian, &rdr.super)
	if err != nil {
		return nil, err
	}
	rdr.r = io.NewSectionReader(r, 0, int64(rdr.super.BytesUsed))
	if rdr.super.Magic != magic {
		return nil, errNoMagic
	}
	if rdr.super.BlockLog != uint16(math.Log2(float64(rdr.super.BlockSize))) {
		return nil, errors.New("BlockSize and BlockLog doesn't match. The archive is probably corrupt")
	}
	rdr.r.Seek(96, io.SeekStart)
	hasUnsupportedOptions := false
	rdr.flags = rdr.super.GetFlags()
	if rdr.flags.compressorOptions {
		switch rdr.super.CompressionType {
		case GzipCompression:
			var gzip *compression.Gzip
			gzip, err = compression.NewGzipCompressorWithOptions(rdr.r)
			if err != nil {
				return nil, err
			}
			if gzip.HasCustomWindow || gzip.HasStrategies {
				hasUnsupportedOptions = true
			}
			rdr.decompressor = gzip
		case XzCompression:
			var xz *compression.Xz
			xz, err = compression.NewXzCompressorWithOptions(rdr.r)
			if err != nil {
				return nil, err
			}
			if xz.HasFilters {
				return nil, errors.New("XZ compression options has filters. These are not yet supported")
			}
			rdr.decompressor = xz
		case Lz4Compression:
			var lz4 *compression.Lz4
			lz4, err = compression.NewLz4CompressorWithOptions(rdr.r)
			if err != nil {
				return nil, err
			}
			rdr.decompressor = lz4
		case ZstdCompression:
			var zstd *compression.Zstd
			zstd, err = compression.NewZstdCompressorWithOptions(rdr.r)
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
	rdr.r.Seek(int64(rdr.super.IDTableStart), io.SeekStart)
	for i := range blockOffsets {
		err = binary.Read(rdr.r, binary.LittleEndian, &blockOffsets[i])
		if err != nil {
			return nil, err
		}
		var idRdr *metadataReader
		idRdr, err = rdr.newMetadataReader(int64(blockOffsets[i]))
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
	metaRdr, err := rdr.newMetadataReaderFromInodeRef(rdr.super.RootInodeRef)
	if err != nil {
		return nil, err
	}
	i, err := inode.ProcessInode(metaRdr, rdr.super.BlockSize)
	if err != nil {
		return nil, err
	}
	entries, err := rdr.readDirFromInode(i)
	if err != nil {
		return nil, err
	}
	rdr.FS = FS{
		r:       &rdr,
		name:    "/",
		entries: entries,
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
