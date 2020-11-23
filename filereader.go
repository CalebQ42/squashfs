package squashfs

type FileReader struct {
	r            *Reader
	data         *DataReader
	fragmentData []byte
}

//TODO: Yes

func (r *Reader) ReadFile(location string) (*FileReader, error) {

}

func (f *FileReader) Read(p []byte) (int, error) {

}
