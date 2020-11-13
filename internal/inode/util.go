package inode

//ProcessInodeRef processes an inode reference and returns two values
//The first value is the inode table offset. AKA, it's where the metadata block of the inode STARTS.
//The second value is the offset of the inode, INSIDE of the metadata.
func ProcessInodeRef(inodeRef uint64) (tableOffset uint32, metaOffset uint16) {
	tableOffset = uint32(inodeRef >> 16)
	metaOffset = uint16(inodeRef &^ 0xFFFFFFFF0000)
	return
}
