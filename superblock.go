package squashfs

//Descriptions provided by https://dr-emann.github.io/squashfs/
type superblock struct {
	Magic               uint32 //Magic will be 0x73717368 if it's a legit Squashfs filesystem
	Inodes              uint32 //Inodes is the number of inodes in the inodes table
	MkfsTime            int32  //MkfsTime is when archive was created
	BlockSize           uint32 //BlockSize is the size of data blocks in bytes
	Fragments           uint32 //Fragments is the number of entries in fragment table
	Compression         uint16 //Compression is what type of compression is used
	BlockLog            uint16 //BlockLog should be log base 2 of BlockSize. If not then the squash might be corrupt
	Flags               uint16 //Flags are the superblock's flags
	IDCount             uint16 //IDCount is the number of IDs in the id lookup table
	Major               uint16
	Minor               uint16
	RootInode           Inode
	BytesUsed           int64
	IDTableStart        int64
	XattrIDTableStart   int64
	InodeTableStart     int64
	DirectoryTableStart int64
	FragmentTableStart  int64
	LookupTableStart    int64
}
