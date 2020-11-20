package squashfs

import (
	"bytes"
	"errors"
	"io"
)

type DataReader struct {
	r             *Reader
	offset        int64 //offset relative to the beginning of the squash file
	blocks        []DataBlock
	curBlock      int //Which block in sizes is currently cached
	curData       []byte
	curReadOffset int //offset relative to the currently cached data
	readOffset    int //offset relative to the beginning
}

type DataBlock struct {
	begOffset        int64 //The offset relative to the beginning of the squash file. Makes it easier to seek to it.
	size             uint32
	compressed       bool
	uncompressedSize uint32
}

func NewDataBlockSize(raw uint32) (dbs DataBlock) {
	dbs.compressed = raw&1<<24 != 1<<24
	dbs.size = raw &^ 1 << 24
	if !dbs.compressed {
		dbs.uncompressedSize = dbs.size
	}
	return
}

func (r *Reader) NewDataReader(offset int64, sizes []uint32) (*DataReader, error) {
	var dr DataReader
	dr.r = r
	dr.offset = offset
	for _, size := range sizes {
		dr.blocks = append(dr.blocks, NewDataBlockSize(size))
	}
	err := dr.readCurBlock()
	if err != nil {
		return nil, err
	}
	return &dr, nil
}

func (d *DataReader) readNextBlock() error {
	d.curBlock++
	if d.curBlock >= len(d.blocks) {
		d.curBlock--
		return errors.New("Ran out of blocks")
	}
	err := d.readCurBlock()
	if err != nil {
		d.curBlock--
		d.readCurBlock()
		return err
	}
	return nil
}

func (d *DataReader) readCurBlock() error {
	if d.blocks[d.curBlock].size == 0 {
		d.curData = make([]byte, d.r.super.BlockSize)
		d.blocks[d.curBlock].uncompressedSize = d.r.super.BlockSize
		d.blocks[d.curBlock].begOffset = d.offset
		return nil
	}
	sec := io.NewSectionReader(d.r.r, d.offset, int64(d.blocks[d.curBlock].size))
	if d.blocks[d.curBlock].compressed {
		btys, err := d.r.decompressor.Decompress(sec)
		if err != nil {
			return err
		}
		d.blocks[d.curBlock].uncompressedSize = uint32(len(btys))
		d.curData = btys
		d.blocks[d.curBlock].begOffset = d.offset
		d.offset += int64(d.blocks[d.curBlock].size)
		return nil
	}
	var buf bytes.Buffer
	_, err := io.Copy(&buf, sec)
	if err != nil {
		return err
	}
	d.curData = buf.Bytes()
	d.blocks[d.curBlock].begOffset = d.offset
	d.offset += int64(d.blocks[d.curBlock].size)
	return err
}

func (d *DataReader) Read(p []byte) (int, error) {
	if d.curReadOffset+len(p) < len(d.curData) {
		for i := 0; i < len(p); i++ {
			p[i] = d.curData[d.curReadOffset+i]
		}
		d.readOffset += len(p)
		d.curReadOffset += len(p)
		return len(p), nil
	}
	read := 0
	curRead := 0
	for read < len(p) {
		if d.curReadOffset == len(d.curData) {
			err := d.readNextBlock()
			if err != nil {
				d.readOffset += read
				return read, err
			}
			curRead = 0
		}
		for ; read < len(p); read++ {
			curRead++
			if d.curReadOffset+curRead < len(d.curData) {
				p[read] = d.curData[d.curReadOffset+curRead]
			} else {
				break
			}
		}
	}
	d.readOffset += read
	if read != len(p) {
		return read, errors.New("Didn't read enough data")
	}
	return read, nil
}

func (d *DataReader) Seek(offset int64, whence int) (int, error) {
	if whence == io.SeekStart {
		d.readOffset = int(offset)
		d.curReadOffset = int(offset)
		d.curBlock = 0
		read := int64(0)
		for i, block := range d.blocks {
			if block.uncompressedSize == 0 {
				d.curBlock = i
				err := d.readCurBlock()
				if err != nil {
					return 0, err
				}
			}
			read += int64(block.uncompressedSize)
			if offset-read < 0 {
				d.curReadOffset = int((offset - read) + int64(block.uncompressedSize))
			}
		}
	}
	switch whence {
	case io.SeekCurrent:
		d.readOffset += int(offset)
		d.curReadOffset += int(offset)
	case io.SeekEnd:
	}
}
