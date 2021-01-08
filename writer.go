package squashfs

import (
	"errors"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type fileHolder struct {
	reader      io.Reader
	path        string
	name        string
	symLocation string
	folder      bool
	symlink     bool
}

//Writer is used to creaste squashfs archives. Currently unusable
//TODO: Make usable
type Writer struct {
	structure       map[string][]*fileHolder
	symlinkTable    map[string]string //[oldpath]newpath
	compressionType int
	allowErrors     bool //AllowErrors allows errors when adding folders and their children.
}

//NewWriter creates a new
func NewWriter() (*Writer, error) {
	return NewWriterWithOptions(GzipCompression, true)
}

func NewWriterWithOptions(compressionType int, allowErrors bool) (*Writer, error) {

}

//AddFile attempts to add an os.File to the archive at it's root.
func (w *Writer) AddFile(file *os.File) error {
	return w.AddFileToFolder("/", file)
}

//AddFileToFolder adds the given file to the squashfs archive, placing it inside the given folder.
func (w *Writer) AddFileToFolder(folder string, file *os.File) error {
	name := path.Base(file.Name())
	if !strings.HasSuffix(folder, "/") {
		folder += "/"
	}
	return w.AddFileTo(folder+name, file)
}

//AddFileTo adds the given file to the squashfs archive at the given filepath.
func (w *Writer) AddFileTo(filepath string, file *os.File) error {
	filepath = path.Clean(filepath)
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	var holder fileHolder
	holder.path, holder.name = path.Split(filepath)
	holder.reader = file
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	holder.folder = stat.IsDir()
	holder.symlink = (stat.Mode()&os.ModeSymlink == os.ModeSymlink)
	if holder.symlink {
		target, err := os.Readlink(file.Name())
		if err != nil {
			return err
		}
		holder.symLocation = target
	} else if holder.folder {
		subDirNames, err := file.Readdirnames(-1)
		if err != nil {
			return err
		}
		dirsAdded := make([]string, 0)
		for _, subDir := range subDirNames {
			fil, err := os.Open(file.Name() + subDir)
			if err != nil {
				return err
			}
			err = w.AddFileToFolder(holder.path+"/"+holder.name, fil)
			if err != nil && !w.AllowErrors {
				for _, dir := range dirsAdded {
					w.Remove(dir)
				}
				return err
			} else if err != nil {
				log.Println("Error while adding", fil.Name())
				log.Println(err)
			}
			if !w.AllowErrors {
				dirsAdded = append(dirsAdded, holder.path+"/"+holder.name)
			}
		}
	} else if !stat.Mode().IsRegular() {
		return errors.New("Unsupported file type " + file.Name())
	}
	w.structure[holder.path] = append(w.structure[holder.path], &holder)
	return nil
}

//AddReaderTo adds the data from the given reader to the archive as a file located at the given filepath.
//Data from the reader is not read until the squashfs archive is writen.
//If the given reader implements io.Closer, it will be closed after it is fully read.
func (w *Writer) AddReaderTo(filepath string, reader io.Reader) error {
	filepath = path.Clean(filepath)
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	var holder fileHolder
	holder.path, holder.name = path.Split(filepath)
	holder.reader = reader
	w.structure[holder.path] = append(w.structure[holder.path], &holder)
	return nil
}

//Remove tries to remove the file(s) at the given filepath. If wildcards are used, it will remove all files that match.
//Returns true if one or more files are removed.
func (w *Writer) Remove(filepath string) bool {
	var matchFound bool
	filepath = path.Clean(filepath)
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	dir, name := path.Split(filepath)
	for structDir, files := range w.structure {
		if match, _ := path.Match(dir, structDir); match {
			for i, fil := range files {
				if match, _ = path.Match(name, fil.name); match {
					matchFound = true
					if i != len(files)-1 {
						w.structure[structDir] = append(w.structure[structDir][:i], w.structure[structDir][i+1:]...)
					} else {
						w.structure[structDir] = w.structure[structDir][:i]
					}
				}
			}
		}
	}
	return matchFound
}

//FixSymlinks will scan through the squashfs archive and try to find broken symlinks and fix them.
//This done by replacing the symlink with the target file and then pointing other symlinks to that file.
//
//If this is not run before writing, you may end up with broken symlinks.
func (w *Writer) FixSymlinks() error {
	return errors.New("DON'T")
}

//WriteToFilename creates the squashfs archive with the given filepath.
func (w *Writer) WriteToFilename(filepath string) error {
	newFil, err := os.Create(filepath)
	if err != nil {
		return err
	}
	_, err = w.WriteTo(newFil)
	return err
}

//WriteTo attempts to write the archive to the given io.Writer.
func (w *Writer) WriteTo(write io.Writer) (int64, error) {
	return 0, errors.New("I SAID DON'T")
}
