package squashfs

func (w *Writer) countInodes() (out uint32) {
	for _, files := range w.structure {
		out++
		out += uint32(len(files))
	}
	return
}

//intilialize the block sizes. These values will be overwritten with their compressed sizes later.
func (w *Writer) calculateBlockSizes(fil *fileHolder) {
	tmp := fil.size
	for {
		if tmp < uint64(w.BlockSize) {
			fil.blockSizes = append(fil.blockSizes, uint32(tmp))
			break
		}
		tmp -= uint64(w.BlockSize)
		fil.blockSizes = append(fil.blockSizes, w.BlockSize)
	}
	return
}
