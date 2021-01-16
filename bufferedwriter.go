package squashfs

import (
	"io"
)

type bufferedWriter struct {
	w      io.Writer
	buffer []bufferedBytes
}

type bufferedBytes struct {
	data   []byte
	offset int
}
