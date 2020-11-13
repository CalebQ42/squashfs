package inode

//ProcessInodeRef processes an inode reference and returns two values
//The first value is the inode table offset. AKA, it's where the metadata block of the inode STARTS.
//The second value is the offset of the inode, INSIDE of the metadata.
func ProcessInodeRef(inodeRef uint64) (tableOffset uint64, metaOffset uint64) {
	tableOffset = inodeRef >> 16
	metaOffset = inodeRef &^ 0xFFFFFFFF0000
	return
}
