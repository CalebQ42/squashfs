package squashfs

import (
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
)

var (
	//ErrNotDirectory is returned when you're trying to do directory things with a non-directory
	ErrNotDirectory = errors.New("File is not a directory")
)

//File is the main way to interact with files within squashfs, or when putting files into a squashfs.
//File can be either a file or folder. When reading from a squashfs, it reads from the datablocks.
//When writing, this holds the information on WHERE the file will be placed inside the archive.
type File struct {
	Name    string       //The name of the file or folder.
	Parent  *File        //The parent directory. If it's the root directory, will be nil
	Reader  io.Reader    //Underlying reader. When writing, will probably be an os.File. When reading will probably be a FileReader
	path    string       //When writing, you can set where a file goes in the archive with this. (not yet tho)
	size    int          //The size of the file. -1 if a directory
	r       *Reader      //The squashfs.Reader where this file is contained.
	in      *inode.Inode //Underlyting inode when reading.
	filType int          //The file's type, using inode types.
}

func (f *File) GetChildren() ([]*File, error) {
	if !f.IsDir() {
		return nil, ErrNotDirectory
	}
	//TODO
	return nil, nil
}

//IsDir returns if the file is a directory.
func (f *File) IsDir() bool {
	return f.filType == inode.BasicDirectoryType || f.filType == inode.ExtDirType
}

//
func (f *File) Close() {
	//nil the reader to free up resources (in theory). Might switch reader to be a readcloser to make it easier.
	f.Reader = nil
}

//Read from the file. Doesn't do anything fancy, just pases it to the underlying io.Reader. If a directory, return io.EOF
func (f *File) Read(p []byte) (int, error) {
	if f.IsDir() {
		return 0, io.EOF
	}
	//Check if reader is nill and create a new one if needed.
	return f.Reader.Read(p)
}
