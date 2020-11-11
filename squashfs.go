package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	//ErrNotMagical happens when making a new Squashfs and it doesn't have the magic number
	ErrNotMagical = errors.New("Not Magical")
)

//Squashfs is a squashfs backed by a ReadSeeker.
type Squashfs struct {
	rdr                *Reader //underlying reader
	offset             int
	super              Superblock
	flags              SuperblockFlags
	compressionOptions CompressionOptions
}

//NewSquashfs creates a new Squashfs backed by the given reader
func NewSquashfs(reader io.ReaderAt) (*Squashfs, error) {
	rdr := NewReader(reader)
	var superblock Superblock
	err := binary.Read(&rdr, binary.LittleEndian, &superblock)
	if err != nil {
		return nil, err
	}
	if superblock.Magic != squashfsMagic {
		return nil, ErrNotMagical
	}
	flags := superblock.GetFlags()
	var compressionOptions CompressionOptions
	switch superblock.Compression {
	case zlibCompression:
		if flags.CompressorOptions {
			var gzipOpRaw gzipOptionsRaw
			err = binary.Read(&rdr, binary.LittleEndian, &gzipOpRaw)
			if err != nil {
				return nil, err
			}
			compressionOptions = NewGzipOptions(gzipOpRaw)
		} else {
			compressionOptions = NewGzipOptions(gzipOptionsRaw{})
		}
	case xzCompression:
		if flags.CompressorOptions {
			var xzOpRaw xzOptionsRaw
			err = binary.Read(&rdr, binary.LittleEndian, xzOpRaw)
			if err != nil {
				return nil, err
			}
			compressionOptions = NewXzOption(xzOpRaw)
		} else {
			compressionOptions = NewXzOption(xzOptionsRaw{})
		}
	default:
		//TODO: all the compression options
		return nil, errors.New("This type of compression isn't supported yet")
	}
	//TODO: parse more info
	return &Squashfs{
		rdr:                &rdr,
		super:              superblock,
		flags:              flags,
		compressionOptions: compressionOptions,
	}, nil
}

//GetFlags return the SuperblockFlags of the Squashfs
func (s *Squashfs) GetFlags() SuperblockFlags {
	return s.super.GetFlags()
}

//metadata is a parsed metadata block
type metadata struct {
	Compressed bool
	Size       uint16
	Data       io.Reader
}

func (m *metadata) close() {
	switch m.Data.(type) {
	case io.ReadCloser:
		m.Data.(io.ReadCloser).Close()
	}
}

func (s *Squashfs) parseNextMetadata() (*metadata, error) {
	var metaHeader uint16
	err := binary.Read(s.rdr, binary.LittleEndian, metaHeader)
	if err != nil {
		return nil, err
	}
	if metaHeader&0x8000 == 0x8000 {
		metaHeader = metaHeader &^ 0x8000
		compressRead, err := s.compressionOptions.Reader(io.NewSectionReader(s.rdr, s.rdr.offset, int64(metaHeader)))
		return &metadata{
			Compressed: true,
			Size:       metaHeader,
			Data:       *compressRead,
		}, err
	}
	return &metadata{
		Compressed: false,
		Size:       metaHeader,
		Data:       io.NewSectionReader(s.rdr, s.rdr.offset, int64(metaHeader)),
	}, nil
}

func (s *Squashfs) parseMetadataAt(offset int64) (*metadata, error) {
	var metaHeader uint16
	err := binary.Read(s.rdr, binary.LittleEndian, metaHeader)
	if err != nil {
		return nil, err
	}
	if metaHeader&0x8000 == 0x8000 {
		metaHeader = metaHeader &^ 0x8000
		compressRead, err := s.compressionOptions.Reader(io.NewSectionReader(s.rdr, offset, int64(metaHeader)))
		return &metadata{
			Compressed: true,
			Size:       metaHeader,
			Data:       *compressRead,
		}, err
	}
	return &metadata{
		Compressed: false,
		Size:       metaHeader,
		Data:       io.NewSectionReader(s.rdr, offset, int64(metaHeader)),
	}, nil
}
