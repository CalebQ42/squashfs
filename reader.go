package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"

	"github.com/CalebQ42/squashfs/internal/compression"
	"github.com/CalebQ42/squashfs/internal/inode"
	"github.com/CalebQ42/squashfs/internal/rawreader"
)

const (
	magic uint32 = 0x73717368
)

var (
	//ErrNoMagic is returned if the magic number in the superblock isn't correct.
	errNoMagic = errors.New("magic number doesn't match. Either isn't a squashfs or corrupted")
	//ErrIncompatibleCompression is returned if the compression type in the superblock doesn't work.
	errIncompatibleCompression = errors.New("compression type unsupported")
)

//Reader processes and reads a squashfs archive.
type Reader struct {
	FS
	r            rawreader.RawReader
	decompressor compression.Decompressor
	fragOffsets  []uint64
	idTable      []uint32
	super        superblock
	flags        SuperblockFlags
}

//NewSquashfsReader returns a new squashfs.Reader from an io.ReaderAt
func NewSquashfsReader(r io.ReaderAt) (*Reader, error) {
	var rdr Reader
	rdr.r = rawreader.ConvertReaderAt(r)
	err := rdr.Init()
	if err != nil {
		return nil, err
	}
	return &rdr, nil
}

//NewSquashfsReaderFromReader returns a new squashfs.Reader from an io.Reader.
//If the io.Reader implements io.Seeker, the seek functions are used.
//It is NOT recommended to use a pure io.Reader as due to how squashfs
//archives are formatted, the ENTIRETY of the io.Reader's data is loaded into
//memory first before it can be used.
func NewSquashfsReaderFromReader(r io.Reader) (*Reader, error) {
	var rdr Reader
	var err error
	rdr.r, err = rawreader.ConvertReader(r)
	if err != nil {
		return nil, err
	}
	err = rdr.Init()
	if err != nil {
		return nil, err
	}
	return &rdr, nil
}

func (r *Reader) Init() error {
	err := binary.Read(r.r, binary.LittleEndian, &r.super)
	if err != nil {
		return err
	}
	if r.super.Magic != magic {
		return errNoMagic
	}
	if r.super.BlockLog != uint16(math.Log2(float64(r.super.BlockSize))) {
		return errors.New("BlockSize and BlockLog doesn't match. The archive is probably corrupt")
	}
	r.r.Seek(96, io.SeekStart)
	r.flags = r.super.GetFlags()
	if r.flags.compressorOptions {
		switch r.super.CompressionType {
		case GzipCompression:
			var gzip *compression.Gzip
			gzip, err = compression.NewGzipCompressorWithOptions(r.r)
			if err != nil {
				return err
			}
			r.decompressor = gzip
		case XzCompression:
			var xz *compression.Xz
			xz, err = compression.NewXzCompressorWithOptions(r.r)
			if err != nil {
				return err
			}
			r.decompressor = xz
		case LzoCompression:
			var lz *compression.Lzo
			lz, err = compression.NewLzoCompressorWithOptions(r.r)
			if err != nil {
				return err
			}
			r.decompressor = lz
		case Lz4Compression:
			var lz4 *compression.Lz4
			lz4, err = compression.NewLz4CompressorWithOptions(r.r)
			if err != nil {
				return err
			}
			r.decompressor = lz4
		case ZstdCompression:
			var zstd *compression.Zstd
			zstd, err = compression.NewZstdCompressorWithOptions(r.r)
			if err != nil {
				return err
			}
			r.decompressor = zstd
		default:
			return errIncompatibleCompression
		}
	} else {
		switch r.super.CompressionType {
		case GzipCompression:
			r.decompressor = &compression.Gzip{}
		case LzmaCompression:
			r.decompressor = &compression.Lzma{}
		case LzoCompression:
			r.decompressor = &compression.Lzo{}
		case XzCompression:
			r.decompressor = &compression.Xz{}
		case Lz4Compression:
			r.decompressor = &compression.Lz4{}
		case ZstdCompression:
			r.decompressor = &compression.Zstd{}
		default:
			//TODO: all compression types.
			return errIncompatibleCompression
		}
	}
	fragBlocks := int(math.Ceil(float64(r.super.FragCount) / 512))
	if fragBlocks > 0 {
		offset := int64(r.super.FragTableStart)
		for i := 0; i < fragBlocks; i++ {
			tmp := make([]byte, 8)
			_, err = r.r.ReadAt(tmp, offset)
			if err != nil {
				return err
			}
			r.fragOffsets = append(r.fragOffsets, binary.LittleEndian.Uint64(tmp))
			offset += 8
		}
	}
	unread := r.super.IDCount
	blockOffsets := make([]uint64, int(math.Ceil(float64(r.super.IDCount)/2048)))
	_, err = r.r.Seek(int64(r.super.IDTableStart), io.SeekStart)
	if err != nil {
		return err
	}
	for i := range blockOffsets {
		err = binary.Read(r.r, binary.LittleEndian, &blockOffsets[i])
		if err != nil {
			return err
		}
		var idRdr *metadataReader
		idRdr, err = r.newMetadataReader(int64(blockOffsets[i]))
		if err != nil {
			return err
		}
		read := uint16(math.Min(float64(unread), 2048))
		for i := uint16(0); i < read; i++ {
			var tmp uint32
			err = binary.Read(idRdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
			r.idTable = append(r.idTable, tmp)
		}
		unread -= read
	}
	metaRdr, err := r.newMetadataReaderFromInodeRef(r.super.RootInodeRef)
	if err != nil {
		return err
	}
	i, err := inode.ProcessInode(metaRdr, r.super.BlockSize)
	if err != nil {
		return err
	}
	entries, err := r.readDirFromInode(i)
	if err != nil {
		return err
	}
	r.FS = FS{
		r:       r,
		name:    "/",
		entries: entries,
	}
	return nil
}

//ModTime is the last time the file was modified/created.
func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.super.CreationTime), 0)
}
