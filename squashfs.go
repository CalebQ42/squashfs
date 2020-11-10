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
	rdr                *io.SectionReader //underlying reader
	offset             int
	super              Superblock
	flags              SuperblockFlags
	compressionOptions CompressionOptions
}

//NewSquashfs creates a new Squashfs backed by the given reader
func NewSquashfs(reader *io.SectionReader) (*Squashfs, error) {
	var superblock Superblock
	err := binary.Read(reader, binary.LittleEndian, &superblock)
	if err != nil {
		return nil, err
	}
	if superblock.Magic != squashfsMagic {
		return nil, ErrNotMagical
	}
	flags := superblock.GetFlags()
	var compressionOptions CompressionOptions
	if flags.CompressorOptions {
		switch superblock.Compression {
		case zlibCompression:
			var gzipOpRaw gzipOptionsRaw
			err = binary.Read(reader, binary.LittleEndian, &gzipOpRaw)
			if err != nil {
				return nil, err
			}
			compressionOptions = NewGzipOptions(gzipOpRaw)
			break
		default:
			//TODO: all the compression options
			return nil, errors.New("This type of compression isn't supported yet")
		}
	}
	//TODO: parse more info
	return &Squashfs{
		rdr:                reader,
		super:              superblock,
		flags:              flags,
		compressionOptions: compressionOptions,
	}, nil
}

//GetFlags return the SuperblockFlags of the Squashfs
func (s *Squashfs) GetFlags() SuperblockFlags {
	return s.super.GetFlags()
}

//Metadata is a parsed metadata block
type Metadata struct {
	Compressed bool
	Size       uint16
	Data       *io.SectionReader
}

func (s *Squashfs) parseNextMetadata() (*Metadata, error) {
	var metaHeader uint16
	err := binary.Read(s.rdr, binary.LittleEndian, metaHeader)
	if err != nil {
		return nil, err
	}
	if metaHeader&0x8000 == 0x8000 {
		metaHeader = metaHeader &^ 0x8000
		//TODO: read compressed metadata
		return nil, errors.New("Metadata is compressed, which is not implemented yet")
	}
	return &Metadata{
		Compressed: false,
		Size:       metaHeader,
		//TODO: Data:       io.NewSectionReader(s.rdr, , metaHeader),
	}, nil
}
