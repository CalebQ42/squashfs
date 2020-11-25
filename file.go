package squashfs

import (
	"io"

	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
)

//File represents a file within a squashfs. File can be either a file or folder.
type File struct {
	Name   string
	Parent *File
	Reader *io.Reader
	path   string
	size   uint32
	r      *Reader
	in     *inode.Inode
}

func (r *Reader) newFileFromEntry(en *directory.Entry) (f *File, err error) {
	f.Name = en.Name
	f.in, err = r.getInodeFromEntry(en)
}
