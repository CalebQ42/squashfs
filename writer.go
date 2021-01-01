package squashfs

import (
	"errors"
	"os"

	"github.com/CalebQ42/squashfs/internal/inode"
)

//Writer is an interface to write a squashfs. Doesn't write until you call Write (TODO: maybe not do Write...)
type Writer struct {
	files       map[string]*File
	directories []string
}

func convertFile(file *os.File) (*File, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	var filType int
	if stat.IsDir() {
		//TODO
	} else if stat.Mode().IsRegular() {
		filType = inode.BasicFileType
		return &File{
			Reader:  file,
			name:    file.Name(),
			filType: filType,
		}, nil
	} else if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		//TODO: implement symlink support
		return nil, errors.New("No symlink support. No support at all actually...")
	}
	return nil, errors.New("File type is NOT supported")
}

//AddFilesToPath adds the give os.Files to the given path within the squashfs archive.
func (w *Writer) AddFilesToPath(squashfsPath string, files ...*os.File) error {
	//TODO
	return errors.New("Don't")
}

//AddFiles adds all files given to the root directory
func (w *Writer) AddFiles(files ...*os.File) error {
	//TODO
	return errors.New("Don't")
}

//RemoveFileAt removes the file at filepath from the Writer.
//If multiple files match the given filepath (such as if there are wildcards), all matching files are removed.
//If one or more files are removed, returns true.
func (w *Writer) RemoveFileAt(filepath string) bool {
	//TODO
	return false
}
