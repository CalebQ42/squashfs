package squashfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

const (
	magic = 0x73717368
)

var (
	//ErrNoMagic is returned if the magic number in the superblock isn't correct.
	ErrNoMagic = errors.New("Magic number doesn't match. Either isn't a squashfs or corrupted")
	//ErrIncompatibleCompression is returned if the compression type in the superblock doesn't work.
	ErrIncompatibleCompression = errors.New("Compression type unsupported")
	//ErrCompressorOptions is returned if compressor options is present. It's not currently supported.
	ErrCompressorOptions = errors.New("Compressor options is not currently supported")
	//ErrFragmentTableIssues is returned if there's trouble reading the fragment table when creating a reader.
	//When this is returned, the reader is still returned.
	ErrFragmentTableIssues = errors.New("Trouble while reading the fragment table")
)

//Reader processes and reads a squashfs archive.
//TODO: Give a way to actually read files :P
type Reader struct {
	r            io.ReaderAt
	super        Superblock
	flags        SuperblockFlags
	decompressor Decompressor
	fragOffsets  []uint64
}

//NewSquashfsReader returns a new squashfs.Reader from an io.ReaderAt
func NewSquashfsReader(r io.ReaderAt) (*Reader, error) {
	var rdr Reader
	rdr.r = r
	err := binary.Read(io.NewSectionReader(rdr.r, 0, int64(binary.Size(rdr.super))), binary.LittleEndian, &rdr.super)
	if err != nil {
		return nil, err
	}
	if rdr.super.Magic != magic {
		return nil, ErrNoMagic
	}
	rdr.flags = rdr.super.GetFlags()
	switch rdr.super.CompressionType {
	case gzipCompression:
		rdr.decompressor = &ZlibDecompressor{}
	default:
		return nil, ErrIncompatibleCompression
	}
	if rdr.flags.CompressorOptions {
		//TODO: parse compressor options
		return nil, ErrCompressorOptions
	}
	fragBlocks := int(math.Ceil(float64(rdr.super.FragCount) / 512.0))
	if fragBlocks > 0 {
		offset := int64(rdr.super.FragTableStart)
		for i := 0; i < fragBlocks; i++ {
			tmp := make([]byte, 8)
			_, err = r.ReadAt(tmp, offset)
			if err != nil {
				fmt.Println("Error while reading fragment block offsets")
				return nil, err
			}
			rdr.fragOffsets = append(rdr.fragOffsets, binary.LittleEndian.Uint64(tmp))
			offset += 8
		}
	}
	return &rdr, nil
}

//GetFilesList returns a list of ALL files in the squashfs, going down every folder.
//Paths that terminate in a folder end with /
func (r *Reader) GetFilesList() ([]string, error) {
	inoderdr, err := r.NewMetadataReaderFromInodeRef(r.super.RootInodeRef)
	if err != nil {
		return nil, err
	}
	i, err := inode.ProcessInode(inoderdr, r.super.BlockSize)
	if err != nil {
		return nil, err
	}
	paths, err := r.readDir(i)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

//readDir returns a list of all decendents of a given inode. Inode given MUST be a directory type.
func (r *Reader) readDir(i *inode.Inode) (paths []string, err error) {
	dir, err := r.ReadDirFromInode(i)
	if err != nil {
		return
	}
	for _, entry := range dir.Entries {
		if entry.Init.Type == inode.BasicFileType {
			in, _ := r.GetInodeFromEntry(&entry)
			fil := in.Info.(inode.BasicFile)
			fmt.Println("name:", entry.Name)
			fmt.Println("frag index:", fil.Init.FragmentIndex, "frag offset:", fil.Init.FragmentOffset)
		}
		if entry.Init.Type == inode.BasicDirectoryType {
			paths = append(paths)
			i, err = r.GetInodeFromEntry(&entry)
			if err != nil {
				return
			}
			var subPaths []string
			subPaths, err = r.readDir(i)
			if err != nil {
				return
			}
			for pathI := range subPaths {
				subPaths[pathI] = entry.Name + "/" + subPaths[pathI]
			}
			paths = append(paths, entry.Name+"/")
			paths = append(paths, subPaths...)
		} else {
			paths = append(paths, entry.Name)
		}
	}
	return
}

//GetFileStructure returns ALL folders and files contained in the squashfs. Folders end with a "/".
func (r *Reader) GetFileStructure() ([]string, error) {
	in, err := r.GetInodeFromPath("")
	if err != nil {
		return nil, err
	}
	return r.readDir(in)
}
