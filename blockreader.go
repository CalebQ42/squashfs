package squashfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type metadata struct {
	raw        uint16
	size       uint16
	compressed bool
}

type BlockReader struct {
	s          *Reader
	offset     int64
	headers    []*metadata
	data       []byte
	readOffset int
}

func (s *Reader) NewBlockReader(offset int64) (*BlockReader, error) {
	var br BlockReader
	br.s = s
	br.offset = offset
	err := br.parseMetadata()
	if err != nil {
		return nil, err
	}
	err = br.readNextDataBlock()
	if err != nil {
		return nil, err
	}
	return &br, nil
}

func (br *BlockReader) parseMetadata() error {
	var raw uint16
	err := binary.Read(io.NewSectionReader(br.s.r, br.offset, 2), binary.LittleEndian, &raw)
	if err != nil {
		return err
	}
	br.offset += 2
	compressed := !(raw&0x8000 == 0x8000)
	size := raw &^ 0x8000
	br.headers = append(br.headers, &metadata{
		raw:        raw,
		size:       size,
		compressed: compressed,
	})
	return nil
}

func (br *BlockReader) readNextDataBlock() error {
	meta := br.headers[len(br.headers)-1]
	r := io.NewSectionReader(br.s.r, br.offset, int64(meta.size))
	if meta.compressed {
		byts, err := br.s.decompressor.Decompress(r)
		if err != nil {
			return err
		}
		br.offset += int64(meta.size)
		br.data = append(br.data, byts...)
		return nil
	}
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return err
	}
	br.offset += int64(meta.size)
	br.data = append(br.data, buf.Bytes()...)
	return nil
}

func (br *BlockReader) Read(p []byte) (int, error) {
	fmt.Println("reading", len(p))
	if br.readOffset+len(p) < len(br.data) {
		for i := 0; i < len(p); i++ {
			p[i] = br.data[br.readOffset+i]
		}
		br.readOffset += len(p)
		fmt.Println("enough data available")
		return len(p), nil
	}
	read := 0
	for read < len(p) {
		fmt.Println("Reading new block")
		err := br.parseMetadata()
		if err != nil {
			br.readOffset += read
			return read, err
		}
		err = br.readNextDataBlock()
		if err != nil {
			br.readOffset += read
			return read, err
		}
		for ; read < len(p); read++ {
			// fmt.Println("Reading...")
			if br.readOffset+read < len(br.data) {
				p[read] = br.data[br.readOffset+read]
			} else {
				break
			}
		}
	}
	br.readOffset += read
	if read != len(p) {
		return read, errors.New("Didn't read enough data")
	}
	return read, nil
}

//Seek will seek to the specified location (if possible).
//When io.SeekCurrent or io.SeekStart is set, if seeking would put the offset beyond the current cached data, it will try to read the next data blocks to accomodate. On a failure it will seek to the end of the data.
//When io.SeekEnd is set, it wil seek reletive to the currently cached data.
func (br *BlockReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		br.readOffset += int(offset)
		for {
			if br.readOffset < len(br.data) {
				break
			}
			err := br.parseMetadata()
			if err != nil {
				br.readOffset = len(br.data)
				return int64(br.readOffset), err
			}
			err = br.readNextDataBlock()
			if err != nil {
				br.readOffset = len(br.data)
				return int64(br.readOffset), err
			}
		}
	case io.SeekStart:
		br.readOffset = int(offset)
		for {
			if br.readOffset < len(br.data) {
				break
			}
			err := br.parseMetadata()
			if err != nil {
				br.readOffset = len(br.data)
				return int64(br.readOffset), err
			}
			err = br.readNextDataBlock()
			if err != nil {
				br.readOffset = len(br.data)
				return int64(br.readOffset), err
			}
		}
	case io.SeekEnd:
		br.readOffset = len(br.data) - int(offset)
		if br.readOffset < 0 {
			br.readOffset = 0
			return int64(br.readOffset), errors.New("Trying to seek to a negative value")
		}
	}
	return int64(br.readOffset), nil
}
