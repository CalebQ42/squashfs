package squashfs

import (
	"bytes"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
)

var (
	//ErrInodeNotFile is given when giving an inode, but the function requires a file inode.
	errInodeNotFile = errors.New("Given inode is NOT a file type")
	//ErrInodeOnlyFragment is given when trying to make a DataReader from an inode, but the inode only had data in a fragment
	errInodeOnlyFragment = errors.New("Given inode ONLY has fragment data")
)

//DataReader reads data from data blocks.
type dataReader struct {
	r             *Reader
	offset        int64 //offset relative to the beginning of the squash file
	blocks        []dataBlock
	curBlock      int //Which block in sizes is currently cached
	curData       []byte
	curReadOffset int //offset relative to the currently cached data
}

//DataBlock holds info about a given data block from it's size
type dataBlock struct {
	begOffset        int64 //The offset relative to the beginning of the squash file. Makes it easier to seek to it.
	size             uint32
	compressed       bool
	uncompressedSize uint32
}

//NewDataBlock creates a new squashfs.datablock from a given size.
func newDataBlock(raw uint32) (dbs dataBlock) {
	dbs.compressed = raw&(1<<24) != (1 << 24)
	dbs.size = raw &^ (1 << 24)
	if !dbs.compressed {
		dbs.uncompressedSize = dbs.size
	}
	return
}

//NewDataReader creates a new data reader at the given offset, with the blocks defined by sizes
func (r *Reader) newDataReader(offset int64, sizes []uint32) (*dataReader, error) {
	var dr dataReader
	dr.r = r
	dr.offset = offset
	for _, size := range sizes {
		dr.blocks = append(dr.blocks, newDataBlock(size))
	}
	err := dr.readCurBlock()
	if err != nil {
		return nil, err
	}
	return &dr, nil
}

//NewDataReaderFromInode creates a new DataReader from a given inode. Inode must be of BasicFile or ExtendedFile types
func (r *Reader) newDataReaderFromInode(i *inode.Inode) (*dataReader, error) {
	var rdr dataReader
	rdr.r = r
	switch i.Type {
	case inode.BasicFileType:
		fil := i.Info.(inode.BasicFile)
		if fil.Init.BlockStart == 0 {
			return nil, errInodeOnlyFragment
		}
		rdr.offset = int64(fil.Init.BlockStart)
		for _, sizes := range fil.BlockSizes {
			rdr.blocks = append(rdr.blocks, newDataBlock(sizes))
		}
		if fil.Fragmented {
			rdr.blocks = rdr.blocks[:len(rdr.blocks)-1]
		}
	case inode.ExtFileType:
		fil := i.Info.(inode.ExtendedFile)
		if fil.Init.BlockStart == 0 {
			return nil, errInodeOnlyFragment
		}
		rdr.offset = int64(fil.Init.BlockStart)
		for _, sizes := range fil.BlockSizes {
			rdr.blocks = append(rdr.blocks, newDataBlock(sizes))
		}
		if fil.Fragmented {
			rdr.blocks = rdr.blocks[:len(rdr.blocks)-1]
		}
	default:
		return nil, errInodeNotFile
	}
	err := rdr.readCurBlock()
	if err != nil {
		return nil, err
	}
	return &rdr, nil
}

func (d *dataReader) readNextBlock() error {
	d.curBlock++
	if d.curBlock >= len(d.blocks) {
		d.curBlock--
		return io.EOF
	}
	err := d.readCurBlock()
	if err != nil {
		d.curBlock--
		d.readCurBlock()
		return err
	}
	return nil
}

func (d *dataReader) readCurBlock() error {
	if d.curBlock >= len(d.blocks) {
		return io.EOF
	}
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

//Close frees up the curData from memory
func (d *dataReader) Close() error {
	d.curData = nil
	return nil
}

func (d *dataReader) Read(p []byte) (int, error) {
	if d.curData == nil {
		d.readCurBlock()
	}
	if d.curReadOffset+len(p) < len(d.curData) {
		for i := 0; i < len(p); i++ {
			p[i] = d.curData[d.curReadOffset+i]
		}
		d.curReadOffset += len(p)
		return len(p), nil
	}
	read := 0
	for read < len(p) {
		if d.curReadOffset == len(d.curData) {
			err := d.readNextBlock()
			if err != nil {
				return read, err
			}
			d.curReadOffset = 0
		}
		for ; read < len(p); read++ {
			d.curReadOffset++
			if d.curReadOffset < len(d.curData) {
				p[read] = d.curData[d.curReadOffset]
			} else {
				break
			}
		}
	}
	if read != len(p) {
		return read, errors.New("Didn't read enough data")
	}
	return read, nil
}
