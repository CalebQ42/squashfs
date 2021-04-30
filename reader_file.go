package squashfs

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

//File represents a file inside a squashfs archive.
type File struct {
	i        *inode.Inode
	parent   *FS
	r        *Reader
	reader   *fileReader
	name     string
	dirsRead int
}

//File creates a File from the FileInfo.
//*File satisfies fs.File and fs.ReadDirFile.
func (f FileInfo) File() (file *File, err error) {
	file = &File{
		name:   f.name,
		r:      f.r,
		parent: f.parent,
		i:      f.i,
	}
	if file.IsRegular() {
		file.reader, err = f.r.newFileReader(f.i)
	}
	return
}

//File creates a File from the DirEntry.
func (d DirEntry) File() (file *File, err error) {
	return d.r.newFileFromDirEntry(d.en, d.parent)
}

func (r Reader) newFileFromDirEntry(en *directory.Entry, parent *FS) (file *File, err error) {
	file = &File{
		name:   en.Name,
		r:      &r,
		parent: parent,
	}
	file.i, err = r.getInodeFromEntry(en)
	if err != nil {
		return nil, err
	}
	if file.IsRegular() {
		file.reader, err = r.newFileReader(file.i)
	}
	return
}

//Stat returns the File's fs.FileInfo
func (f File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		i:      f.i,
		name:   f.name,
		parent: f.parent,
		r:      f.r,
	}, nil
}

//Read reads the data from the file. Only works if file is a normal file.
func (f File) Read(p []byte) (int, error) {
	if f.i.Type == inode.FileType || f.i.Type == inode.ExtFileType {
		if f.reader == nil {
			return 0, fs.ErrClosed
		}
		return f.reader.Read(p)
	}
	return 0, errors.New("can only read files")
}

//WriteTo writes all data from the file to the writer. This is multi-threaded.
func (f File) WriteTo(w io.Writer) (int64, error) {
	if f.i.Type == inode.FileType || f.i.Type == inode.ExtFileType {
		if f.reader == nil {
			return 0, fs.ErrClosed
		}
		return f.reader.WriteTo(w)
	}
	return 0, errors.New("can only read files")
}

//Close simply nils the underlying reader. Here mostly to satisfy fs.File
func (f *File) Close() error {
	f.reader = nil
	return nil
}

//ReadDir returns n fs.DirEntry's that's contained in the File (if it's a directory).
//If n <= 0 all fs.DirEntry's are returned.
func (f File) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, errors.New("File is not a directory")
	}
	ffs, err := f.FS()
	if err != nil {
		return nil, err
	}
	var beg, end int
	if n <= 0 {
		beg, end = 0, len(ffs.entries)
	} else {
		beg, end = f.dirsRead, f.dirsRead+n
		if end > len(ffs.entries) {
			end = len(ffs.entries)
			err = io.EOF
		}
	}
	out := make([]fs.DirEntry, end-beg)
	for i, ent := range ffs.entries[beg:end] {
		out[i] = f.r.newDirEntry(ent, ffs)
	}
	return out, err
}

//FS returns the File as a FS.
func (f File) FS() (*FS, error) {
	if !f.IsDir() {
		return nil, errors.New("File is not a directory")
	}
	ents, err := f.r.readDirFromInode(f.i)
	if err != nil {
		return nil, err
	}
	return &FS{
		entries: ents,
		parent:  f.parent,
		r:       f.r,
		name:    f.name,
	}, nil
}

//IsDir Yep.
func (f File) IsDir() bool {
	return f.i.Type == inode.DirType || f.i.Type == inode.ExtDirType
}

func (f File) path() string {
	if f.name == "/" {
		return f.name
	}
	return f.parent.path() + "/" + f.name
}

//IsRegular yep.
func (f File) IsRegular() bool {
	return f.i.Type == inode.FileType || f.i.Type == inode.ExtFileType
}

//IsSymlink yep.
func (f File) IsSymlink() bool {
	return f.i.Type == inode.SymType || f.i.Type == inode.ExtSymType
}

//SymlinkPath returns the symlink's target path. Is the File isn't a symlink, returns an empty string.
func (f File) SymlinkPath() string {
	switch f.i.Type {
	case inode.SymType:
		return f.i.Info.(inode.Sym).Path
	case inode.ExtSymType:
		return f.i.Info.(inode.ExtSym).Path
	}
	return ""
}

//GetSymlinkFile returns the File the symlink is pointing to.
//If not a symlink, or the target is unobtainable (such as it being outside the archive or it's absolute) returns nil
func (f File) GetSymlinkFile() *File {
	if !f.IsSymlink() {
		return nil
	}
	if strings.HasPrefix(f.SymlinkPath(), "/") {
		return nil
	}
	sym, err := f.parent.Open(f.SymlinkPath())
	if err != nil {
		return nil
	}
	return sym.(*File)
}

//ExtractionOptions are available options on how to extract.
type ExtractionOptions struct {
	notBase            bool
	DereferenceSymlink bool        //Replace symlinks with the target file
	UnbreakSymlink     bool        //Try to make sure symlinks remain unbroken when extracted, without changing the symlink
	Verbose            bool        //Prints extra info to log on an error
	FolderPerm         fs.FileMode //The permissions used when creating the extraction folder
}

//DefaultOptions is the default ExtractionOptions.
func DefaultOptions() ExtractionOptions {
	return ExtractionOptions{
		DereferenceSymlink: false,
		UnbreakSymlink:     false,
		Verbose:            false,
		FolderPerm:         fs.ModePerm,
	}
}

//ExtractTo extracts the File to the given folder with the default options.
//If the File is a directory, it instead extracts the directory's contents to the folder.
func (f File) ExtractTo(folder string) error {
	return f.ExtractWithOptions(folder, DefaultOptions())
}

//ExtractSymlink extracts the File to the folder with the DereferenceSymlink option.
//If the File is a directory, it instead extracts the directory's contents to the folder.
func (f File) ExtractSymlink(folder string) error {
	return f.ExtractWithOptions(folder, ExtractionOptions{
		DereferenceSymlink: true,
		FolderPerm:         fs.ModePerm,
	})
}

//ExtractWithOptions extracts the File to the given folder with the given ExtrationOptions.
//If the File is a directory, it instead extracts the directory's contents to the folder.
func (f File) ExtractWithOptions(folder string, op ExtractionOptions) error {
	folder = path.Clean(folder)
	if !op.notBase {
		err := os.MkdirAll(folder, op.FolderPerm)
		if err != nil {
			return err
		}
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if f.IsDir() {
		if op.notBase {
			err = os.Mkdir(folder+"/"+f.name, stat.Mode())
			if err != nil && !os.IsExist(err) {
				return err
			}
		} else {
			op.notBase = true
		}
		var ents []fs.DirEntry
		ents, err = f.ReadDir(0)
		if err != nil {
			if op.Verbose {
				log.Println("Error while reading children of", f.path())
			}
			return err
		}
		errChan := make(chan error)
		for i := 0; i < len(ents); i++ {
			go func(ent *DirEntry) {
				fil, goErr := ent.File()
				if goErr != nil {
					errChan <- goErr
					fil.Close()
					return
				}
				errChan <- fil.ExtractWithOptions(folder+"/"+f.name, op)
				fil.Close()
			}(ents[i].(*DirEntry))
		}
		for i := 0; i < len(ents); i++ {
			err = <-errChan
			if err != nil {
				return err
			}
		}
		return nil
	} else if f.IsRegular() {
		var fil *os.File
		fil, err = os.Create(folder + "/" + f.name)
		if os.IsExist(err) {
			os.Remove(folder + "/" + f.name)
			fil, err = os.Create(folder + "/" + f.name)
			if err != nil {
				log.Println("Error while creating", folder+"/"+f.name)
				return err
			}
		} else if err != nil {
			return err
		}
		_, err = io.Copy(fil, f)
		if err != nil {
			log.Println("Error while copying data to", folder+"/"+f.name)
			return err
		}
		return nil
	} else if f.IsSymlink() {
		symPath := f.SymlinkPath()
		if op.DereferenceSymlink {
			fil := f.GetSymlinkFile()
			if fil == nil {
				if op.Verbose {
					log.Println("Symlink path(", symPath, ") is unobtainable:", folder+"/"+f.name)
				}
				return errors.New("cannot get symlink target")
			}
			fil.name = f.name
			err = fil.ExtractWithOptions(folder, op)
			if err != nil {
				if op.Verbose {
					log.Println("Error while extracting the symlink's file:", folder+"/"+f.name)
				}
				return err
			}
			return nil
		} else if op.UnbreakSymlink {
			fil := f.GetSymlinkFile()
			if fil == nil {
				if op.Verbose {
					log.Println("Symlink path(", symPath, ") is unobtainable:", folder+"/"+f.name)
				}
				return errors.New("cannot get symlink target")
			}
			extractLoc := path.Clean(folder + "/" + path.Dir(symPath))
			err = fil.ExtractWithOptions(extractLoc, op)
			if err != nil {
				if op.Verbose {
					log.Println("Error while extracting ", folder+"/"+f.name)
				}
				return err
			}
		}
		err = os.Symlink(f.SymlinkPath(), folder+"/"+f.name)
		if os.IsExist(err) {
			os.Remove(folder + "/" + f.name)
			err = os.Symlink(f.SymlinkPath(), folder+"/"+f.name)
		}
		if err != nil {
			if op.Verbose {
				log.Println("Error while making symlink:", folder+"/"+f.name)
			}
			return err
		}
		return nil
	}
	return errors.New("Unsupported file type. Inode type: " + strconv.Itoa(int(f.i.Type)))
}

//ReadDirFromInode returns a fully populated Directory from a given Inode.
//If the given inode is not a directory it returns an error.
func (r *Reader) readDirFromInode(i *inode.Inode) ([]*directory.Entry, error) {
	var offset uint32
	var metaOffset uint16
	var size uint32
	switch i.Type {
	case inode.DirType:
		offset = i.Info.(inode.Dir).DirectoryIndex
		metaOffset = i.Info.(inode.Dir).DirectoryOffset
		size = uint32(i.Info.(inode.Dir).DirectorySize)
	case inode.ExtDirType:
		offset = i.Info.(inode.ExtDir).DirectoryIndex
		metaOffset = i.Info.(inode.ExtDir).DirectoryOffset
		size = i.Info.(inode.ExtDir).DirectorySize
	default:
		return nil, errors.New("not a directory inode")
	}
	br, err := r.newMetadataReader(int64(r.super.DirTableStart + uint64(offset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(metaOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	ents, err := directory.NewDirectory(br, size)
	if err != nil {
		return nil, err
	}
	return ents, nil
}
