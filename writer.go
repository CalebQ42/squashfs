package squashfs

import (
	"errors"
	"log"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/squashfs/internal/inode"
)

//Writer is an interface to write a squashfs. Doesn't write until you call Write (TODO: maybe not do Write...).
//If AllowErrors is true, when errors are encountered, it just prints to the log instead of failing.
type Writer struct {
	files           map[string][]*File
	directories     []string
	symlinkTable    map[string]string //symlinkTable holds info about symlink'd to files that had to be moved from their original position. [originalpath]newpath
	symTableTemp    map[string]string
	ResolveSymlinks bool
	AllowErrors     bool
	compression     int
}

//NewWriter creates a new squashfs.Writer with the default settings (gzip compression, autoresolving symlinks, and allowErrors)
func NewWriter() (*Writer, error) {
	return NewWriterWithOptions(true, true, GzipCompression)
}

//NewWriterWithOptions creates a new squashfs.Writer with the given options.
//ResolveSymlinks tries to make sure symlinks aren't broken. It will either try to make the link's location work
func NewWriterWithOptions(resolveSymlinks, allowErrors bool, compressionType int) (*Writer, error) {
	if compressionType < 0 || compressionType > 6 {
		return nil, errors.New("Incorrect compression type")
	}
	if compressionType == 3 {
		return nil, errors.New("Lzo compression is not currently supported")
	}
	out := Writer{
		files: map[string][]*File{
			"/": make([]*File, 0),
		},
		ResolveSymlinks: resolveSymlinks,
		AllowErrors:     allowErrors,
		compression:     compressionType,
	}
	if resolveSymlinks {
		out.symlinkTable = make(map[string]string)
	}
	return &out, nil
}

type fileError struct {
	files []*File
	err   error
}

//convertFile converts the given os.File to a squashfs.File. Returns the errors and converted file to the channels.
func (w *Writer) convertFile(squashfsPath string, file *os.File, subDir bool, fileErrChan chan fileError) {
	var out fileError
	var fil File
	fil.Reader = file
	fil.path = squashfsPath
	fil.name = path.Base(file.Name())
	mode := fil.Mode()

	if mode.IsRegular() {
		fil.filType = inode.BasicFileType
		goto successExit
	} else if mode.IsDir() {
		fil.filType = inode.BasicSymlinkType
		subDirs, err := file.Readdirnames(-1)
		if err != nil {
			if w.AllowErrors && !subDir {
				log.Println("Can't get sub-directories for", file.Name())
				log.Println(err)
			} else {
				out.err = err
			}
			goto failExit
		}
		subDirChan := make(chan fileError)
		for _, filName := range subDirs {
			go func(filename string, returnChan chan fileError) {
				subFil, err := os.Open(filename)
				if err != nil {
					out.err = err
					returnChan <- fileError{
						err: err,
					}
					return
				}
				w.convertFile(fil.Path(), subFil, true, subDirChan)
			}(file.Name()+filName, subDirChan)
		}
		for range subDirs {
			filErr := <-subDirChan
			if filErr.err != nil {
				if w.AllowErrors && !subDir {
					log.Println("Error while adding subdirectory of", file.Name())
					log.Println(filErr.err)
				} else if subDir {
					if out.err == nil {
						out.err = filErr.err
					}
				} else {
					out.err = err
					goto failExit
				}
				continue
			}
			out.files = append(out.files, filErr.files...)
		}
		goto successExit
	} else if mode&os.ModeSymlink == os.ModeSymlink {
		fil.filType = inode.BasicSymlinkType
		symLocation, err := os.Readlink(file.Name())
		if err != nil {
			if w.AllowErrors && !subDir {
				log.Println("Error while getting symlink's information for", file.Name())
				log.Println(err)
			} else {
				out.err = err
			}
			goto failExit
		}
		if w.ResolveSymlinks {
			if val, ok := w.symlinkTable[symLocation]; ok {
				symLocation = val
			} else if val, ok := w.symTableTemp[symLocation]; ok {
				symLocation = val
			} else {
				//TODO: either add the file, or place the file in this location. Maybe defer this until after all the other files are added?
			}
		}
		//TODO: store the symLocation inside the File somehow....
	}
	if w.AllowErrors && !subDir {
		log.Println("Unsupported file type for", file.Name())
	} else {
		out.err = errors.New("Unsupported file type")
	}
failExit: //before this is used, make sure to log or set the error.
	fileErrChan <- out
	return
successExit:
	out.files = []*File{&fil}
	fileErrChan <- out
	return
}

//AddFilesToPath adds the give os.Files to the given path within the squashfs archive.
//If AllowErrors is true, this will ALWAYS return nil
func (w *Writer) AddFilesToPath(squashfsPath string, files ...*os.File) error {
	squashfsPath = path.Clean(squashfsPath)
	if strings.HasPrefix(squashfsPath, "/") {
		squashfsPath = strings.TrimPrefix(squashfsPath, "/")
	}
	if squashfsPath == "." {
		squashfsPath = "/"
	}
	fileErrChan := make(chan fileError)
	for _, fil := range files {
		go w.convertFile(squashfsPath, fil, false, fileErrChan)
	}
}

//AddFiles adds all files given to the root directory
//If AllowErrors is true, this will ALWAYS return nil
func (w *Writer) AddFiles(files ...*os.File) error {
	return w.AddFilesToPath("/", files...)
}

//RemoveFileAt removes the file at filepath from the Writer.
//If multiple files match the given filepath (such as if there are wildcards), all matching files are removed.
//If one or more files are removed, returns true.
func (w *Writer) RemoveFileAt(filepath string) bool {
	//TODO
	return false
}
