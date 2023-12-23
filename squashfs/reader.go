package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"

	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/toreader"
)

// The types of compression supported by squashfs
const (
	ZlibCompression = uint16(iota + 1)
	LZMACompression
	LZOCompression
	XZCompression
	LZ4Compression
	ZSTDCompression
)

var (
	ErrorMagic         = errors.New("magic incorrect. probably not reading squashfs archive or archive is corrupted")
	ErrorLog           = errors.New("block log is incorrect. possible corrupted archive")
	ErrorVersion       = errors.New("squashfs version of archive is not 4.0. may be corrupted")
	ErrorNotExportable = errors.New("archive does not have an export table")
)

type Reader struct {
	r           io.ReaderAt
	d           decompress.Decompressor
	root        *Directory
	fragTable   []fragEntry
	idTable     []uint32
	exportTable []uint64
	sup         superblock
}

func NewReader(r io.ReaderAt) (rdr *Reader, err error) {
	rdr = new(Reader)
	rdr.r = r
	err = binary.Read(toreader.NewReader(r, 0), binary.LittleEndian, &rdr.sup)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read superblock"), err)
	}
	if !rdr.sup.checkMagic() {
		return nil, ErrorMagic
	}
	if !rdr.sup.checkBlockLog() {
		return nil, ErrorLog
	}
	if !rdr.sup.checkVersion() {
		return nil, ErrorVersion
	}
	switch rdr.sup.CompType {
	case ZlibCompression:
		rdr.d = decompress.Zlib{}
	case LZMACompression:
		rdr.d = decompress.Lzma{}
	case LZOCompression:
		rdr.d = decompress.Lzo{}
	case XZCompression:
		rdr.d = decompress.Xz{}
	case LZ4Compression:
		rdr.d = decompress.Lz4{}
	case ZSTDCompression:
		rdr.d = &decompress.Zstd{}
	default:
		return nil, errors.New("invalid compression type. possible corrupted archive")
	}
	rdr.root, err = rdr.directoryFromRef(rdr.sup.RootInodeRef, "")
	if err != nil {
		return nil, errors.Join(errors.New("failed to read root directory"), err)
	}
	return
}

// Returns the last time the archive was modified.
func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.sup.ModTime), 0)
}

// Get a uid/gid at the given index. Lazily populates the reader's id table as necessary.
func (r *Reader) id(i uint16) (uint32, error) {
	if len(r.idTable) > int(i) {
		return r.idTable[i], nil
	} else if i >= r.sup.IdCount {
		return 0, errors.New("id out of bounds")
	}
	// Populate the id table as needed
	blockNum := uint16(math.Ceil(float64(i) / 2048))
	blocksRead := len(r.idTable) / 2048
	blocksToRead := int(blockNum) - blocksRead

	var offset uint64
	var idsToRead uint16
	var idsTmp []uint32
	var err error
	for i := blocksRead; i < int(blockNum)+blocksToRead; i++ {
		err = binary.Read(toreader.NewReader(r.r, int64(r.sup.IdTableStart)+int64(8*i)), binary.LittleEndian, &offset)
		if err != nil {
			return 0, err
		}
		idsToRead = r.sup.IdCount - uint16(len(r.idTable))
		if idsToRead > 2048 {
			idsToRead = 2048
		}
		idsTmp = make([]uint32, idsToRead)
		err = binary.Read(toreader.NewReader(r.r, int64(offset)), binary.LittleEndian, &idsTmp)
		if err != nil {
			return 0, err
		}
		r.idTable = append(r.idTable, idsTmp...)
	}
	return r.idTable[i], nil
}

// Get a fragment entry at the given index. Lazily populates the reader's fragment table as necessary.
func (r *Reader) fragEntry(i uint32) (fragEntry, error) {
	if len(r.fragTable) > int(i) {
		return r.fragTable[i], nil
	} else if i >= r.sup.FragCount {
		return fragEntry{}, errors.New("fragment out of bounds")
	}
	// Populate the fragment table as needed
	blockNum := uint32(math.Ceil(float64(i) / 512))
	blocksRead := len(r.fragTable) / 512
	blocksToRead := int(blockNum) - blocksRead

	var offset uint64
	var fragsToRead uint32
	var fragsTmp []fragEntry
	var err error
	for i := blocksRead; i < int(blockNum)+blocksToRead; i++ {
		err = binary.Read(toreader.NewReader(r.r, int64(r.sup.FragTableStart)+int64(8*i)), binary.LittleEndian, &offset)
		if err != nil {
			return fragEntry{}, err
		}
		fragsToRead = r.sup.FragCount - uint32(len(r.fragTable))
		if fragsToRead > 512 {
			fragsToRead = 512
		}
		fragsTmp = make([]fragEntry, fragsToRead)
		err = binary.Read(toreader.NewReader(r.r, int64(offset)), binary.LittleEndian, &fragsTmp)
		if err != nil {
			return fragEntry{}, err
		}
		r.fragTable = append(r.fragTable, fragsTmp...)
	}
	return r.fragTable[i], nil
}

// Get an inode reference at the given index. Lazily populates the reader's export table as necessary.
func (r *Reader) inodeRef(i uint32) (uint64, error) {
	if !r.sup.exportable() {
		return 0, ErrorNotExportable
	}
	if len(r.exportTable) > int(i) {
		return r.exportTable[i], nil
	} else if i >= r.sup.InodeCount {
		return 0, errors.New("inode out of bounds")
	}
	// Populate the export table as neede
	blockNum := uint32(math.Ceil(float64(i) / 1024))
	blocksRead := len(r.exportTable) / 1024
	blocksToRead := int(blockNum) - blocksRead

	var offset uint64
	var refsToRead uint32
	var refsTmp []uint64
	var err error
	for i := blocksRead; i < int(blockNum)+blocksToRead; i++ {
		err = binary.Read(toreader.NewReader(r.r, int64(r.sup.ExportTableStart)+int64(8*i)), binary.LittleEndian, &offset)
		if err != nil {
			return 0, err
		}
		refsToRead = r.sup.InodeCount - uint32(len(r.exportTable))
		if refsToRead > 1024 {
			refsToRead = 1024
		}
		refsTmp = make([]uint64, refsToRead)
		err = binary.Read(toreader.NewReader(r.r, int64(offset)), binary.LittleEndian, &refsTmp)
		if err != nil {
			return 0, err
		}
		r.exportTable = append(r.exportTable, refsTmp...)
	}
	return r.exportTable[i], nil
}
