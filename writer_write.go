package squashfs

import (
	"errors"
	"io"
	"math"
	"path"
)

func (w *Writer) fixFolders() error {
	for folder := range w.structure {
		if folder == "/" {
			continue
		}
		dir, name := path.Dir(folder), path.Base(folder)
	}
}

//WriteTo attempts to write the archive to the given io.Writer.
func (w *Writer) WriteTo(write io.Writer) (int64, error) {
	if w.BlockSize > 1048576 {
		w.BlockSize = 1048576
	} else if w.BlockSize < 4096 {
		w.BlockSize = 4096
	}
	w.Flags.RemoveDuplicates = false
	w.Flags.Exportable = false
	w.Flags.NoXattr = true
	w.superblock = superblock{
		Magic:           magic,
		BlockSize:       w.BlockSize,
		BlockLog:        uint16(math.Log2(float64(w.BlockSize))),
		CompressionType: uint16(w.compressionType),
		Flags:           w.Flags.ToUint(),
		IDCount:         uint16(len(w.uidGUIDTable)),
		MajorVersion:    4,
		MinorVersion:    0,
	}
	return 0, errors.New("I SAID DON'T")
}
