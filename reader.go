package squashfs

import (
	"io"
	"time"

	"github.com/CalebQ42/squashfs/internal/toreader"
	squashfslow "github.com/CalebQ42/squashfs/low"
)

type Reader struct {
	FS
	Low squashfslow.Reader
}

func NewReader(r io.ReaderAt) (Reader, error) {
	rdr, err := squashfslow.NewReader(r)
	if err != nil {
		return Reader{}, err
	}
	out := Reader{
		Low: rdr,
	}
	out.FS = FS{
		LowDir: rdr.Root,
		r:   &out,
	}
	return out, nil
}

func NewReaderAtOffset(r io.ReaderAt, offset int64) (Reader, error) {
	return NewReader(toreader.NewOffsetReader(r, offset))
}

func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.Low.Superblock.ModTime), 0)
}
