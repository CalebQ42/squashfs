package squashfs

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
)

//FragmentEntry is an entry in the fragment table
type fragmentEntry struct {
	Start uint64
	Size  uint32
	// Unused uint32
}

//GetFragmentDataFromInode returns the fragment data for a given inode.
//If the inode does not have a fragment, harmlessly returns an empty slice without an error.
func (r *Reader) getFragmentDataFromInode(in *inode.Inode) ([]byte, error) {
	var size uint64
	var fragIndex uint32
	var fragOffset uint32
	if in.Type == inode.FileType {
		bf := in.Info.(inode.File)
		if !bf.Fragmented {
			return make([]byte, 0), nil
		}
		if bf.BlockStart == 0 {
			size = uint64(bf.Size)
		} else {
			size = uint64(bf.BlockSizes[len(bf.BlockSizes)-1])
		}
		fragIndex = bf.FragmentIndex
		fragOffset = bf.FragmentOffset
	} else if in.Type == inode.ExtFileType {
		bf := in.Info.(inode.ExtFile)
		if !bf.Fragmented {
			return make([]byte, 0), nil
		}
		if bf.BlockStart == 0 {
			size = bf.Size
		} else {
			size = uint64(bf.BlockSizes[len(bf.BlockSizes)-1])
		}
		fragIndex = bf.FragmentIndex
		fragOffset = bf.FragmentOffset
	} else {
		return nil, errors.New("Inode type not supported")
	}
	//reading the fragment entry first
	fragEntryRdr, err := r.newMetadataReader(int64(r.fragOffsets[int(fragIndex/512)]))
	if err != nil {
		return nil, err
	}
	_, err = fragEntryRdr.Seek(int64(16*fragIndex), io.SeekStart)
	if err != nil {
		return nil, err
	}
	var entry fragmentEntry
	err = binary.Read(fragEntryRdr, binary.LittleEndian, &entry)
	if err != nil {
		return nil, err
	}
	//now reading the actual fragment
	dr, err := r.newDataReader(int64(entry.Start), []uint32{entry.Size})
	if err != nil {
		return nil, err
	}
	_, err = dr.Read(make([]byte, fragOffset))
	if err != nil {
		return nil, err
	}
	tmp := make([]byte, size)
	err = binary.Read(dr, binary.LittleEndian, &tmp)
	if err != nil {
		return nil, err
	}
	return tmp, nil
}
