package squashfs

import (
	"errors"
	"io"
	"math"
	"time"
)

func (w Writer) countInodes() (out uint32) {
	out++ // for the root indode
	for _, fold := range w.folders {
		out += uint32(len(w.structure[fold]))
	}
	return
}

//WriteTo attempts to write the archive to the given io.Writer.
func (w *Writer) WriteTo(write io.WriteSeeker) (int64, error) {
	if w.BlockSize > 1048576 {
		w.BlockSize = 1048576
	} else if w.BlockSize < 4096 {
		w.BlockSize = 4096
	}
	w.Flags.Duplicates = false
	w.Flags.Exportable = false
	w.Flags.NoXattr = true
	//TODO: set forced Flag values
	//TODO: make sure we aren't missing folders
	super := superblock{
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
	_ = super
	return 0, errors.New("I SAID DON'T")
}
