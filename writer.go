package squashfs

import (
	"errors"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/CalebQ42/squashfs/internal/inode"
)

//Writer is an interface to write a squashfs. Doesn't write until you call Write (TODO: maybe not do Write...).
//If AllowErrors is true, when errors are encountered, it just prints to the log instead of failing.
type Writer struct {
	files           map[string][]*File
	directories     []string
	symlinkTable    map[string]string //symlinkTable holds info about symlink'd to files that had to be moved from their original position. [originalpath]newpath
	ResolveSymlinks bool
	AllowErrors     bool
	compression     int
	temp            []*File
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

//convertFile converts the given os.File to a squashfs.File and then adds it to the Writer's temp File slice.
func (w *Writer) convertFile(squashfsPath string, file *os.File, errChan chan error) {
	var fil File
	fil.Reader = file
	fil.name = path.Base(file.Name())
	fil.path = squashfsPath
	stat, err := file.Stat()
	if err != nil {
		if w.AllowErrors {
			log.Println("Error while getting FileInfo for", file.Name()+":")
			log.Println(err)
			err = nil
		}
		errChan <- err
		return
	}
	if stat.IsDir() {
		fil.filType = inode.BasicDirectoryType
		dirs, err := file.Readdirnames(-1)
		if err != nil {
			if w.AllowErrors {
				log.Println("Error when getting directory names for", file.Name()+":")
				log.Println(err)
				err = nil
			}
			errChan <- err
			return
		}
		subDirErrChan := make(chan error)
		for _, dir := range dirs {
			go func(newFilename string, errChan chan error) {
				subFil, err := os.Open(file.Name() + newFilename)
				if err != nil {
					if w.AllowErrors {
						log.Println("Error when opening sub-directory", subFil.Name()+":")
						log.Println(err)
						err = nil
					}
					errChan <- err
					return
				}
				subDirErrChan := make(chan error)
				w.convertFile(fil.Path(), subFil, subDirErrChan)
				errChan <- <-subDirErrChan
				return
			}(dir, subDirErrChan)
		}
		for range dirs {
			err = <-subDirErrChan
			if err != nil {
				errChan <- err
				return
			}
		}
		w.temp = append(w.temp, &fil)
		errChan <- nil
		return
	} else if stat.Mode().IsRegular() {
		fil.filType = inode.BasicFileType
		w.temp = append(w.temp, &fil)
		errChan <- nil
		return
	} else if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkLocation, err := os.Readlink(file.Name())
		if err != nil {
			if w.AllowErrors {
				log.Println("Error when reading symlink's target", file.Name()+":")
				log.Println(err)
				err = nil
			}
			errChan <- err
			return
		}
		if w.ResolveSymlinks {
			if w.symlinkTable[linkLocation] != "" {
				linkLocation = w.symlinkTable[linkLocation]
			}
		}
		//TODO: finish symlink support
	}
	errChan <- errors.New("Unsupported file type")
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
	errChan := make(chan error)
	for _, fil := range files {
		go w.convertFile(squashfsPath, fil, errChan)
	}
	var firstError error
	for range files {
		err := <-errChan
		if firstError != nil && err != nil {
			firstError = err
		}
	}
	if firstError != nil {
		w.temp = nil
		return firstError
	}
	for _, tempFil := range w.temp {
		if tempFil.path != "/" {
			ind := sort.SearchStrings(w.directories, tempFil.path)
			if ind == len(w.directories) {
				w.directories = append(w.directories, tempFil.path)
				sort.Strings(w.directories)
			}
		}
		w.files[tempFil.path] = append(w.files[tempFil.path], tempFil)
	}
	w.temp = nil
	return nil
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
