package data

import (
	"bytes"
	"io"
	"testing"
)

func TestNewReaderEmptyFullReader(t *testing.T) {
	full := NewFullReader(bytes.NewReader(nil), nil, 4096, 0, 0, nil)

	reader, err := NewReader(&full)
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}

	got, err := io.ReadAll(&reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ReadAll() returned %d bytes, want 0", len(got))
	}
}

func TestNewReaderPropagatesFirstBlockError(t *testing.T) {
	tests := []struct {
		name  string
		sizes []uint32
		want  string
	}{
		{name: "missing block", want: "invalid block index"},
		{name: "truncated block", sizes: []uint32{1}, want: io.EOF.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			full := NewFullReader(bytes.NewReader(nil), nil, 4096, 1, 0, tt.sizes)

			_, err := NewReader(&full)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("NewReader() error = %v, want %v", err, tt.want)
			}
		})
	}
}
