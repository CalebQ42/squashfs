package squashfs

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/CalebQ42/GoSquashfs/bytereadwrite"
)

type MetadataHeader struct {
	rawHeader  uint16
	Compressed bool
	Size       uint16
}

type BlockReader struct {
	initalOffset int64
	offset       int64
	squash       *Squashfs
	headers      []MetadataHeader
	dataCache    []byte
	dataOffset   int64
}

//NewBlockReader creates a new BlockReader from a squashfs.Reader. Reads the first header and caches the first set of data.
func (s *Squashfs) NewBlockReader(offset int64) (*BlockReader, error) {
	var br BlockReader
	br.squash = s
	br.initalOffset = offset
	br.offset = offset
	br.headers = make([]MetadataHeader, 0)
	br.dataCache = make([]byte, 0)
	err := br.parseNewBlock()
	if err != nil {
		fmt.Println("Problem creating BlockReader")
		return nil, err
	}
	return &br, nil
}

func (br *BlockReader) parseNewBlock() error {
	var header MetadataHeader
	err := binary.Read(io.NewSectionReader(&br.squash.r, br.offset, 2), binary.LittleEndian, &header.rawHeader)
	if err != nil {
		fmt.Println("Error while reading the header ", len(br.headers), " in BlockReader")
		return err
	}
	header.Compressed = (header.rawHeader&0x8000 == 0x8000)
	header.Size = header.rawHeader &^ 0x8000
	br.headers = append(br.headers, header)
	br.offset += 2
	sectionReader := io.NewSectionReader(&br.squash.r, br.offset, br.offset+int64(header.Size))
	dataWriter := bytereadwrite.NewByteReaderWriter()
	_, err = io.Copy(dataWriter, sectionReader)
	br.offset += int64(header.Size)
	if header.Compressed {
		data, err := br.squash.compression.Decompress(dataWriter)
		if err != nil {
			fmt.Println("Error while reading the encrypted data block in header", len(br.headers))
			return err
		}
		br.dataCache = append(br.dataCache, data...)
		return nil
	}
	if err != nil {
		fmt.Println("Error while reading uncompressed data in header", len(br.headers))
		return err
	}
	br.dataCache = append(br.dataCache, dataWriter.GetBytes()...)
	return nil
}

//Read reads data into p. If it reaches EOF, tries to read a new block and add it's data to the cache
func (br *BlockReader) Read(p []byte) (n int, err error) {
	read := 0
	for {
		byter := bytereadwrite.NewByteReaderWriterFromBytes(br.dataCache[br.dataOffset:])
		temp, err := byter.Read(p[read:])
		br.dataOffset += int64(temp)
		read += temp
		if err == nil {
			return read, nil
		} else if err != io.EOF {
			return read, err
		}
		if err == io.EOF {
			err = br.parseNewBlock()
			if err != nil {
				fmt.Println("Error while reading a new block")
				return read, err
			}
		}
	}
}

//ReadAt reads data into p from the offset. If it reaches EOF, tries to read a new block and add it's data to the cache.
//
//Offset is reletive to the offset set on creation.
func (br *BlockReader) ReadAt(p []byte, offset int) (n int, err error) {
	read := 0
	for err == io.EOF {
		byter := bytereadwrite.NewByteReaderWriterFromBytes(br.dataCache[offset+read:])
		temp, inErr := byter.Read(p[read:])
		err = inErr
		read += temp
		br.dataOffset += int64(temp)
		if err == io.EOF {
			inErr = br.parseNewBlock()
			if inErr != nil {
				fmt.Println("Error while reading a new block")
				return read, inErr
			}
		}
	}
	return
}
