package squashfs

type fragEntry struct {
	start uint64
	size  uint32
	_     uint32
}
