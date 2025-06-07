package squashfslow

// TODO: Make work
// func requireNoError(t *testing.T, err error) {
// 	t.Helper()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func assertEqual(t *testing.T, want int, got int) {
// 	t.Helper()
// 	if want != got {
// 		t.Errorf("want %d, got %d", want, got)
// 	}
// }

// func assertLength(t *testing.T, want int, slice []int) {
// 	t.Helper()
// 	if len(slice) != want {
// 		t.Errorf("want len %d, got %d", want, len(slice))
// 	}
// }

// func assertErrorIs(t *testing.T, err error, wantErr error) {
// 	t.Helper()
// 	if err == nil {
// 		t.Errorf("want %s, got nil", wantErr)
// 		return
// 	}
// 	if !errors.Is(err, wantErr) {
// 		t.Errorf("want %s, got %v", wantErr, err)
// 	}
// }

// func TestCachingPagedReader(t *testing.T) {
// 	// Mock readBlocks function
// 	mockReadNMore := func(startBlock, numItems int) ([]int, error) {
// 		if startBlock < 0 {
// 			return nil, errors.New("invalid block start")
// 		}
// 		var result []int
// 		for i := 0; i < numItems; i++ {
// 			result = append(result, startBlock*512+i)
// 		}
// 		return result, nil
// 	}

// 	t.Run("ValidRequestWithinFirstBlock", func(t *testing.T) {
// 		tab := NewTable[int]()
// 		currentItems := make([]int, 0)
// 		item, err := readPagedItems(300, 512, &currentItems, 2048, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 300, item)
// 		assertLength(t, 512, currentItems) // Ensure one block is read
// 	})

// 	t.Run("ValidRequestAcrossMultipleBlocks", func(t *testing.T) {
// 		currentItems := make([]int, 0)
// 		item, err := readPagedItems(600, 512, &currentItems, 2048, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 600, item)
// 		assertLength(t, 1024, currentItems)
// 	})

// 	t.Run("SequentialRequestsWithinBlocks", func(t *testing.T) {
// 		currentItems := make([]int, 0)
// 		// First request
// 		item, err := readPagedItems(300, 512, &currentItems, 2048, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 300, item)

// 		// Second request in the same block
// 		item, err = readPagedItems(400, 512, &currentItems, 2048, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 400, item)
// 		assertLength(t, 512, currentItems)
// 	})

// 	t.Run("RequestExactBlockBoundary", func(t *testing.T) {
// 		currentItems := make([]int, 0)
// 		item, err := readPagedItems(511, 512, &currentItems, 2048, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 511, item)
// 		assertLength(t, 512, currentItems)

// 		// Request the next block's first item
// 		item, err = readPagedItems(512, 512, &currentItems, 2048, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 512, item)
// 		assertLength(t, 1024, currentItems)
// 	})

// 	t.Run("OutOfBoundsRequest", func(t *testing.T) {
// 		currentItems := make([]int, 0)
// 		_, err := readPagedItems(2048, 512, &currentItems, 2048, mockReadNMore)
// 		assertErrorIs(t, err, errOutOfBounds)
// 	})

// 	t.Run("RequestBeyondReadBlocks", func(t *testing.T) {
// 		readFail := errors.New("failed to read block")
// 		failingReadBlocks := func(startBlock, numBlocks int) ([]int, error) {
// 			if startBlock > 1 {
// 				return nil, readFail
// 			}
// 			var result []int
// 			for i := 0; i < numBlocks*512; i++ {
// 				result = append(result, startBlock*512+i)
// 			}
// 			return result, nil
// 		}

// 		currentItems := make([]int, 0)
// 		_, err := readPagedItems(1024, 512, &currentItems, 2048, failingReadBlocks)
// 		assertErrorIs(t, err, readFail)
// 	})

// 	t.Run("partial last page", func(t *testing.T) {
// 		currentItems := make([]int, 0)

// 		// Request the next block's first item
// 		item, err := readPagedItems(512, 512, &currentItems, 612, mockReadNMore)
// 		requireNoError(t, err)
// 		assertEqual(t, 512, item)
// 		assertLength(t, 612, currentItems)
// 	})
// }