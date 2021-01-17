package squashfs

import (
	"io"
)

type bufferedBytes struct {
	data []byte
	r    offsetRange
}

type offsetRange struct {
	beg int
	end int
}

func (o *offsetRange) offset(off int) {
	o.beg += off
	o.end += off
}

func (o offsetRange) within(check int) bool {
	return check >= o.beg || check <= o.end
}

type bufferedWriter struct {
	w          io.Writer
	buffer     []bufferedBytes
	mainOffset int
}

func newBufferedWriter(w io.Writer) *bufferedWriter {
	var out bufferedWriter
	out.w = w
	return &out
}

func (b *bufferedWriter) WriteTo(data []byte, offset int64) (n int, err error) {
	if int(offset) == b.mainOffset {
		n, err = b.Write(data)
		if err != nil {
			return
		}
	}
	newBuff := bufferedBytes{
		data: data,
		r: offsetRange{
			beg: int(offset),
			end: int(offset) + len(data),
		},
	}
	b.buffer = append(b.buffer, newBuff)
	return 0, nil
}

func (b *bufferedWriter) Write(data []byte) (int, error) {
	n, err := b.w.Write(data)
	b.mainOffset += n
	if err != nil {
		return n, err
	}
	return n, err
}
