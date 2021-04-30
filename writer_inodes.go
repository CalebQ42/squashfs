package squashfs

import (
	"encoding/binary"
	"io"
)

func (w *Writer) countInodes() (out uint32) {
	out++ //for the root directory
	for _, files := range w.structure {
		out += uint32(len(files))
	}
	return
}

func (w *Writer) calculateInodeTableSize() (out int, err error) {
	for _, files := range w.structure {
		for i := range files {
			_ = i
			//set up each file's inode and add it's binary.size to out
			out += binary.Size(files[i].inode)
		}
	}
	return
}

func (w *Writer) writeInodeTable(wrt io.WriterAt, off int64) (newOff int64, err error) {
	newOff = off
	return
}
