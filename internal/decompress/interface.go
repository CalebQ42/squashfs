package decompress

import (
	"io"
)

type Decompressor interface {
	//Creates a new decompressor reading from src.
	Reader(src io.Reader) (io.ReadCloser, error)
}

type Resetable interface {
	//Reset attempts to re-use an old decompressor with new data.
	//Will return ErrNotResetable if not Resetable().
	//Must ALWAYS be provided with a reader created with Reader.
	Reset(old, src io.Reader) error
}

type Decoder interface {
	//Decodes a chunk of data all at once.
	Decode(in []byte) ([]byte, error)
}
