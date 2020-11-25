package squashfs

import (
	"bytes"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
)

//FileReader provides a io.Reader interface for files within a squashfs archive
type FileReader struct {
	r            *Reader
	data         *dataReader
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
	in, err := r.getInodeFromPath(location)
	if err != nil {
		return nil, err
	}
	if in.Type != inode.BasicFileType && in.Type != inode.ExtFileType {
		return nil, ErrPathIsNotFile
	}
	switch in.Type {
	case inode.BasicFileType:
		fil := in.Info.(inode.BasicFile)
		rdr.fragged = fil.Fragmented
		rdr.fragOnly = fil.Init.BlockStart == 0
		rdr.FileSize = int(fil.Init.Size)
	case inode.ExtFileType:
		fil := in.Info.(inode.ExtendedFile)
		rdr.fragged = fil.Fragmented
		rdr.fragOnly = fil.Init.BlockStart == 0
		rdr.FileSize = int(fil.Init.Size)
	}
	if rdr.fragged {
		rdr.fragmentData, err = r.getFragmentDataFromInode(in)
		if err != nil {
			return nil, err
		}
	}
	if !rdr.fragOnly {
		rdr.data, err = r.newDataReaderFromInode(in)
	}
	return &rdr, nil
}

func (f *FileReader) Read(p []byte) (int, error) {
	if f.fragOnly {
		n, err := bytes.NewBuffer(f.fragmentData[f.read:]).Read(p)
		f.read += n
		if err != nil {
			return n, err
		}
		return n, nil
	}
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
