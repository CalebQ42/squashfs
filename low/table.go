package squashfslow

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/CalebQ42/squashfs/internal/metadata"
	"github.com/CalebQ42/squashfs/internal/toreader"
)

var errOutOfBounds = errors.New("out of bounds")
var errUnexpectedOutOfBounds = errors.New("unexpected out of bounds")
var errNilCollection = errors.New("nil collection")

type CreateFunction[T any] = func(io.Reader) (T, error)

type Table[T any] struct {
	totalItems    uint32
	itemsPerBlock uint32
	offset        uint64
	mut           sync.RWMutex
	currentItems  []T
	rdr           *Reader
	createFunc    CreateFunction[T]
}

func NewTable[T any](rdr *Reader, start uint64, totalItems uint32, createFunc CreateFunction[T]) *Table[T] {
	var zero T
	return &Table[T]{
		totalItems:    totalItems,
		itemsPerBlock: 8192 / uint32(binary.Size(zero)),
		offset:        start,
		mut:           sync.RWMutex{},
		rdr:           rdr,
		createFunc:    createFunc,
	}
}

func (t *Table[T]) Get(requestedItemIndex uint32) (T, error) {
	t.mut.RLock()
	if requestedItemIndex >= t.totalItems {
		t.mut.RUnlock()
		var zero T
		return zero, errOutOfBounds
	}
	if uint32(len(t.currentItems)) > requestedItemIndex {
		t.mut.RUnlock()
		return t.currentItems[requestedItemIndex], nil
	}
	t.mut.RUnlock()
	return t.fillAndGet(requestedItemIndex)
}

func (t *Table[T]) fillAndGet(requestedItemIndex uint32) (T, error) {
	t.mut.Lock()
	defer t.mut.Unlock()
	var offset uint64
	var toRead uint32
	var rdr *toreader.Reader
	var metaRdr metadata.Reader
	var err error
	for uint32(len(t.currentItems)) <= requestedItemIndex {
		rdr = toreader.NewReader(t.rdr.r, int64(t.offset))
		err = binary.Read(rdr, binary.LittleEndian, &offset)
		if err != nil {
			var zero T
			return zero, err
		}
		t.offset += 8
		toRead = min(t.itemsPerBlock, t.totalItems-uint32(len(t.currentItems)))
		oldLen := uint32(len(t.currentItems))
		t.currentItems = append(t.currentItems, make([]T, toRead)...)
		metaRdr = metadata.NewReader(toreader.NewReader(t.rdr.r, int64(offset)), t.rdr.d)
		for i := range toRead {
			t.currentItems[oldLen+i], err = t.createFunc(&metaRdr)
			if err != nil {
				var zero T
				return zero, err
			}
		}
	}
	return t.currentItems[requestedItemIndex], nil
}
