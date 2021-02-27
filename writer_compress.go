package squashfs

import (
	"io"
	"reflect"
	"sync"
)

func (w *Writer) compressData(data []byte) ([]byte, error) {
	if reflect.DeepEqual(data, make([]byte, len(data))) {
		return nil, nil
	}
	if w.Flags.UncompressedData || w.compressor == nil {
		return data, nil
	}
	compressedData, err := w.compressor.Compress(data)
	if err != nil {
		return nil, err
	}
	if len(data) <= len(compressedData) {
		return data, nil
	}
	return compressedData, nil
}

//Writes the given fileHolder to the WriterAt at the given offset.
//If fil.Reader implements io.ReaderAt, the process is threaded.
func (w *Writer) writeFile(fil *fileHolder, write io.WriterAt, startOffset int64) (endOffset int64, err error) {
	endOffset = startOffset
	var sizes []uint32
	if fil.fragIndex != -1 {
		sizes = fil.blockSizes[:len(fil.blockSizes)-1]
	} else {
		sizes = fil.blockSizes
	}
	if rdrAt, ok := fil.reader.(io.ReaderAt); ok {
		type writeReturn struct {
			err  error
			byts []byte
			i    int
		}
		out := make(chan *writeReturn)
		var filOffset int64
		var sync sync.WaitGroup
		sync.Add(len(sizes))
		for i, size := range sizes {
			go func(offset int64, size uint32, i int) {
				var ret writeReturn
				ret.i = i
				defer func() {
					out <- &ret
				}()
				ret.byts = make([]byte, size)
				_, ret.err = rdrAt.ReadAt(ret.byts, offset)
				if ret.err != nil {
					return
				}
				ret.byts, ret.err = w.compressData(ret.byts)
				sync.Done()
			}(filOffset, size, i)
			filOffset += int64(size)
		}
		var curInd int
		var holdingArea []*writeReturn
		for curInd < len(sizes) {
			var tmp *writeReturn
			for _, ret := range holdingArea {
				if ret.i == curInd {
					tmp = ret
					break
				}
			}
			if tmp == nil {
				tmp = <-out
				if tmp.err != nil {
					sync.Wait()
					return endOffset, tmp.err
				}
				if tmp.i != curInd {
					holdingArea = append(holdingArea, tmp)
					continue
				}
			}
			fil.blockSizes[curInd] = uint32(len(tmp.byts))
			if len(tmp.byts) == int(w.BlockSize) {
				//set uncompressed bit if not compressed
				fil.blockSizes[curInd] |= (1 << 24)
			}
			var n int
			n, err = write.WriteAt(tmp.byts, endOffset)
			endOffset += int64(n)
			if err != nil {
				sync.Wait()
				return
			}
			curInd++
		}
		return
	}
	var byts []byte
	for i, size := range sizes {
		byts = make([]byte, size)
		_, err = fil.reader.Read(byts)
		if err != nil {
			return
		}
		byts, err = w.compressData(byts)
		if err != nil {
			return
		}
		fil.blockSizes[i] = uint32(len(byts))
		if len(byts) == int(w.BlockSize) {
			//set uncompressed bit if not compressed
			fil.blockSizes[i] |= (1 << 24)
		}
		var n int
		n, err = write.WriteAt(byts, endOffset)
		endOffset += int64(n)
		if err != nil {
			return
		}
	}
	return
}
