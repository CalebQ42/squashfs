package squashfs

import (
	"errors"
	"io"
	"math"
	"os"
	"time"
)

//WriteToFilename creates the squashfs archive with the given filepath.
func (w *Writer) WriteToFilename(filepath string) error {
	newFil, err := os.Create(filepath)
	if err != nil {
		return err
	}
	_, err = w.WriteTo(newFil)
	return err
}

//WriteTo attempts to write the archive to the given io.WriterAt.
//
//Not working. Yet.
func (w *Writer) WriteTo(write io.WriterAt) (int64, error) {
	if w.BlockSize > 1048576 {
		w.BlockSize = 1048576
	} else if w.BlockSize < 4096 {
		w.BlockSize = 4096
	}
	w.Flags.RemoveDuplicates = false
	w.Flags.Exportable = false
	w.Flags.NoXattr = true
	w.calculateFragsAndBlockSizes()
	w.superblock = superblock{
		Magic:           magic,
		InodeCount:      w.countInodes(),
		CreationTime:    uint32(time.Now().Unix()),
		BlockSize:       w.BlockSize,
		CompressionType: uint16(w.compressionType),
		BlockLog:        uint16(math.Log2(float64(w.BlockSize))),
		Flags:           w.Flags.ToUint(),
		IDCount:         uint16(len(w.uidGUIDTable)),
		MajorVersion:    4,
		MinorVersion:    0,
	}
	w.dataOffset = 96 //superblock size
	//write compression options
	//write/calculate compressed data sizes

	return 0, errors.New("i said don't")
}

//splits up the size of files into
func (w *Writer) calculateFragsAndBlockSizes() {
	for _, files := range w.structure {
		for i := range files {
			files[i].fragIndex = -1
			files[i].blockSizes = make([]uint32, files[i].size/uint64(w.BlockSize))
			for j := range files[i].blockSizes {
				files[i].blockSizes[j] = w.BlockSize
			}
			fragSize := uint32(files[i].size % uint64(w.BlockSize))
			if fragSize > 0 {
				files[i].blockSizes = append(files[i].blockSizes, fragSize)
				if !w.Flags.NoFragments {
					w.addToFragments(files[i])
				}
			}
		}
	}
}
