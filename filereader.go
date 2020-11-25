package squashfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

//FileReader provides a io.Reader interface for files within
type FileReader struct {
	r            *Reader
	data         *DataReader
	fragmentData []byte
	fragged      bool
	fragOnly     bool
	read         int
	FileSize     int //FileSize is the total size of the given file
}

var (
	//ErrPathIsNotFile returns when trying to read from a file, but the given path is NOT a file.
	ErrPathIsNotFile = errors.New("The given path is not a file")
)

//ReadFile provides a squashfs.FileReader for the file at the given location.
func (r *Reader) ReadFile(location string) (*FileReader, error) {
	var rdr FileReader
	rdr.r = r
	in, err := r.GetInodeFromPath(location)
	if err != nil {
		return nil, err
	}
	if in.Type != inode.BasicFileType && in.Type != inode.ExtFileType {
		return nil, ErrPathIsNotFile
	}
	var offset uint32
	var sizes []uint32
	switch in.Type {
	case inode.BasicFileType:
		rdr.fragged = in.Info.(inode.BasicFile).Fragmented
		rdr.fragOnly = in.Info.(inode.BasicFile).Init.BlockStart == 0
		rdr.FileSize = int(in.Info.(inode.BasicFile).Init.Size)
		offset = in.Info.(inode.BasicFile).Init.BlockStart
		sizes = in.Info.(inode.BasicFile).BlockSizes
	case inode.ExtFileType:
		rdr.fragged = in.Info.(inode.ExtendedFile).Fragmented
		rdr.fragOnly = in.Info.(inode.ExtendedFile).Init.BlockStart == 0
		rdr.FileSize = int(in.Info.(inode.ExtendedFile).Init.Size)
		offset = in.Info.(inode.ExtendedFile).Init.BlockStart
		sizes = in.Info.(inode.ExtendedFile).BlockSizes
	}
	fmt.Println("HIIII")
	if rdr.fragged {
		rdr.fragmentData, err = r.GetFragmentDataFromInode(in)
		if err != nil {
			return nil, err
		}
	}
	if rdr.fragged {
		rdr.data, err = r.NewDataReader(int64(offset), sizes[:len(sizes)-1])
	} else {
		rdr.data, err = r.NewDataReader(int64(offset), sizes)
	}
	return &rdr, nil
}

func (f *FileReader) Read(p []byte) (int, error) {
	fmt.Println("reading!")
	var read int
	n, err := f.data.Read(p)
	read += n
	if f.fragged && err == io.EOF {
		n, err = bytes.NewBuffer(f.fragmentData).Read(p[read:])
		read += n
		if err != nil {
			return read, err
		}
	} else if err != nil {
		return read, err
	}
	return read, nil
}
