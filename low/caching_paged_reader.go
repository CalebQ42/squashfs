package squashfslow

import (
	"errors"
	"math"
)

var errOutOfBounds = errors.New("out of bounds")
var errUnexpectedOutOfBounds = errors.New("unexpected out of bounds")
var errNilCollection = errors.New("nil collection")

// readPagedItems calls readBLockOrPartial the correct number of times to cache
// requestedItemIndex in currentItems, and then returns currentItems[requestedItemIndex].
// Parameters:
// - requestedItemIndex: The index of the item to be retrieved.
// - blockSize: The number of items per block.
// - currentItems: A slice of already-read items to manage in-memory storage. Must not be nil.
// - readBlockOrPartial: A callback function that reads the next block. It takes the index of the block
// to be read, and the number of items to read. It is normally passed block size, but if the last
// block is incomplete, it will be passed the number of items in the last block.
// Returns:
//   - the T at requestedItemIndex
//   - a non-nil error and the zero value of T if an error was encountered.
func readPagedItems[T any](
	requestedItemIndex int,
	blockSize int,
	currentItems *[]T,
	totalItems int,
	readBlockOrPartial func(idxBlock, numItems int) ([]T, error),
) (T, error) {
	var zero T // Zero value for the item type, used for default return in error cases.
	if currentItems == nil {
		return zero, errNilCollection
	}

	if requestedItemIndex < 0 || requestedItemIndex >= totalItems {
		return zero, errOutOfBounds
	}

	if len(*currentItems) > requestedItemIndex {
		return (*currentItems)[requestedItemIndex], nil
	}

	// Calculate which block contains the requested item
	blockNum := int(math.Ceil(float64(requestedItemIndex+1)/float64(blockSize))) - 1

	// Calculate blocks to read
	blocksRead := len(*currentItems) / blockSize
	blocksToRead := blockNum - blocksRead + 1

	// Read and append new blocks
	for i := 0; i < blocksToRead; i++ {
		startBlock := blocksRead + i
		itemsLeft := totalItems - len(*currentItems)
		itemsToRead := blockSize
		if itemsToRead > itemsLeft {
			itemsToRead = itemsLeft
		}
		items, err := readBlockOrPartial(startBlock, itemsToRead)
		if err != nil {
			return zero, err
		}
		*currentItems = append(*currentItems, items...)
	}

	// Ensure the slice contains the requested index after reading
	if len(*currentItems) <= requestedItemIndex {
		return zero, errUnexpectedOutOfBounds
	}

	return (*currentItems)[requestedItemIndex], nil
}
