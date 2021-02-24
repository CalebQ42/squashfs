package squashfs

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"syscall"

	"github.com/CalebQ42/squashfs/internal/compression"
)

//Writer is used to creaste squashfs archives. Currently unusable.
//TODO: Make usable
type Writer struct {
	compressor      compression.Compressor
	structure       map[string][]*fileHolder
	symlinkTable    map[string]string //[oldpath]newpath
	uidGUIDTable    []int
	compressionType int
	//BlockSize is how large the data blocks are. Can be between 4096 (4KB) and 1048576 (1 MB).
	//If BlockSize is not inside that range, it will be set to within the range before writing.
	//Default is 1048576.
	BlockSize uint32
	//Flags are the SuperblockFlags used when writing the archive.
	//Currently Duplicates, Exportable, UncompressedXattr, NoXattr values are ignored
	Flags       SuperblockFlags
	allowErrors bool

	//variables used when actually writing.
	superblock superblock
	frags      []fragment
}

//NewWriter creates a new with the default options (Gzip compression and allow errors)
func NewWriter() (*Writer, error) {
	return NewWriterWithOptions(GzipCompression, true)
}

//NewWriterWithOptions creates a new squashfs.Writer with the given options.
//compressionType can be of any types, except LZO (which this library doesn't have support for yet)
//allowErrors determines if, when adding folders, it allows errors encountered with it's sub-directories and instead logs the errors.
func NewWriterWithOptions(compressionType int, allowErrors bool) (*Writer, error) {
	if compressionType < 0 || compressionType > 6 {
		return nil, errors.New("Incorrect compression type")
	}
	if compressionType == 3 {
		return nil, errors.New("LZO compression is not (currently) supported")
	}
	return &Writer{
		structure: map[string][]*fileHolder{
			"/": make([]*fileHolder, 0),
		},
		symlinkTable:    make(map[string]string),
		compressionType: compressionType,
		allowErrors:     allowErrors,
		BlockSize:       uint32(1048576),
		Flags:           DefaultFlags,
	}, nil
}

//fileHolder holds the necessary information about a given file inside of a squashfs
type fileHolder struct {
	reader      io.Reader
	path        string
	name        string
	symLocation string
	blockSizes  []uint32
	GUID        int
	perm        int
	size        uint64
	UID         int
	folder      bool
	symlink     bool
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
	if w.Contains(filepath) {
		return errors.New("File already exists at " + filepath)
	}
	var holder fileHolder
	holder.path = path.Dir(filepath)
	holder.name = path.Base(filepath)
	holder.reader = file
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	holder.folder = stat.IsDir()
	holder.symlink = (stat.Mode()&os.ModeSymlink == os.ModeSymlink)
	holder.perm = int(stat.Mode().Perm())
	//Thanks to https://stackoverflow.com/questions/58179647/getting-uid-and-gid-of-a-file for uid and guid getting
	if stat, ok := stat.Sys().(*syscall.Stat_t); ok {
		holder.UID = int(stat.Uid)
		holder.GUID = int(stat.Gid)
	}
	if sort.SearchInts(w.uidGUIDTable, holder.UID) == len(w.uidGUIDTable) {
		w.uidGUIDTable = append(w.uidGUIDTable, holder.UID)
		sort.Ints(w.uidGUIDTable)
	}
	if sort.SearchInts(w.uidGUIDTable, holder.GUID) == len(w.uidGUIDTable) {
		w.uidGUIDTable = append(w.uidGUIDTable, holder.GUID)
		sort.Ints(w.uidGUIDTable)
	}
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
			if err != nil && !w.allowErrors {
				for _, dir := range dirsAdded {
					w.Remove(dir)
				}
				return err
			} else if err != nil {
				log.Println("Error while adding", fil.Name())
				log.Println(err)
			}
			if !w.allowErrors {
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
func (w *Writer) AddReaderTo(filepath string, reader io.Reader, size uint64) error {
	filepath = path.Clean(filepath)
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	if w.Contains(filepath) {
		return errors.New("File already exists at " + filepath)
	}
	var holder fileHolder
	holder.path = path.Dir(filepath)
	holder.name = path.Base(filepath)
	holder.size = size
	holder.reader = reader
	w.structure[holder.path] = append(w.structure[holder.path], &holder)
	return nil
}

//AddFolderTo adds a folder at the given path. IF the folder is already present, it sets the folder's permissions.
//If the path points to a non-folder (such as a file or symlink), an error is returned
func (w *Writer) AddFolderTo(folderpath string, permission fs.FileMode) error {
	folderpath = path.Clean(folderpath)
	tmp := w.holderAt(folderpath)
	if tmp != nil {
		if !tmp.folder {
			return errors.New("Path is not a folder: " + folderpath)
		}
		tmp.perm = int(permission.Perm())
		return nil
	}
	file := fileHolder{
		path:   path.Dir(folderpath),
		name:   path.Base(folderpath),
		perm:   int(permission | fs.ModePerm),
		folder: true,
	}
	w.structure[file.path] = append(w.structure[file.path], &file)
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
					if len(w.structure[structDir]) > 1 {
						w.structure[structDir][i] = w.structure[structDir][len(w.structure[structDir])-1]
						w.structure[structDir] = w.structure[structDir][:len(w.structure[structDir])-1]
					} else {
						w.structure[structDir] = nil
					}
				}
			}
		}
	}
	return matchFound
}

//FixSymlinks will scan through the squashfs archive and try to find broken symlinks and fix them.
//This done by replacing the symlink with the target file and then pointing other symlinks to that file.
//If all symlinks can be resolved, the error slice will be nil, and the bool false, otherwise all errors occured will be in the slice.
func (w *Writer) FixSymlinks() (errs []error, problems bool) {
	for dir, holderSlice := range w.structure {
		for i, holder := range holderSlice {
			if !holder.symlink {
				continue
			}
			sym := holder.symLocation
			if !path.IsAbs(holder.symLocation) {
				sym = path.Join(dir, holder.symLocation)
			}
			if path, ok := w.symlinkTable[sym]; ok {
				w.structure[dir][i].symLocation = path
				continue
			}
			if path.IsAbs(sym) || strings.HasPrefix(sym, "../") {
				var symFil *os.File
				var err error
				if strings.HasPrefix(sym, "../") {
					holderFil, ok := holder.reader.(*os.File)
					if !ok {
						problems = true
						errs = append(errs, errors.New("Cannot resolve symlink at "+dir+holder.name))
						continue
					}
					symFilPath := path.Dir(holderFil.Name())
					symFilPath = path.Join(symFilPath, holder.symLocation)
					symFil, err = os.Open(symFilPath)
				} else {
					symFil, err = os.Open(sym)
				}
				if err != nil {
					problems = true
					errs = append(errs, err)
					continue
				}
				suc := w.Remove(dir + holder.name)
				if !suc {
					problems = true
					errs = append(errs, errors.New("Cannot resolve symlink at "+dir+holder.name))
					continue
				}
				err = w.AddFileTo(dir+holder.name, symFil)
				if err != nil {
					w.structure[dir] = append(w.structure[dir], holder)
					problems = true
					errs = append(errs, err)
					continue
				}
				w.symlinkTable[sym] = dir + holder.name
			} else {
				symHolder := w.holderAt(sym)
				if symHolder != nil {
					w.symlinkTable[sym] = sym
					continue
				}
				holderFil, ok := holder.reader.(*os.File)
				if !ok {
					problems = true
					errs = append(errs, errors.New("Cannot resolve symlink at "+dir+holder.name))
					continue
				}
				symFilPath := path.Dir(holderFil.Name())
				symFilPath = path.Join(symFilPath, holder.symLocation)
				symFil, err := os.Open(symFilPath)
				if err != nil {
					problems = true
					errs = append(errs, err)
					continue
				}
				err = w.AddFileTo(sym, symFil)
				if err != nil {
					problems = true
					errs = append(errs, err)
					continue
				}
				w.symlinkTable[sym] = sym
			}
		}
	}
	return
}

func (w *Writer) holderAt(filepath string) *fileHolder {
	filepath = path.Clean(filepath)
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	dir, name := path.Split(filepath)
	if holderSlice, ok := w.structure[dir]; ok {
		for _, holder := range holderSlice {
			if holder.name == name {
				return holder
			}
		}
	}
	return nil
}

//Contains returns whether a file is present at the given filepath
func (w *Writer) Contains(filepath string) bool {
	filepath = path.Clean(filepath)
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	dir, name := path.Split(filepath)
	if holderSlice, ok := w.structure[dir]; ok {
		for _, holder := range holderSlice {
			if holder.name == name {
				return true
			}
		}
	}
	return false
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
