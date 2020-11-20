package squashfs

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

type FragmentEntryRaw struct {
	Start   uint64
	Size    uint32
	_unused uint32
}

type FragmentEntry struct {
	start      uint64
	size       uint32
	compressed bool
}

//NewFragmentEntry reads a fragment entry from the given io.Reader.
func (r *Reader) NewFragmentEntry(rdr io.Reader) (*FragmentEntry, error) {
	var entry FragmentEntry
	var raw FragmentEntryRaw
	err := binary.Read(rdr, binary.LittleEndian, &raw)
	if err != nil {
		return nil, err
	}
	entry.start = raw.Start
	entry.compressed = raw.Size&0x1000000 == 0x1000000
	entry.size = raw.Size &^ 0x1000000
	return &entry, nil
}

//GetFragmentDataFromInode returns the fragment data for a given inode.
//If the inode does not have a fragment, harmlessly returns an empty slice without an error.
func (r *Reader) GetFragmentDataFromInode(in *inode.Inode) ([]byte, error) {
	var size uint32
	var fragIndex uint32
	var fragOffset uint32
	if in.Type == inode.BasicFileType {
		bf := in.Info.(inode.BasicFile)
		if !bf.Fragmented {
			return make([]byte, 0), nil
		}
		if bf.Init.BlockStart == 0 {
			size = bf.Init.Size
		} else {
			size = bf.BlockSizes[len(bf.BlockSizes)-1]
		}
		fragIndex = bf.Init.FragmentIndex
		fragOffset = bf.Init.FragmentOffset
	} else if in.Type == inode.ExtFileType {
		bf := in.Info.(inode.ExtendedFile)
		if !bf.Fragmented {
			return make([]byte, 0), nil
		}
		if bf.Init.BlockStart == 0 {
			size = bf.Init.Size
		} else {
			size = bf.BlockSizes[len(bf.BlockSizes)-1]
		}
		fragIndex = bf.Init.FragmentIndex
		fragOffset = bf.Init.FragmentOffset
	} else {
		return nil, errors.New("Inode type not supported")
	}
	frag := r.fragEntries[fragIndex]
	datRdr, err := r.NewDataReader(int64(frag.start), []uint32{frag.size})
	if err != nil {
		return nil, err
	}
	_, err = datRdr.Seek(int64(fragOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	tmp := make([]byte, size)
	_, err = datRdr.Read(tmp)
	if err != nil {
		return nil, err
	}
	return tmp, nil
}
