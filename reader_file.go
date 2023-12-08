package squashfs

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/CalebQ42/squashfs/internal/data"
	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
	"github.com/CalebQ42/squashfs/internal/threadmanager"
)

// File represents a file inside a squashfs archive.
type File struct {
	i        inode.Inode
	rdr      io.Reader
	fullRdr  *data.FullReader
	r        *Reader
	parent   *FS
	e        directory.Entry
	dirsRead int
}

var (
	ErrReadNotFile = errors.New("read called on non-file")
)

func (r Reader) newFile(en directory.Entry, parent *FS) (*File, error) {
	i, err := r.inodeFromDir(en)
	if err != nil {
		return nil, err
	}
	var rdr io.Reader
	var full *data.FullReader
	if i.Type == inode.Fil || i.Type == inode.EFil {
		full, rdr, err = r.getReaders(i)
		if err != nil {
			return nil, err
		}
	}
	return &File{
		e:       en,
		i:       i,
		rdr:     rdr,
		fullRdr: full,
		r:       &r,
		parent:  parent,
	}, nil
}

// Stat returns the File's fs.FileInfo
func (f File) Stat() (fs.FileInfo, error) {
	return newFileInfo(f.e, f.i), nil
}

// Mode returns the file's fs.FileMode
func (f File) Mode() fs.FileMode {
	switch f.e.Type {
	case inode.Dir:
		return fs.FileMode(f.i.Perm) | fs.ModeDir
	case inode.Char:
		return fs.FileMode(f.i.Perm) | fs.ModeCharDevice
	case inode.Block:
		return fs.FileMode(f.i.Perm) | fs.ModeDevice
	case inode.Sym:
		return fs.FileMode(f.i.Perm) | fs.ModeSymlink
	}
	return fs.FileMode(f.i.Perm)
}

// Read reads the data from the file. Only works if file is a normal file.
func (f File) Read(p []byte) (int, error) {
	if f.i.Type != inode.Fil && f.i.Type != inode.EFil {
		return 0, ErrReadNotFile
	}
	if f.rdr == nil {
		return 0, fs.ErrClosed
	}
	return f.rdr.Read(p)
}

func (f File) ReadAt(p []byte, off int64) (int, error) {
	if f.i.Type != inode.Fil && f.i.Type != inode.EFil {
		return 0, ErrReadNotFile
	}
	return f.fullRdr.ReadAt(p, off)
}

// WriteTo writes all data from the file to the writer. This is multi-threaded.
// The underlying reader is seperate from the one used with Read and can be reused.
func (f File) WriteTo(w io.Writer) (int64, error) {
	if f.i.Type != inode.Fil && f.i.Type != inode.EFil {
		return 0, ErrReadNotFile
	}
	return f.fullRdr.WriteTo(w)
}

// Close simply nils the underlying reader.
func (f *File) Close() error {
	f.rdr = nil
	return nil
}

// ReadDir returns n fs.DirEntry's that's contained in the File (if it's a directory).
// If n <= 0 all fs.DirEntry's are returned.
func (f *File) ReadDir(n int) (out []fs.DirEntry, err error) {
	if !f.IsDir() {
		return nil, errors.New("file is not a directory")
	}
	ents, err := f.r.readDirectory(f.i)
	if err != nil {
		return nil, err
	}
	start, end := 0, len(ents)
	if n > 0 {
		start, end = f.dirsRead, f.dirsRead+n
		if end > len(f.r.e) {
			end = len(f.r.e)
			err = io.EOF
		}
	}
	var fi fileInfo
	for _, e := range ents[start:end] {
		fi, err = f.r.newFileInfo(e)
		if err != nil {
			f.dirsRead += len(out)
			return
		}
		out = append(out, fs.FileInfoToDirEntry(fi))
	}
	f.dirsRead += len(out)
	return
}

// FS returns the File as a FS.
func (f *File) FS() (*FS, error) {
	if !f.IsDir() {
		return nil, errors.New("File is not a directory")
	}
	ents, err := f.r.readDirectory(f.i)
	if err != nil {
		return nil, err
	}
	return &FS{
		File: f,
		e:    ents,
	}, nil
}

// IsDir Yep.
func (f File) IsDir() bool {
	return f.i.Type == inode.Dir || f.i.Type == inode.EDir
}

// IsRegular yep.
func (f File) IsRegular() bool {
	return f.i.Type == inode.Fil || f.i.Type == inode.EFil
}

// IsSymlink yep.
func (f File) IsSymlink() bool {
	return f.i.Type == inode.Sym || f.i.Type == inode.ESym
}

func (f File) isDeviceOrFifo() bool {
	return f.i.Type == inode.Char || f.i.Type == inode.Block || f.i.Type == inode.EChar || f.i.Type == inode.EBlock || f.i.Type == inode.Fifo || f.i.Type == inode.EFifo
}

func (f File) deviceDevices() (maj uint32, min uint32) {
	var dev uint32
	if f.i.Type == inode.Char || f.i.Type == inode.Block {
		dev = f.i.Data.(inode.Device).Dev
	} else if f.i.Type == inode.EChar || f.i.Type == inode.EBlock {
		dev = f.i.Data.(inode.EDevice).Dev
	}
	return dev >> 8, dev & 0x000FF
}

// SymlinkPath returns the symlink's target path. Is the File isn't a symlink, returns an empty string.
func (f File) SymlinkPath() string {
	switch f.i.Type {
	case inode.Sym:
		return string(f.i.Data.(inode.Symlink).Target)
	case inode.ESym:
		return string(f.i.Data.(inode.ESymlink).Target)
	}
	return ""
}

func (f File) path() string {
	if f.parent == nil {
		return f.e.Name
	}
	return f.parent.path() + "/" + f.e.Name
}

// GetSymlinkFile returns the File the symlink is pointing to.
// If not a symlink, or the target is unobtainable (such as it being outside the archive or it's absolute) returns nil
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

// ExtractionOptions are available options on how to extract.
type ExtractionOptions struct {
	manager            *threadmanager.Manager
	LogOutput          io.Writer   //Where error log should write.
	DereferenceSymlink bool        //Replace symlinks with the target file.
	UnbreakSymlink     bool        //Try to make sure symlinks remain unbroken when extracted, without changing the symlink.
	Verbose            bool        //Prints extra info to log on an error.
	IgnorePerm         bool        //Ignore file's permissions and instead use Perm.
	Perm               fs.FileMode //Permission to use when IgnorePerm. Defaults to 0755.
	notFirst           bool
}

func DefaultOptions() *ExtractionOptions {
	return &ExtractionOptions{
		Perm: 0755,
	}
}

// ExtractTo extracts the File to the given folder with the default options.
// If the File is a directory, it instead extracts the directory's contents to the folder.
func (f File) ExtractTo(folder string) error {
	return f.realExtract(folder, DefaultOptions())
}

// ExtractVerbose extracts the File to the folder with the Verbose option.
func (f File) ExtractVerbose(folder string) error {
	op := DefaultOptions()
	op.Verbose = true
	return f.realExtract(folder, op)
}

// ExtractIgnorePermissions extracts the File to the folder with the IgnorePerm option.
func (f File) ExtractIgnorePermissions(folder string) error {
	op := DefaultOptions()
	op.IgnorePerm = true
	return f.realExtract(folder, op)
}

// ExtractSymlink extracts the File to the folder with the DereferenceSymlink option.
// If the File is a directory, it instead extracts the directory's contents to the folder.
func (f File) ExtractSymlink(folder string) error {
	op := DefaultOptions()
	op.DereferenceSymlink = true
	return f.realExtract(folder, op)
}

// ExtractWithOptions extracts the File to the given folder with the given ExtrationOptions.
// If the File is a directory, it instead extracts the directory's contents to the folder.
func (f File) ExtractWithOptions(folder string, op *ExtractionOptions) error {
	if op.Verbose && op.LogOutput != nil {
		log.SetOutput(op.LogOutput)
	}
	return f.realExtract(folder, op)
}

func (f File) realExtract(folder string, op *ExtractionOptions) (err error) {
	if op.manager == nil {
		op.manager = threadmanager.NewManager(runtime.NumCPU())
	}
	extDir := filepath.Join(folder, f.e.Name)
	if !op.notFirst {
		op.notFirst = true
		if f.IsDir() {
			extDir = folder
			_, err = os.Open(folder)
			if err != nil && os.IsNotExist(err) {
				err = os.Mkdir(extDir, op.Perm)
			}
			if err != nil {
				if op.Verbose {
					log.Println("Error while making", folder)
				}
				return
			}
			if !op.IgnorePerm {
				defer os.Chmod(extDir, f.Mode())
				defer os.Chown(extDir, int(f.r.ids[f.i.UidInd]), int(f.r.ids[f.i.GidInd]))
			}
		}
	}
	switch {
	case f.IsDir():
		if folder != extDir && f.e.Name != "" {
			//First extract it with a permisive permission.
			err = os.Mkdir(extDir, op.Perm)
			if err != nil {
				if op.Verbose {
					log.Println("Error while making directory", extDir)
				}
				return
			}
			//Then set it to it's actual permissions once we're done with it
			if !op.IgnorePerm {
				defer os.Chmod(extDir, f.Mode())
				defer os.Chown(extDir, int(f.r.ids[f.i.UidInd]), int(f.r.ids[f.i.GidInd]))
			}
		}
		var filFS *FS
		filFS, err = f.FS()
		if err != nil {
			if op.Verbose {
				log.Println("Error while converting", f.path(), "to FS")
			}
			return err
		}
		errChan := make(chan error, len(filFS.e))
		files := make([]directory.Entry, 0)
		//Focus on making the folder tree first...
		var i int
		for i = 0; i < len(filFS.e); i++ {
			if filFS.e[i].Type == inode.Fil {
				files = append(files, filFS.e[i])
			} else {
				go func(index int) {
					subF, goErr := f.r.newFile(filFS.e[index], filFS)
					if goErr != nil {
						if op.Verbose {
							log.Println("Error while resolving", extDir)
						}
						errChan <- goErr
						return
					}
					errChan <- subF.ExtractWithOptions(extDir, op)
				}(i)
			}
		}
		for i = 0; i < len(filFS.e)-len(files); i++ {
			err = <-errChan
			if err != nil {
				return err
			}
		}
		//Then we extract the files.
		for i = 0; i < len(files); i++ {
			go func(index int) {
				n := op.manager.Lock()
				defer op.manager.Unlock(n)
				subF, goErr := f.r.newFile(files[index], filFS)
				if goErr != nil {
					if op.Verbose {
						log.Println("Error while resolving", extDir)
					}
					errChan <- goErr
					return
				}
				errChan <- subF.ExtractWithOptions(extDir, op)
			}(i)
		}
		for i = 0; i < len(files); i++ {
			err = <-errChan
			if err != nil {
				return err
			}
		}
	case f.IsRegular():
		var fil *os.File
		fil, err = os.Create(extDir)
		if os.IsExist(err) {
			os.Remove(extDir)
			fil, err = os.Create(extDir)
			if err != nil {
				if op.Verbose {
					log.Println("Error while creating", extDir)
				}
				return err
			}
		} else if err != nil {
			if op.Verbose {
				log.Println("Error while creating", extDir)
			}
			return err
		}
		defer fil.Close()
		_, err = io.Copy(fil, f)
		if err != nil {
			if op.Verbose {
				log.Println("Error while copying data to", extDir)
			}
			return err
		}
		if op.IgnorePerm {
			os.Chmod(extDir, op.Perm|(f.Mode()&fs.ModeType))
		} else {
			os.Chmod(extDir, f.Mode())
			os.Chown(extDir, int(f.r.ids[f.i.UidInd]), int(f.r.ids[f.i.GidInd]))
		}
	case f.IsSymlink():
		symPath := f.SymlinkPath()
		if op.DereferenceSymlink {
			fil := f.GetSymlinkFile()
			if fil == nil {
				if op.Verbose {
					log.Println("Symlink path(", symPath, ") is unobtainable:", extDir)
				}
				return errors.New("cannot get symlink target")
			}
			fil.e.Name = f.e.Name
			err = fil.realExtract(folder, op)
			if err != nil {
				if op.Verbose {
					log.Println("Error while extracting the symlink's file:", extDir)
				}
				return err
			}
			return nil
		} else if op.UnbreakSymlink {
			fil := f.GetSymlinkFile()
			if fil == nil {
				if op.Verbose {
					log.Println("Symlink path(", symPath, ") is unobtainable:", extDir)
				}
				return errors.New("cannot get symlink target")
			}
			extractLoc := filepath.Join(folder, filepath.Dir(symPath))
			err = fil.realExtract(extractLoc, op)
			if err != nil {
				if op.Verbose {
					log.Println("Error while extracting ", extDir)
				}
				return err
			}
		}
		err = os.Symlink(f.SymlinkPath(), extDir)
		if os.IsExist(err) {
			os.Remove(extDir)
			err = os.Symlink(f.SymlinkPath(), extDir)
		}
		if err != nil {
			if op.Verbose {
				log.Println("Error while making symlink:", extDir)
			}
			return err
		}
		if op.IgnorePerm {
			os.Chmod(extDir, op.Perm|(f.Mode()&fs.ModeType))
		} else {
			os.Chmod(extDir, f.Mode())
			os.Chown(extDir, int(f.r.ids[f.i.UidInd]), int(f.r.ids[f.i.GidInd]))
		}
	case f.isDeviceOrFifo():
		if runtime.GOOS == "windows" {
			if op.Verbose {
				log.Println(extDir, "ignored since it's a device link and can't be created on Windows.")
			}
			return nil
		}
		_, err = exec.LookPath("mknod")
		if err != nil {
			if op.Verbose {
				log.Println("Extracting Fifo IPC or Device and mknod is not in PATH")
			}
			return err
		}
		var typ string
		if f.i.Type == inode.Char || f.i.Type == inode.EChar {
			typ = "c"
		} else if f.i.Type == inode.Block || f.i.Type == inode.EBlock {
			typ = "b"
		} else { //Fifo IPC
			if runtime.GOOS == "darwin" {
				if op.Verbose {
					log.Println(extDir, "ignored since it's a Fifo file and can't be created on Darwin.")
				}
				return nil
			}
			typ = "p"
		}
		cmd := exec.Command("mknod", extDir, typ)
		if typ != "p" {
			maj, min := f.deviceDevices()
			cmd.Args = append(cmd.Args, strconv.Itoa(int(maj)), strconv.Itoa(int(min)))
		}
		if op.Verbose {
			cmd.Stdout = op.LogOutput
			cmd.Stderr = op.LogOutput
		}
		err = cmd.Run()
		if err != nil {
			if op.Verbose {
				log.Println("Error while running mknod for", extDir)
			}
			return err
		}
		if op.IgnorePerm {
			os.Chmod(extDir, op.Perm|(f.Mode()&fs.ModeType))
		} else {
			os.Chmod(extDir, f.Mode())
			os.Chown(extDir, int(f.r.ids[f.i.UidInd]), int(f.r.ids[f.i.GidInd]))
		}
	case f.e.Type == inode.Sock:
		if op.Verbose {
			log.Println(extDir, "ignored since it's a socket file.")
		}
	default:
		return errors.New("Unsupported file type. Inode type: " + strconv.Itoa(int(f.i.Type)))
	}
	return nil
}
