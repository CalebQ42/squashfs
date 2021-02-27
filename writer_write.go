package squashfs

import (
	"errors"
	"io"
	"math"
	"time"
)

//WriteTo attempts to write the archive to the given io.Writer.
//
//Not working. Yet.
func (w *Writer) WriteTo(write io.Writer) (int64, error) {
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
	return 0, errors.New("I SAID DON'T")
}
