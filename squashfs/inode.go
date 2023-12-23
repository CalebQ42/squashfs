package squashfs

import (
	"github.com/CalebQ42/squashfs/internal/metadata"
	"github.com/CalebQ42/squashfs/internal/toreader"
	"github.com/CalebQ42/squashfs/squashfs/inode"
)

func (r *Reader) inodeFromRef(ref uint64) (*inode.Inode, error) {
	offset, meta := (ref>>16)+r.sup.InodeTableStart, ref&0xFFFF
	rdr := metadata.NewReader(toreader.NewReader(r.r, int64(offset)), r.d)
	_, err := rdr.Read(make([]byte, meta))
	if err != nil {
		return nil, err
	}
	return inode.Read(rdr, r.sup.BlockSize)
}
