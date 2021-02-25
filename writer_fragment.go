package squashfs

type fragment struct {
	w     *Writer
	files []*fileHolder
	sizes []uint32
}

func (f *fragment) SizeLeft() uint32 {
	totalSize := uint32(0)
	for _, siz := range f.sizes {
		totalSize += siz
	}
	return f.w.BlockSize - uint32(totalSize)
}

func (f *fragment) AddFragment(fil *fileHolder) {
	//SizeLeft should already be checked
	fil.fragOffset = len(f.files)
	f.files = append(f.files, fil)
	f.sizes = append(f.sizes, fil.blockSizes[len(fil.blockSizes)-1])
}

func (w *Writer) addToFragments(fil *fileHolder) {
	fragSize := fil.blockSizes[len(fil.blockSizes)-1]
	//only fragment if the final block is less then 80% of a full block or AlwaysFragment
	if w.Flags.AlwaysFragment || fragSize < uint32(float32(w.BlockSize)*0.8) {
		var possibleFrags []int
		for i := range w.frags {
			left := w.frags[i].SizeLeft()
			if left == fragSize {
				fil.fragIndex = i
				w.frags[i].AddFragment(fil)
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

func (w *Writer) calculateFragsAndBlockSizes() {
	for _, files := range w.structure {
		for i := range files {
			files[i].fragIndex = -1
			files[i].blockSizes = make([]uint32, files[i].size/uint64(w.BlockSize))
			for j := range files[i].blockSizes {
				files[i].blockSizes[j] = w.BlockSize
			}
			fragSize := uint32(files[i].size % uint64(w.BlockSize))
			if fragSize > 0 {
				files[i].blockSizes = append(files[i].blockSizes, fragSize)
				if !w.Flags.NoFragments {
					w.addToFragments(files[i])
				}
			}
		}
	}
}
