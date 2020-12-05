package squashfs

import (
	"errors"
	"fmt"
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
	//ErrBrokenSymlink is returned when using ExtractWithOptions with the unbreakSymlink set to true, but the symlink's file cannot be extracted.
	ErrBrokenSymlink = errors.New("Extracted symlink is probably broken")
)

//File is the main way to interact with files within squashfs, or when putting files into a squashfs.
//File can be either a file or folder. When reading from a squashfs, it reads from the datablocks.
//When writing, this holds the information on WHERE the file will be placed inside the archive.
type File struct {
	Name    string       //The name of the file or folder. Root folder will not have a name ("")
	Parent  *File        //The parent directory. Should ALWAYS be a folder. If it's the root directory, will be nil
	Reader  io.Reader    //Underlying reader. When writing, will probably be an os.File. When reading this is kept nil UNTIL reading to save memory.
	path    string       //The path to the folder the File is located in.
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
			fil.path = f.Path()
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
	foldChil := make(chan []*File)
	errChan := make(chan error)
	for _, folds := range childFolders {
		go func(fil *File) {
			childs, err := fil.GetChildrenRecursively()
			errChan <- err
			foldChil <- childs
		}(folds)
	}
	for range childFolders {
		err = <-errChan
		if err != nil {
			return
		}
		children = append(children, <-foldChil...)
	}
	return
}

//Path returns the path of the file within the archive.
func (f *File) Path() string {
	if f.Name == "" {
		return f.path
	}
	return f.path + "/" + f.Name
}

//GetFileAtPath tries to return the File at the given path, relative to the file.
//Returns nil if called on something other then a folder, OR if the path goes oustide the archive.
//Allows * wildcards.
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
outer:
	for _, child := range children {
		if strings.Contains(split[0], "*") {
			wilds := strings.Split(split[0], "*")
			curIndex := 0
			for i, section := range wilds {
				ind := strings.Index(child.Name, section)
				if ind == -1 {
					continue outer
				}
			}
		} else if child.Name == split[0] {
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

//IsFile returns if the file is a file.
func (f *File) IsFile() bool {
	return f.filType == inode.BasicFileType || f.filType == inode.ExtFileType
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

//Permission returns the os.FileMode of the File. Sets mode bits for directories and symlinks.
func (f *File) Permission() os.FileMode {
	mode := os.FileMode(f.in.Header.Permissions)
	switch {
	case f.IsDir():
		mode = mode | os.ModeDir
	case f.IsSymlink():
		mode = mode | os.ModeSymlink
	}
	return mode
}

//ExtractTo extracts the file to the given path. This is the same as ExtractWithOptions(path, false, os.ModePerm, false).
//Will NOT try to keep symlinks valid, folders extracted will have the permissions set by the squashfs, but the folder to make path will have full permissions (777).
//
//Will try it's best to extract all files, and if any errors come up, they will be appended to the error slice that's returned.
func (f *File) ExtractTo(path string) []error {
	return f.ExtractWithOptions(path, false, os.ModePerm, false)
}

//ExtractWithOptions will extract the file to the given path, while allowing customization on how it works. ExtractTo is the "default" options.
//Will try it's best to extract all files, and if any errors come up, they will be appended to the error slice that's returned.
//Should only return multiple errors if extracting a folder.
//
//If unbreakSymlink is set, it will also try to extract the symlink's associated file. WARNING: the symlink's file may have to go up the directory to work.
//If unbreakSymlink is set and the file cannot be extracted, a ErrBrokenSymlink will be appended to the returned error slice.
//
//folderPerm only applies to the folders created to get to path. Folders from the archive are given the correct permissions defined by the archive.
func (f *File) ExtractWithOptions(path string, unbreakSymlink bool, folderPerm os.FileMode, verbose bool) (errs []error) {
	errs = make([]error, 0)
	err := os.MkdirAll(path, folderPerm)
	if err != nil {
		return []error{err}
	}
	switch {
	case f.IsDir():
		if f.Name != "" {
			//TODO: check if folder is present, and if so, try to set it's permission
			err = os.Mkdir(path+"/"+f.Name, os.ModePerm)
			if err != nil {
				if verbose {
					fmt.Println("Error while making: ", path+"/"+f.Name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
			fil, err := os.Open(path + "/" + f.Name)
			if err != nil {
				if verbose {
					fmt.Println("Error while opening:", path+"/"+f.Name)
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
			err = fil.Chmod(f.Permission())
			if err != nil {
				if verbose {
					fmt.Println("Error while changing owner:", path+"/"+f.Name)
					fmt.Println(err)
				}
				errs = append(errs, err)
			}
		}
		children, err := f.GetChildren()
		if err != nil {
			if verbose {
				fmt.Println("Error getting children for:", f.Path())
				fmt.Println(err)
			}
			errs = append(errs, err)
			return
		}
		finishChan := make(chan []error)
		defer close(finishChan)
		for _, child := range children {
			go func(child *File) {
				if f.Name == "" {
					finishChan <- child.ExtractWithOptions(path, unbreakSymlink, folderPerm, verbose)
				} else {
					finishChan <- child.ExtractWithOptions(path+"/"+f.Name, unbreakSymlink, folderPerm, verbose)
				}
			}(child)
		}
		for range children {
			errs = append(errs, (<-finishChan)...)
		}
		return
	case f.IsFile():
		fil, err := os.Create(path + "/" + f.Name)
		if os.IsExist(err) {
			err = os.Remove(path + "/" + f.Name)
			if err != nil {
				if verbose {
					fmt.Println("Error while making:", path+"/"+f.Name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
			fil, err = os.Create(path + "/" + f.Name)
			if err != nil {
				if verbose {
					fmt.Println("Error while making:", path+"/"+f.Name)
					fmt.Println(err)
				}
				errs = append(errs, err)
				return
			}
		} else if err != nil {
			if verbose {
				fmt.Println("Error while making:", path+"/"+f.Name)
				fmt.Println(err)
			}
			errs = append(errs, err)
			return
		} //Since we will be reading from the file
		_, err = io.Copy(fil, f)
		if err != nil {
			if verbose {
				fmt.Println("Error while Copying data to:", path+"/"+f.Name)
				fmt.Println(err)
			}
			errs = append(errs, err)
			return
		}
		f.Close()
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
		err = fil.Chmod(f.Permission())
		if err != nil {
			if verbose {
				fmt.Println("Error while setting permissions for:", path+"/"+f.Name)
				fmt.Println(err)
			}
			errs = append(errs, err)
		}
		return
	case f.IsSymlink():
		symPath := f.SymlinkPath()
		if unbreakSymlink {
			fil := f.GetSymlinkFile()
			if fil != nil {
				symPath = path + "/" + symPath
				paths := strings.Split(symPath, "/")
				extracSymErrs := fil.ExtractWithOptions(strings.Join(paths[:len(paths)-1], "/"), unbreakSymlink, folderPerm, verbose)
				if len(extracSymErrs) > 0 {
					if verbose {
						fmt.Println("Error(s) while extracting the symlink's file:", path+"/"+f.Name)
						fmt.Println(extracSymErrs)
					}
					errs = append(errs, extracSymErrs...)
				}
			} else if verbose {
				fmt.Println("Symlink path(", symPath, ") is outside the archive:"+path+"/"+f.Name)
			}
		}
		err = os.Symlink(f.SymlinkPath(), path+"/"+f.Name)
		if err != nil {
			if verbose {
				fmt.Println("Error while making symlink:", path+"/"+f.Name)
				fmt.Println(err)
			}
			errs = append(errs, err)
		}
	}
	return
}

//Close frees up the memory held up by the underlying reader. Should NOT be called when writing.
//When reading, Close is safe to use, but any subsequent Read calls resets to the beginning of the file.
func (f *File) Close() error {
	if f.IsDir() {
		return errNotFile
	}
	if f.Reader != nil {
		if closer, is := f.Reader.(io.Closer); is {
			closer.Close()
		}
		f.Reader = nil
	}
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

//ReadDirFromInode returns a fully populated Directory from a given Inode.
//If the given inode is not a directory it returns an error.
func (r *Reader) readDirFromInode(i *inode.Inode) (*directory.Directory, error) {
	var offset uint32
	var metaOffset uint16
	var size uint32
	switch i.Type {
	case inode.BasicDirectoryType:
		offset = i.Info.(inode.BasicDirectory).DirectoryIndex
		metaOffset = i.Info.(inode.BasicDirectory).DirectoryOffset
		size = uint32(i.Info.(inode.BasicDirectory).DirectorySize)
	case inode.ExtDirType:
		offset = i.Info.(inode.ExtendedDirectory).Init.DirectoryIndex
		metaOffset = i.Info.(inode.ExtendedDirectory).Init.DirectoryOffset
		size = i.Info.(inode.ExtendedDirectory).Init.DirectorySize
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
