package directory

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

//Header is the header for a directory in the directory table
type Header struct {
	Count       uint32
	InodeOffset uint32
	InodeNumber uint32
}

//EntryInit is the values that can be easily decoded
type EntryInit struct {
	Offset      uint16
	InodeOffset int16
	Type        uint16
	NameSize    uint16
}

//Entry is an entry in a directory.
type Entry struct {
	Init EntryInit
	Name []byte
}

//NewEntry creates a new directory entry
func NewEntry(rdr io.Reader) (*Entry, error) {
	var entry Entry
	err := binary.Read(rdr, binary.LittleEndian, &entry.Init)
	if err != nil {
		return nil, err
	}
	entry.Name = make([]byte, entry.Init.NameSize+1)
	err = binary.Read(rdr, binary.LittleEndian, &entry.Name)
	if err != nil {
		return nil, err
	}
	return &entry, err
}

//Directory is an entry in the directory table of a squashfs.
//Will only have multiple headers if there are more then 256 entries
type Directory struct {
	Headers []Header
	Entries []*Entry
}

//NewDirectory reads the directory from rdr
func NewDirectory(rdr io.Reader) (*Directory, error) {
	var dir Directory
	var hdr Header
	err := binary.Read(rdr, binary.LittleEndian, &hdr)
	if err != nil {
		return nil, err
	}
	hdr.Count++
	headers := hdr.Count / 256
	if hdr.Count%256 > 0 {
		headers++
	}
	fmt.Println("headers", headers)
	fmt.Println("headers ceil", math.Ceil(float64(hdr.Count)/256))
	headersRead := 1
	dir.Headers = make([]Header, headers)
	dir.Headers[0] = hdr
	for i := uint32(0); i < hdr.Count; i++ {
		if i != 0 && i%256 == 0 {
			var newHdr Header
			err = binary.Read(rdr, binary.LittleEndian, &newHdr)
			if err != nil {
				return &dir, err
			}
			dir.Headers[headersRead] = newHdr
			headersRead++
		}
		ent, err := NewEntry(rdr)
		if err != nil {
			return &dir, err
		}
		dir.Entries = append(dir.Entries, ent)
	}
	return &dir, nil
}
