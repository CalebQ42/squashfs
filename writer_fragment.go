package squashfs

import (
	"bytes"
	"io"
)

type fragment struct {
	w     *Writer
	files []*fileHolder
	sizes []uint32
}

func (f *fragment) sizeLeft() uint32 {
	totalSize := uint32(0)
	for _, siz := range f.sizes {
		totalSize += siz
	}
	return f.w.BlockSize - uint32(totalSize)
}

func (f *fragment) addFragment(fil *fileHolder) {
	//SizeLeft should already be checked
	fil.fragOffset = len(f.files)
	f.files = append(f.files, fil)
	f.sizes = append(f.sizes, fil.blockSizes[len(fil.blockSizes)-1])
}

//TODO: give info about the frags for the frag table.
func (w *Writer) writeFragments(write io.WriterAt, off int64) (newOff int64, err error) {
	newOff = off
	var buf bytes.Buffer
	var byts []byte
	var n int
	for _, frag := range w.frags {
		blockOffset := 0
		for i, fil := range frag.files {
			_, err = io.CopyN(&buf, fil.reader, int64(frag.sizes[i]))
			if err != nil {
				return
			}
			fil.fragOffset = blockOffset
			blockOffset += int(frag.sizes[i])
		}
		if !w.Flags.UncompressedFragments && w.compressor != nil {
			byts, err = w.compressor.Compress(buf.Bytes())
			if err != nil {
				return
			}
		} else {
			byts = buf.Bytes()
		}
		n, err = write.WriteAt(byts, newOff)
		newOff += int64(n)
		if err != nil {
			return
		}
		buf.Reset()
	}
	return
}

func (w *Writer) addToFragments(fil *fileHolder) {
	fragSize := fil.blockSizes[len(fil.blockSizes)-1]
	//only fragment if the final block is less then 80% of a full block or AlwaysFragment.
	//TODO: Make this check after looking at all fragment blocks. Below option is better, but this is easier to implement...
	if w.Flags.AlwaysFragment || fragSize < uint32(float32(w.BlockSize)*0.8) {
		//Try to slot the fragment into a fragment that has the perfect size left. If not, just pick the first one.
		//TODO: possibly make this more efficient, possibly by calculating fragments all at once and seeing which combos match BlockSize perfectly.
		var possibleFrags []int
		for i := range w.frags {
			left := w.frags[i].sizeLeft()
			if left == fragSize {
				fil.fragIndex = i
				w.frags[i].addFragment(fil)
				return
			} else if left > fragSize {
				possibleFrags = append(possibleFrags, i)
			}
		}
		if len(possibleFrags) > 0 {
			fil.fragIndex = possibleFrags[0]
		} else {
			fil.fragIndex = len(w.frags)
			w.frags = append(w.frags, fragment{
				w:     w,
				files: []*fileHolder{fil},
				sizes: []uint32{fragSize},
			})
		}
	}
}
