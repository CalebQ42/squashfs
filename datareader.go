package squashfs

import (
	"bytes"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
)

var (
	//ErrInodeNotFile is given when giving an inode, but the function requires a file inode.
	errInodeNotFile = errors.New("given inode is NOT a file type")
	//ErrInodeOnlyFragment is given when trying to make a DataReader from an inode, but the inode only had data in a fragment
	errInodeOnlyFragment = errors.New("given inode ONLY has fragment data")
)

//DataReader reads data from data blocks.
type dataReader struct {
	r             *Reader
	curData       []byte
	sizes         []uint32
	offset        int64 //offset relative to the beginning of the squash file
	curBlock      int   //Which block in sizes is currently cached
	curReadOffset int   //offset relative to the currently cached data
}

//NewDataReader creates a new data reader at the given offset, with the blocks defined by sizes
func (r *Reader) newDataReader(offset int64, sizes []uint32) (*dataReader, error) {
	var dr dataReader
	dr.r = r
	dr.offset = offset
	dr.sizes = sizes
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
	case inode.FileType:
		fil := i.Info.(inode.File)
		if fil.BlockStart == 0 {
			return nil, errInodeOnlyFragment
		}
		rdr.offset = int64(fil.BlockStart)
		rdr.sizes = append(rdr.sizes, fil.BlockSizes...)
		if fil.Fragmented {
			rdr.sizes = rdr.sizes[:len(rdr.sizes)-1]
		}
	case inode.ExtFileType:
		fil := i.Info.(inode.ExtFile)
		if fil.BlockStart == 0 {
			return nil, errInodeOnlyFragment
		}
		rdr.offset = int64(fil.BlockStart)
		rdr.sizes = append(rdr.sizes, fil.BlockSizes...)
		if fil.Fragmented {
			rdr.sizes = rdr.sizes[:len(rdr.sizes)-1]
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

//removed the compression bit from a data block size
func actualDataSize(size uint32) uint32 {
	return size &^ (1 << 24)
}

func (d *dataReader) readNextBlock() error {
	d.curBlock++
	if d.curBlock >= len(d.sizes) {
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

func (d *dataReader) readBlockAt(offset int64, size uint32) ([]byte, error) {
	compressed := size&(1<<24) != (1 << 24)
	size = size &^ (1 << 24)
	if d.sizes[d.curBlock] == 0 {
		return make([]byte, d.r.super.BlockSize), nil
	}
	sec := io.NewSectionReader(d.r.r, offset, int64(size))
	if compressed {
		btys, err := d.r.decompressor.Decompress(sec)
		if err != nil {
			return nil, err
		}
		return btys, nil
	}
	var buf bytes.Buffer
	_, err := io.Copy(&buf, sec)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *dataReader) offsetForBlock(index int) int64 {
	out := d.offset
	for i := 0; i < index; i++ {
		out += int64(actualDataSize(d.sizes[i]))
	}
	return out
}

func (d *dataReader) readCurBlock() error {
	if d.curBlock >= len(d.sizes) {
		return io.EOF
	}
	offset := d.offsetForBlock(d.curBlock)
	data, err := d.readBlockAt(offset, d.sizes[d.curBlock])
	if err != nil {
		return err
	}
	d.curData = data
	return nil
}

func (d *dataReader) Read(p []byte) (int, error) {
	if d.curData == nil {
		err := d.readCurBlock()
		if err != nil {
			return 0, err
		}
	}
	if d.curReadOffset+len(p) <= len(d.curData) {
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
			if d.curReadOffset < len(d.curData) {
				p[read] = d.curData[d.curReadOffset]
			} else {
				break
			}
			d.curReadOffset++
		}
	}
	if read != len(p) {
		return read, errors.New("didn't read enough data")
	}
	return read, nil
}

// WriteTo writes all the data in the datablock to the writer. MUST BE USED ON A FRESH DATA READER.
func (d *dataReader) WriteTo(w io.Writer) (int64, error) {
	type dataCache struct {
		err   error
		data  []byte
		index int
	}
	dataChan := make(chan *dataCache)
	for i := range d.sizes {
		go func(index int, c chan *dataCache) {
			var cache dataCache
			cache.index = index
			defer func() {
				c <- &cache
			}()
			data, err := d.readBlockAt(d.offsetForBlock(index), d.sizes[index])
			if err != nil {
				cache.err = err
				return
			}
			cache.data = data
		}(i, dataChan)
	}
	curIndex := 0
	totalWrite := int64(0)
	var backlog []*dataCache
mainLoop:
	for {
		if curIndex == len(d.sizes) {
			return totalWrite, nil
		}
		if len(backlog) > 0 {
			for i, cache := range backlog {
				if cache.index == curIndex {
					writen, err := w.Write(cache.data)
					totalWrite += int64(writen)
					if err != nil {
						return totalWrite, err
					}
					if len(backlog) > 0 {
						backlog[i] = backlog[len(backlog)-1]
						backlog = backlog[:len(backlog)-1]
					} else {
						backlog = nil
					}
					curIndex++
					continue mainLoop
				}
			}
		}
		cache := <-dataChan
		if cache.err != nil {
			return totalWrite, cache.err
		}
		if cache.index == curIndex {
			writen, err := w.Write(cache.data)
			totalWrite += int64(writen)
			if err != nil {
				return totalWrite, err
			}
			curIndex++
		} else {
			backlog = append(backlog, cache)
		}
	}
}
