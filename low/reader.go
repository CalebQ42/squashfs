package squashfslow

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/toreader"
	"github.com/CalebQ42/squashfs/low/inode"
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
	Root        Directory
	Superblock  superblock
	r           io.ReaderAt
	d           decompress.Decompressor
	fragTable   *Table[fragEntry]
	idTable     *Table[uint32]
	exportTable *Table[InodeRef]
}

func NewReader(r io.ReaderAt) (rdr Reader, err error) {
	rdr.r = r
	err = binary.Read(toreader.NewReader(r, 0), binary.LittleEndian, &rdr.Superblock)
	if err != nil {
		return rdr, errors.Join(errors.New("failed to read superblock"), err)
	}
	if !rdr.Superblock.ValidMagic() {
		return rdr, ErrorMagic
	}
	if !rdr.Superblock.ValidBlockLog() {
		return rdr, ErrorLog
	}
	if !rdr.Superblock.ValidVersion() {
		return rdr, ErrorVersion
	}
	switch rdr.Superblock.CompType {
	case ZlibCompression:
		rdr.d = decompress.NewZlib()
	case LZMACompression:
		rdr.d, err = decompress.NewLzma()
		if err != nil {
			return rdr, err
		}
	case LZOCompression:
		rdr.d, err = decompress.NewLzo()
		if err != nil {
			return rdr, err
		}
	case XZCompression:
		rdr.d = decompress.NewXz()
	case LZ4Compression:
		rdr.d = decompress.NewLz4()
	case ZSTDCompression:
		rdr.d = decompress.NewZstd()
	default:
		return rdr, errors.New("invalid compression type. possible corrupted archive")
	}
	rdr.Root, err = rdr.directoryFromRef(rdr.Superblock.RootInodeRef, "")
	if err != nil {
		return rdr, errors.Join(errors.New("failed to read root directory"), err)
	}
	rdr.fragTable = NewTable(&rdr, rdr.Superblock.FragTableStart, rdr.Superblock.FragCount, readFrag)
	rdr.idTable = NewTable(&rdr, rdr.Superblock.IdTableStart, uint32(rdr.Superblock.IdCount), readId)
	rdr.exportTable = NewTable(&rdr, rdr.Superblock.ExportTableStart, rdr.Superblock.InodeCount, readRef)
	return
}

func readFrag(r io.Reader) (fragEntry, error) {
	dat := make([]byte, 16)
	_, err := r.Read(dat)
	if err != nil {
		return fragEntry{}, err
	}
	return fragEntry{
		Start: binary.LittleEndian.Uint64(dat[0:8]),
		Size:  binary.LittleEndian.Uint32(dat[8:12]),
	}, nil
}

func readId(r io.Reader) (uint32, error) {
	dat := make([]byte, 4)
	_, err := r.Read(dat)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(dat), nil
}

func readRef(r io.Reader) (InodeRef, error) {
	dat := make([]byte, 8)
	_, err := r.Read(dat)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(dat), nil
}

// Get a uid/gid at the given index. Lazily populates the reader's Id table as necessary.
func (r *Reader) Id(i uint16) (uint32, error) {
	return r.idTable.Get(uint32(i))
}

// Get a fragment entry at the given index. Lazily populates the reader's fragment table as necessary.
func (r *Reader) fragEntry(i uint32) (fragEntry, error) {
	return r.fragTable.Get(i)
}

// Get an inode reference at the given index. Lazily populates the reader's export table as necessary.
func (r *Reader) inodeRef(i uint32) (InodeRef, error) {
	return r.exportTable.Get(i)
}

func (r Reader) Inode(i uint32) (inode.Inode, error) {
	ref, err := r.inodeRef(i - 1) // Inode table is 1 indexed
	if err != nil {
		return inode.Inode{}, err
	}
	return r.InodeFromRef(ref)
}
