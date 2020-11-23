package squashfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/CalebQ42/GoSquashfs/internal/inode"
)

type FragmentEntry struct {
	Start  uint64
	Size   uint32
	Unused uint32
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
		fmt.Println(fragIndex, fragOffset, size)
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
	fmt.Println("fragment index", fragIndex)
	//reading the fragment entry first
	fragBlockIndex := int(fragIndex / 512)
	fragEntryRdr, err := r.NewBlockReader(int64(r.fragOffsets[fragBlockIndex]))
	if err != nil {
		return nil, err
	}
	_, err = fragEntryRdr.Seek(int64(16*fragIndex), io.SeekStart)
	if err != nil {
		return nil, err
	}
	var entry FragmentEntry
	err = binary.Read(fragEntryRdr, binary.LittleEndian, &entry)
	if err != nil {
		return nil, err
	}
	//now reading the actual fragment
	dr, err := r.NewDataReader(int64(entry.Start), []uint32{entry.Size})
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
