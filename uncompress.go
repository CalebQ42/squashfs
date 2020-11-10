package squashfs

import "io"

func uncompressData(data []byte, compressionType int) []byte {
	//TODO: check compression type and uncompress the data
	return make([]byte, 0)
}

//same os uncompressData, but uses a reader instead. reader's seek will be
func uncompressReaderData(reader *io.Reader, compressionType int) []byte {
	//TODO: check compression type and uncompress the data
	return make([]byte, 0)
}
