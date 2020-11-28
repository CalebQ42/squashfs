package squashfs

import (
	"errors"
	"io"
	"strings"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

var (
	//ErrNotDirectory is returned when you're trying to do directory things with a non-directory
	ErrNotDirectory = errors.New("File is not a directory")
	//ErrNotFile is returned when you're trying to do file things with a directory
	ErrNotFile = errors.New("File is not a file")
	//ErrNotReading is returned when running functions that are only meant to be used when reading a squashfs
	ErrNotReading = errors.New("Function only supported when reading a squashfs")
)

//File is the main way to interact with files within squashfs, or when putting files into a squashfs.
//File can be either a file or folder. When reading from a squashfs, it reads from the datablocks.
//When writing, this holds the information on WHERE the file will be placed inside the archive.
type File struct {
	Name    string       //The name of the file or folder. Root folder will not have a name ("")
	Parent  *File        //The parent directory. Should ALWAYS be a folder. If it's the root directory, will be nil
	Reader  io.Reader    //Underlying reader. When writing, will probably be an os.File. When reading this is kept nil UNTIL reading to save memory.
	Path    string       //The path to the folder the File is located in.
	r       *Reader      //The squashfs.Reader where this file is contained.
	in      *inode.Inode //Underlyting inode when reading.
	filType int          //The file's type, using inode types.
}

//get a File from a directory.entry
func (r *Reader) newFileFromDirEntry(entry *directory.Entry) (fil *File, err error) {
	fil = new(File)
	fil.in, err = r.getInodeFromEntry(entry)
	if err != nil {
		return nil, err
	}
	fil.Name = entry.Name
	fil.r = r
	fil.filType = fil.in.Type
	return
}

//GetChildren returns a *squashfs.File slice of every direct child of the directory. If the File is not a directory, will return ErrNotDirectory
func (f *File) GetChildren() (children []*File, err error) {
	children = make([]*File, 0)
	if f.r == nil {
		return nil, ErrNotReading
	}
	if !f.IsDir() {
		return nil, ErrNotDirectory
	}
	dir, err := f.r.readDirFromInode(f.in)
	if err != nil {
		return
	}
	var fil *File
	for _, entry := range dir.Entries {
		fil, err = f.r.newFileFromDirEntry(&entry)
		if err != nil {
			return
		}
		fil.Parent = f
		if f.Name != "" {
			fil.Path = f.Path + "/" + f.Name
		}
		children = append(children, fil)
	}
	return
}

//GetChildrenRecursively returns ALL children. Goes down ALL folder paths.
func (f *File) GetChildrenRecursively() (children []*File, err error) {
	children = make([]*File, 0)
	if f.r == nil {
		return nil, ErrNotReading
	}
	if !f.IsDir() {
		return nil, ErrNotDirectory
	}
	chil, err := f.GetChildren()
	if err != nil {
		return
	}
	var childFolders []*File
	for _, child := range chil {
		children = append(children, child)
		if child.IsDir() {
			childFolders = append(childFolders, child)
		}
	}
	for _, folds := range childFolders {
		var childs []*File
		childs, err = folds.GetChildrenRecursively()
		if err != nil {
			return
		}
		children = append(children, childs...)
	}
	return
}

//GetFileAtPath tries to return the File at the given path, relative to the file.
//Returns nil if called on something other then a folder, OR if the path goes oustide the archive.
func (f *File) GetFileAtPath(path string) *File {
	if path == "" {
		return f
	}
	path = strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/")
	if path != "" && !f.IsDir() {
		return nil
	}
	for strings.HasSuffix(path, "./") {
		//since you can TECHNICALLY have an infinite amount of ./ and it would still be valid.
		path = strings.TrimPrefix(path, "./")
	}
	split := strings.Split(path, "/")
	if split[0] == ".." && f.Name == "" {
		return nil
	} else if split[0] == ".." {
		if f.Parent != nil {
			return f.Parent.GetFileAtPath(strings.Join(split[1:], "/"))
		}
		return nil
	}
	children, err := f.GetChildren()
	if err != nil {
		return nil
	}

	for _, child := range children {
		if child.Name == split[0] {
			return child.GetFileAtPath(strings.Join(split[1:], "/"))
		}
	}
	return nil
}

//IsDir returns if the file is a directory.
func (f *File) IsDir() bool {
	return f.filType == inode.BasicDirectoryType || f.filType == inode.ExtDirType
}

//IsSymlink returns if the file is a symlink.
func (f *File) IsSymlink() bool {
	return f.filType == inode.BasicSymlinkType || f.filType == inode.ExtSymlinkType
}

//SymlinkPath returns the path the symlink is pointing to. If the file ISN'T a symlink, will return an empty string
func (f *File) SymlinkPath() string {
	switch f.filType {
	case inode.BasicSymlinkType:
		return f.in.Info.(inode.BasicSymlink).Path
	case inode.ExtSymlinkType:
		return f.in.Info.(inode.ExtendedSymlink).Path
	default:
		return ""
	}
}

//GetSymlinkFile tries to return the squashfs.File associated with the symlink
func (f *File) GetSymlinkFile() *File {
	if !f.IsSymlink() {
		return nil
	}
	if strings.HasSuffix(f.SymlinkPath(), "/") {
		return nil
	}
	return f.r.GetFileAtPath(f.SymlinkPath())
}

//Close frees up the memory held up by the underlying reader. Should NOT be called when writing.
//When reading, Close is safe to use, but any subsequent Read calls resets to the beginning of the file.
func (f *File) Close() error {
	if f.IsDir() {
		return ErrNotFile
	}
	if closer, is := f.Reader.(io.Closer); is {
		closer.Close()
	}
	f.Reader = nil
	return nil
}

//Read from the file. Doesn't do anything fancy, just pases it to the underlying io.Reader. If a directory, return io.EOF.
func (f *File) Read(p []byte) (int, error) {
	if f.IsDir() {
		return 0, io.EOF
	}
	var err error
	if f.Reader == nil && f.r != nil {
		f.Reader, err = f.r.newFileReader(f.in)
		if err != nil {
			return 0, err
		}
	}
	return f.Reader.Read(p)
}
