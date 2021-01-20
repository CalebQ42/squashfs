package squashfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

//TODO: implement fs.FS, fs.ReadDirFile, fs.ReadFileFS, fs.StatFS, fs.SubFS with 1.16

var (
	//ErrNotDirectory is returned when you're trying to do directory things with a non-directory
	errNotDirectory = errors.New("File is not a directory")
	//ErrNotFile is returned when you're trying to do file things with a directory
	errNotFile = errors.New("File is not a file")
	//ErrNotReading is returned when running functions that are only meant to be used when reading a squashfs
	errNotReading = errors.New("Function only supported when reading a squashfs")
	//ErrBrokenSymlink is returned when using ExtractWithOptions with the unbreakSymlink set to true, but the symlink's file cannot be extracted.
	ErrBrokenSymlink = errors.New("Extracted symlink is probably broken")
)

//File is the main way to interact with files within squashfs, or when putting files into a squashfs.
//File can be either a file or folder. When reading from a squashfs, it reads from the datablocks.
//When writing, this holds the information on WHERE the file will be placed inside the archive.
//
//If copying data from a squashfs, the returned reader from io.Sys() implements io.WriterTo which
//will be significantly faster then calling Read directly.
//Ex: use io.Sys().(io.Reader) for io.Copy instead of using the File directly.
//
//Implements os.FileInfo and io.Reader
type File struct {
	reader  io.Reader
	Parent  *File
	r       *Reader //Underlying reader. When writing, will probably be an os.File. When reading this is kept nil UNTIL reading to save memory.
	in      *inode.Inode
	name    string
	dir     string
	filType int //The file's type, using inode types.

}

//get a File from a directory.entry
func (r *Reader) newFileFromDirEntry(entry *directory.Entry) (fil *File, err error) {
	fil = new(File)
	fil.in, err = r.getInodeFromEntry(entry)
	if err != nil {
		return nil, err
	}
	fil.name = entry.Name
	fil.r = r
	fil.filType = fil.in.Type
	return
}

//Name is the file's name
func (f *File) Name() string {
	return f.name
}

//Size is the complete size of the file. Zero if it's not a file.
func (f *File) Size() int64 {
	switch f.filType {
	case inode.FileType:
		return int64(f.in.Info.(inode.File).Size)
	case inode.ExtFileType:
		return int64(f.in.Info.(inode.ExtFile).Size)
	default:
		return 0
	}
}

//ModTime is the time of last modification.
func (f *File) ModTime() time.Time {
	return time.Unix(int64(f.in.Header.ModifiedTime), 0)
}

//Sys returns the underlying reader. If the reader isn't initialized, it will initialize it.
//If called on something other then a file, returns nil.
func (f *File) Sys() interface{} {
	if !f.IsFile() {
		return nil
	}
	if f.reader == nil && f.r != nil {
		var err error
		f.reader, err = f.r.newFileReader(f.in)
		if err != nil {
			return nil
		}
	}
	return f.reader
}

//TODO: Implement below when 1.16 drops to satisfy fs.File

//Stat simply returns the file. It's simply here to satisfy fs.File
// func (f *File) Stat() (fs.FileInfo, error) {
// 	return f, nil
// }

//Close does nothing. It's simply here to satisfy fs.File
// func (f *File) Close() error {
// 	return nil
// }

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
		if f.name != "" {
			fil.dir = f.Path()
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
	children, err = f.GetChildren()
	if err != nil {
		return
	}
	var childFolders []*File
	for _, child := range children {
		if child.IsDir() {
			childFolders = append(childFolders, child)
		}
	}
	for _, folds := range childFolders {
		var childs []*File
		childs, err = folds.GetChildrenRecursively()
		if err != nil {
			fmt.Println(err)
			return
		}
		children = append(children, childs...)
	}
	return
}

//Path returns the path of the file within the archive.
func (f *File) Path() string {
	if f.name == "" {
		return f.dir
	}
	return f.dir + "/" + f.name
}

//GetFileAtPath tries to return the File at the given path, relative to the file.
//Returns nil if called on something other then a folder, OR if the path goes oustide the archive.
//Allows wildcards supported by path.Match (namely * and ?) and will return the FIRST file that matches.
func (f *File) GetFileAtPath(dirPath string) *File {
	dirPath = path.Clean(dirPath)
	dirPath = strings.TrimPrefix(dirPath, "/")
	if dirPath == "" || dirPath == "." {
		return f
	}
	if dirPath != "." && !f.IsDir() {
		return nil
	}
	split := strings.Split(dirPath, "/")
	if split[0] == ".." && f.name == "" {
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
		eq, _ := path.Match(split[0], child.name)
		if eq {
			return child.GetFileAtPath(strings.Join(split[1:], "/"))
		}
	}
	return nil
}

//TODO: add with 1.16
//Open is the same as GetFileAtPath to implement fs.FS
// func (f *File) Open(name string) (fs.File, error) {
// 	tmp := f.GetFileAtPath(name)
// 	if tmp == nil {
// 		return tmp, fs.ErrNotExist
// 	}
// 	return tmp, nil
// }

//IsDir returns if the file is a directory.
func (f *File) IsDir() bool {
	return f.filType == inode.DirType || f.filType == inode.ExtDirType
}

//IsSymlink returns if the file is a symlink.
func (f *File) IsSymlink() bool {
	return f.filType == inode.SymType || f.filType == inode.ExtSymType
}

//IsFile returns if the file is a file.
func (f *File) IsFile() bool {
	return f.filType == inode.FileType || f.filType == inode.ExtFileType
}

//SymlinkPath returns the path the symlink is pointing to. If the file ISN'T a symlink, will return an empty string.
//If a path begins with "/" then the symlink is pointing to an absolute path (starting from root, and not a file inside the archive)
func (f *File) SymlinkPath() string {
	switch f.filType {
	case inode.SymType:
		return f.in.Info.(inode.Sym).Path
	case inode.ExtSymType:
		return f.in.Info.(inode.ExtSym).Path
	default:
		return ""
	}
}

//GetSymlinkFile tries to return the squashfs.File associated with the symlink. If the file isn't a symlink
//or the symlink points to a location outside the archive, nil is returned.
func (f *File) GetSymlinkFile() *File {
	if !f.IsSymlink() {
		return nil
	}
	if strings.HasSuffix(f.SymlinkPath(), "/") {
		return nil
	}
	return f.Parent.GetFileAtPath(f.SymlinkPath())
}

//GetSymlinkFileRecursive tries to return the squasfs.File associated with the symlink. It will recursively
//try to get the symlink's file. This will return either a non-symlink File, or nil.
func (f *File) GetSymlinkFileRecursive() *File {
	if !f.IsSymlink() {
		return nil
	}
	if strings.HasSuffix(f.SymlinkPath(), "/") {
		return nil
	}
	sym := f
	for {
		sym = sym.GetSymlinkFile()
		if sym == nil {
			return nil
		}
		if !sym.IsSymlink() {
			return sym
		}
	}
}

//Mode returns the os.FileMode of the File. Sets mode bits for directories and symlinks.
func (f *File) Mode() os.FileMode {
	mode := os.FileMode(f.in.Header.Permissions)
	switch {
	case f.IsDir():
		mode = mode | os.ModeDir
	case f.IsSymlink():
		mode = mode | os.ModeSymlink
	}
	return mode
}

//TODO: Implement with 1.16

//Type returns the type bits from fs.FileMode
// func (f *File) Type() fs.FileInfo {
// 	return Mod() ^&fs.ModePerm
// }

//ExtractTo extracts the file to the given path. This is the same as ExtractWithOptions(path, false, false, os.ModePerm, false).
//Will NOT try to keep symlinks valid, folders extracted will have the permissions set by the squashfs, but the folder to make path will have full permissions (777).
//
//Will try it's best to extract all files, and if any errors come up, they will be appended to the error slice that's returned.
func (f *File) ExtractTo(path string) []error {
	return f.ExtractWithOptions(path, false, false, os.ModePerm, false)
}

//ExtractSymlink is similar to ExtractTo, but when it extracts a symlink, it instead extracts the file associated with the symlink in it's place.
//This is the same as ExtractWithOptions(path, true, false, os.ModePerm, false)
func (f *File) ExtractSymlink(path string) []error {
	return f.ExtractWithOptions(path, true, false, os.ModePerm, false)
}

//ExtractWithOptions will extract the file to the given path, while allowing customization on how it works. ExtractTo is the "default" options.
//Will try it's best to extract all files, and if any errors come up, they will be appended to the error slice that's returned.
//Should only return multiple errors if extracting a folder.
//
//If dereferenceSymlink is set, instead of extracting a symlink, it will extract the file the symlink is pointed to in it's place.
//If both dereferenceSymlink and unbreakSymlink is set, dereferenceSymlink takes precendence.
//
//If unbreakSymlink is set, it will also try to extract the symlink's associated file. WARNING: the symlink's file may have to go up the directory to work.
//If unbreakSymlink is set and the file cannot be extracted, a ErrBrokenSymlink will be appended to the returned error slice.
//
//folderPerm only applies to the folders created to get to path. Folders from the archive are given the correct permissions defined by the archive.
func (f *File) ExtractWithOptions(path string, dereferenceSymlink, unbreakSymlink bool, folderPerm os.FileMode, verbose bool) (errs []error) {
	errs = make([]error, 0)
	err := os.MkdirAll(path, folderPerm)
	if err != nil {
		return []error{err}
	}
	switch {
	case f.IsDir():
		if f.name != "" {
			//TODO: check if folder is present, and if so, try to set it's permission
			err = os.Mkdir(path+"/"+f.name, os.ModePerm)
			if err != nil {
				if verbose {
					fmt.Println("Error while making: ", path+"/"+f.name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
			var fil *os.File
			fil, err = os.Open(path + "/" + f.name)
			if err != nil {
				if verbose {
					fmt.Println("Error while opening:", path+"/"+f.name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
			fil.Chown(int(f.r.idTable[f.in.Header.UID]), int(f.r.idTable[f.in.Header.GID]))
			//don't mention anything when it fails. Because it fails often. Probably has something to do about uid & gid 0
			// if err != nil {
			// 	if verbose {
			// 		fmt.Println("Error while changing owner:", path+"/"+f.Name)
			// 		fmt.Println(err)
			// 	}
			// 	errs = append(errs, err)
			// }
			err = fil.Chmod(f.Mode())
			if err != nil {
				if verbose {
					fmt.Println("Error while changing owner:", path+"/"+f.name)
					fmt.Println(err)
				}
				errs = append(errs, err)
			}
		}
		var children []*File
		children, err = f.GetChildren()
		if err != nil {
			if verbose {
				fmt.Println("Error getting children for:", f.Path())
				fmt.Println(err)
			}
			errs = append(errs, err)
			return
		}
		finishChan := make(chan []error)
		for _, child := range children {
			go func(child *File) {
				if f.name == "" {
					finishChan <- child.ExtractWithOptions(path, dereferenceSymlink, unbreakSymlink, folderPerm, verbose)
				} else {
					finishChan <- child.ExtractWithOptions(path+"/"+f.name, dereferenceSymlink, unbreakSymlink, folderPerm, verbose)
				}
			}(child)
		}
		for range children {
			errs = append(errs, (<-finishChan)...)
		}
		return
	case f.IsFile():
		var fil *os.File
		fil, err = os.Create(path + "/" + f.name)
		if os.IsExist(err) {
			err = os.Remove(path + "/" + f.name)
			if err != nil {
				if verbose {
					fmt.Println("Error while making:", path+"/"+f.name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
			fil, err = os.Create(path + "/" + f.name)
			if err != nil {
				if verbose {
					fmt.Println("Error while making:", path+"/"+f.name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
		} else if err != nil {
			if verbose {
				fmt.Println("Error while making:", path+"/"+f.name)
				fmt.Println(err)
			}
			errs = append(errs, err)
			return
		} //Since we will be reading from the file
		_, err = io.Copy(fil, f.Sys().(io.Reader))
		if err != nil {
			if verbose {
				fmt.Println("Error while Copying data to:", path+"/"+f.name)
				fmt.Println(err)
			}
			errs = append(errs, err)
			return
		}
		fil.Chown(int(f.r.idTable[f.in.Header.UID]), int(f.r.idTable[f.in.Header.GID]))
		//don't mention anything when it fails. Because it fails often. Probably has something to do about uid & gid 0
		// if err != nil {
		// 	if verbose {
		// 		fmt.Println("Error while changing owner:", path+"/"+f.Name)
		// 		fmt.Println(err)
		// 	}
		// 	errs = append(errs, err)
		// 	return
		// }
		err = fil.Chmod(f.Mode())
		if err != nil {
			if verbose {
				fmt.Println("Error while setting permissions for:", path+"/"+f.name)
				fmt.Println(err)
			}
			errs = append(errs, err)
		}
		return
	case f.IsSymlink():
		symPath := f.SymlinkPath()
		if dereferenceSymlink {
			fil := f.GetSymlinkFile()
			if fil == nil {
				if verbose {
					fmt.Println("Symlink path(", symPath, ") is outside the archive:"+path+"/"+f.name)
				}
				return
			}
			fil.name = f.name
			extracSymErrs := fil.ExtractWithOptions(path, dereferenceSymlink, unbreakSymlink, folderPerm, verbose)
			if len(extracSymErrs) > 0 {
				if verbose {
					fmt.Println("Error(s) while extracting the symlink's file:", path+"/"+f.name)
					fmt.Println(extracSymErrs)
				}
				errs = append(errs, extracSymErrs...)
			}
			return
		} else if unbreakSymlink {
			fil := f.GetSymlinkFile()
			if fil != nil {
				symPath = path + "/" + symPath
				paths := strings.Split(symPath, "/")
				extracSymErrs := fil.ExtractWithOptions(strings.Join(paths[:len(paths)-1], "/"), dereferenceSymlink, unbreakSymlink, folderPerm, verbose)
				if len(extracSymErrs) > 0 {
					if verbose {
						fmt.Println("Error(s) while extracting the symlink's file:", path+"/"+f.name)
						fmt.Println(extracSymErrs)
					}
					errs = append(errs, extracSymErrs...)
				}
			} else {
				if verbose {
					fmt.Println("Symlink path(", symPath, ") is outside the archive:"+path+"/"+f.name)
				}
				return
			}
		}
		err = os.Symlink(f.SymlinkPath(), path+"/"+f.name)
		if err != nil {
			if verbose {
				fmt.Println("Error while making symlink:", path+"/"+f.name)
				fmt.Println(err)
			}
			errs = append(errs, err)
		}
	}
	return
}

//Read from the file. Doesn't do anything fancy, just pases it to the underlying io.Reader. If a directory, return io.EOF.
func (f *File) Read(p []byte) (int, error) {
	if !f.IsFile() {
		return 0, io.EOF
	}
	var err error
	if f.reader == nil && f.r != nil {
		f.reader, err = f.r.newFileReader(f.in)
		if err != nil {
			return 0, err
		}
	}
	return f.reader.Read(p)
}

//ReadDirFromInode returns a fully populated Directory from a given Inode.
//If the given inode is not a directory it returns an error.
func (r *Reader) readDirFromInode(i *inode.Inode) (*directory.Directory, error) {
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
	_, err = br.Seek(int64(en.Offset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	i, err := inode.ProcessInode(br, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	return i, nil
}
