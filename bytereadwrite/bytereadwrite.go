package bytereadwrite

import "io"

//ByteReaderWriter allows you to read to and from a byte slice. When writing, it expands the slice to accomodate any data.
type ByteReaderWriter struct {
	byts   []byte
	offset int
}

//NewByteReaderWriter creates a ByteReaderWriter with an internal byte slice of the given length.
func NewByteReaderWriter() *ByteReaderWriter {
	return &ByteReaderWriter{
		byts:   make([]byte, 0),
		offset: 0,
	}
}

//NewByteReaderWriter creates a ByteReaderWriter with an internal byte slice of the given length.
func NewByteReaderWriterWithLength(length int) *ByteReaderWriter {
	return &ByteReaderWriter{
		byts:   make([]byte, length),
		offset: 0,
	}
}

//NewByteReaderWriterFromBytes creates a new ByteReaderWriter initialized with the given bytes
func NewByteReaderWriterFromBytes(byts []byte) *ByteReaderWriter {
	return &ByteReaderWriter{
		byts:   byts,
		offset: 0,
	}
}

//GetBytes return the underlyting byte slice of the readerwriter
func (bwr *ByteReaderWriter) GetBytes() []byte {
	return bwr.byts
}

//Read reads the bytes.
func (bwr *ByteReaderWriter) Read(byt []byte) (int, error) {
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

//Write writes to the end of the bytes. WILL expand to accept the incoming bytes.
func (bwr *ByteReaderWriter) Write(byts []byte) (int, error) {
	bwr.byts = append(bwr.byts, byts...)
	return len(byts), nil
}
