package squashfs

import (
	"errors"
	"io"
)

//WriteTo attempts to write the archive to the given io.Writer.
func (w *Writer) WriteTo(write io.Writer) (int64, error) {
	//TODO
	return 0, errors.New("I SAID DON'T")
}
