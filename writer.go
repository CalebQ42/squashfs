package squashfs

import (
	"errors"
	"os"
	"path"
)

//Writer is an interface to write a squashfs. Doesn't write until you call Write (TODO: maybe not do Write...)
type Writer struct {
	files           map[string][]*File
	directories     []string
	resolveSymlinks bool
	compression     int
	temp            []*File
}

//NewWriter creates a new squashfs.Writer with the default settings (gzip compression and autoresolving symlinks)
func NewWriter() (*Writer, error) {
	return NewWriterWithOptions(true, GzipCompression)
}

//NewWriterWithOptions creates a new squashfs.Writer with the given options.
//ResolveSymlinks tries to make sure symlinks aren't broken, and if they would be
func NewWriterWithOptions(resolveSymlinks bool, compressionType int) (*Writer, error) {
	if compressionType < 0 || compressionType > 6 || compressionType == 3 {
		return nil, errors.New("Incompatible compression type")
	}
	return &Writer{
		files: map[string][]*File{
			"/": make([]*File, 0),
		},
		directories:     []string{"/"},
		resolveSymlinks: resolveSymlinks,
		compression:     compressionType,
	}, nil
}

func (w *Writer) convertFile(squashfsPath string, file *os.File) error {
	var fil File
	fil.Reader = file
	fil.name = path.Base(file.Name())
	fil.path = squashfsPath
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	defer func() { w.temp = append(w.temp, &fil) }()
	if stat.IsDir() {
		dirs, err := file.Readdirnames(-1)
		if err != nil {
			return err
		}
		for _, dir := range dirs {
			subFil, err := os.Open(file.Name() + dir)
			if err != nil {
				return err
			}
			err = w.convertFile(fil.Path(), subFil)
			if err != nil {
				return err
			}
		}
	}
	//TODO: reg files & symlinks
	return nil
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
