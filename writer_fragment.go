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
	f.files = append(f.files, fil)
	f.sizes = append(f.sizes, fil.blockSizes[len(fil.blockSizes)-1])
}
