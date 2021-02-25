package squashfs

import "reflect"

func (w *Writer) compressData(data []byte) ([]byte, error) {
	if reflect.DeepEqual(data, make([]byte, len(data))) {
		return nil, nil
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
