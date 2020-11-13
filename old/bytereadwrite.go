package squashfs

import "io"

type byteWriterReader struct {
	byts   []byte
	offset int
}

func newByteReadWrite(length int) *byteWriterReader {
	return &byteWriterReader{
		byts:   make([]byte, length),
		offset: 0,
	}
}

func newByteReadWriteFromBytes(byts []byte) *byteWriterReader {
	return &byteWriterReader{
		byts:   byts,
		offset: 0,
	}
}

func (bwr *byteWriterReader) getBytes() []byte {
	return bwr.byts
}

//Read reads the bytes.
func (bwr *byteWriterReader) Read(byt []byte) (int, error) {
	if len(bwr.byts) < bwr.offset+len(byt) {
		bytesWritten := len(bwr.byts) - bwr.offset
		for i := 0; i < bytesWritten; i++ {
			byt[i] = bwr.byts[i+bwr.offset]
		}
		return bytesWritten, io.EOF
	}
	for i := 0; i < len(byt); i++ {
		byt[i] = bwr.byts[bwr.offset+i]
	}
	bwr.offset += len(byt)
	return len(byt), nil
}

//Write writes to the bytes. WILL expand to accept the incoming bytes.
func (bwr *byteWriterReader) Write(byts []byte) (int, error) {
	bwr.byts = append(bwr.byts, byts...)
	return len(byts), nil
}
