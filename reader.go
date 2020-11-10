package squashfs

import "io"

//TODO: possible custom reader because I'm havng some issuse...

type Reader struct {
	rdr    *io.SectionReader
	Offset int //Offset is the current offset of the reader
}

func NewReader(io.ReaderAt)
