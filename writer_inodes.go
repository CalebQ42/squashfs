package squashfs

func (w *Writer) countInodes() (out uint32) {
	for _, files := range w.structure {
		out++
		out += uint32(len(files))
	}
	return
}
