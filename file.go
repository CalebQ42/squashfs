package squashfs

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

var (
	//ErrNotDirectory is returned when you're trying to do directory things with a non-directory
	errNotDirectory = errors.New("File is not a directory")
	//ErrNotFile is returned when you're trying to do file things with a directory
	errNotFile = errors.New("File is not a file")
	//ErrNotReading is returned when running functions that are only meant to be used when reading a squashfs
	errNotReading = errors.New("Function only supported when reading a squashfs")
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
		return nil, errNotReading
	}
	if !f.IsDir() {
		return nil, errNotDirectory
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
		return nil, errNotReading
	}
	if !f.IsDir() {
		return nil, errNotDirectory
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

//SymlinkPath returns the path the symlink is pointing to. If the file ISN'T a symlink, will return an empty string.
//If a path begins with "/" then the symlink is pointing to an absolute path (starting from root, and not a file inside the archive)
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

//Permission returns the os.FileMode of the File. Currently only has the permission bits (the last 9) populated.
func (f *File) Permission() os.FileMode {
	//TODO: possibly populate more os.FileMode bits
	return os.FileMode(f.in.Header.Permissions)
}

func (f *File) ExtractTo(path string) error {
	if f.IsDir() {
		//TODO
	} else if f.IsSymlink() {

	}
}

//Close frees up the memory held up by the underlying reader. Should NOT be called when writing.
//When reading, Close is safe to use, but any subsequent Read calls resets to the beginning of the file.
func (f *File) Close() error {
	if f.IsDir() {
		return errNotFile
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

//ReadDirFromInode returns a fully populated directory.Directory from a given inode.Inode.
//If the given inode is not a directory it returns an error.
func (r *Reader) readDirFromInode(i *inode.Inode) (*directory.Directory, error) {
	var offset uint32
	var metaOffset uint16
	var size uint16
	switch i.Type {
	case inode.BasicDirectoryType:
		offset = i.Info.(inode.BasicDirectory).DirectoryIndex
		metaOffset = i.Info.(inode.BasicDirectory).DirectoryOffset
		size = i.Info.(inode.BasicDirectory).DirectorySize
	case inode.ExtDirType:
		offset = i.Info.(inode.ExtendedDirectory).Init.DirectoryIndex
		metaOffset = i.Info.(inode.ExtendedDirectory).Init.DirectoryOffset
		size = uint16(i.Info.(inode.ExtendedDirectory).Init.DirectorySize)
	default:
		return nil, errors.New("Not a directory inode")
	}
	br, err := r.newMetadataReader(int64(r.super.DirTableStart + uint64(offset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(metaOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	dir, err := directory.NewDirectory(br, size)
	if err != nil {
		return dir, err
	}
	return dir, nil
}

//GetInodeFromEntry returns the inode associated with a given directory.Entry
func (r *Reader) getInodeFromEntry(en *directory.Entry) (*inode.Inode, error) {
	br, err := r.newMetadataReader(int64(r.super.InodeTableStart + uint64(en.Header.InodeOffset)))
	if err != nil {
		return nil, err
	}
	_, err = br.Seek(int64(en.Init.Offset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	i, err := inode.ProcessInode(br, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	return i, nil
}
