package squashfs

import "io"

//Squashfs is a squashfs backed by a reader.
type Squashfs struct {
	rdr   *io.Reader //underlyting reader
	super Superblock
}
