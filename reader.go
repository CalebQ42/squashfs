package squashfs

import (
	"io"
	"time"

	"github.com/CalebQ42/squashfs/squashfs"
)

type Reader struct {
	*FS
	r *squashfs.Reader
}

func NewReader(r io.ReaderAt) (*Reader, error) {
	rdr, err := squashfs.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &Reader{
		r: rdr,
		FS: &FS{
			d: rdr.Root,
		},
	}, nil
}

func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.r.Superblock.ModTime), 0)
}
