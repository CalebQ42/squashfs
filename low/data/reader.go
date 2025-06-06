package data

import "io"

type Reader struct {
	f         *FullReader
	curBlock  []byte
	nextIdx   uint32
	curOffset uint32
}

func NewReader(f *FullReader) (Reader, error) {
	dat, err := f.Block(0)
	if err != nil {
		return Reader{}, err
	}
	return Reader{
		f:         f,
		curBlock:  dat,
		nextIdx:   1,
		curOffset: 0,
	}, nil
}

func (d *Reader) Close() error {
	d.curBlock = nil
	return nil
}

func (d *Reader) advanceBlock() error {
	if d.nextIdx >= d.f.BlockNum() {
		d.curBlock = nil
		return io.EOF
	}
	var err error
	d.curBlock, err = d.f.Block(d.nextIdx)
	if err != nil {
		return err
	}
	d.nextIdx++
	d.curOffset = 0
	return nil
}

func (d *Reader) Read(buf []byte) (int, error) {
	totRed := 0
	toRead := 0
	var err error
	for totRed < len(buf) {
		if int(d.curOffset) >= len(d.curBlock) {
			err = d.advanceBlock()
			if err != nil {
				return totRed, err
			}
		}
		toRead = min(len(d.curBlock)-int(d.curOffset), len(buf)-totRed)
		copy(buf[totRed:], d.curBlock[d.curOffset:d.curOffset+uint32(toRead)])
	}
	return totRed, nil
}
