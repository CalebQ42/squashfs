package squashfs

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/inode"
)

func (w *Writer) countInodes() (out uint32) {
	out++ //for the root directory
	for _, files := range w.structure {
		out += uint32(len(files))
	}
	return
}

func (w *Writer) setupInodes() (size int, err error) {
	w.rootInode.Type = inode.DirType
	//setup
	size += binary.Size(w.rootInode)
	return
}

func (w *Writer) writeInodeTable(wrt io.WriterAt, off int64) (newOff int64, err error) {
	newOff = off
	return
}
