package squashfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/inode"
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
	case gzipCompression:
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

func (s *Squashfs) readRootDirectoryTable() error {
	offset, metaOffset := inode.ProcessInodeRef(s.super.RootInode)
	meta, err := s.parseMetadataAt(int64(s.super.InodeTableOffset) + int64(offset))
	if err != nil {
		fmt.Println("Error processing metadata")
		return err
	}
	_, err = meta.Data.Read(make([]byte, metaOffset))
	if err != nil {
		fmt.Println("error reading forward to offset")
		return err
	}
	common, _, err := inode.ProcessInode(&meta.Data, s.super.BlockSize)
	if err != nil {
		fmt.Println("Error reading inode")
		return err
	}
	if common.InodeType != inode.BasicDirectoryType {
		return errors.New("Not a basic directory")
	}
	// dirTable, err := directory.NewDirectory(meta.Data)
	// if err != nil {
	// 	return err
	// }
	// for _, entry := range dirTable.Entries {
	// 	fmt.Println(entry.Name)
	// }
	return nil
}

//GetFlags return the SuperblockFlags of the Squashfs
func (s *Squashfs) GetFlags() SuperblockFlags {
	return s.flags
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
	err := binary.Read(s.rdr, binary.LittleEndian, &metaHeader)
	if err != nil {
		return nil, err
	}
	reader := io.NewSectionReader(s.rdr, s.rdr.offset, s.rdr.offset+int64(metaHeader))
	if metaHeader&0x8000 == 0x8000 {
		metaHeader = metaHeader &^ 0x8000
		compressRead, err := s.compressionOptions.Reader(reader)
		return &metadata{
			Compressed: true,
			Size:       metaHeader,
			Data:       *compressRead,
		}, err
	}
	return &metadata{
		Compressed: false,
		Size:       metaHeader,
		Data:       reader,
	}, nil
}

func (s *Squashfs) parseMetadataAt(offset int64) (*metadata, error) {
	var metaHeader uint16
	var headerBytes []byte
	headerBytes = make([]byte, 2)
	s.rdr.ReadAt(headerBytes, offset)
	metaHeader = binary.LittleEndian.Uint16(headerBytes)
	if metaHeader&0x8000 == 0x8000 {
		metaHeader = metaHeader &^ 0x8000
		compressRead, err := s.compressionOptions.Reader(io.NewSectionReader(s.rdr, s.rdr.offset, s.rdr.offset+int64(s.super.BlockSize)))
		return &metadata{
			Compressed: true,
			Size:       metaHeader,
			Data:       *compressRead,
		}, err
	}
	return &metadata{
		Compressed: false,
		Size:       metaHeader,
		Data:       io.NewSectionReader(s.rdr, s.rdr.offset, s.rdr.offset+int64(s.super.BlockSize)),
	}, nil
}
