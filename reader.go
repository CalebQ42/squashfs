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
	out := &Reader{
		r: rdr,
	}
	out.FS = &FS{
		d: rdr.Root,
		r: out,
	}
	return out, nil
}

func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.r.Superblock.ModTime), 0)
}
