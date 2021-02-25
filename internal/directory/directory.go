package directory

import (
	"bytes"
	"encoding/binary"
	"io"
)

//Header is the header for a directory in the directory table
type Header struct {
	Count       uint32
	InodeOffset uint32
	InodeNumber uint32
}

//EntryRaw is the values that can be easily decoded
type EntryRaw struct {
	Offset      uint16
	InodeOffset int16
	Type        uint16
	NameSize    uint16
}

//Entry is an entry in a directory.
type Entry struct {
	Name             string
	InodeOffset      uint32
	InodeBlockOffset uint16
	Type             uint16
}

//NewEntry creates a new directory entry
func NewEntry(rdr io.Reader) (*Entry, error) {
	var raw EntryRaw
	err := binary.Read(rdr, binary.LittleEndian, &raw)
	if err != nil {
		return nil, err
	}
	tmp := make([]byte, raw.NameSize+1)
	err = binary.Read(rdr, binary.LittleEndian, &tmp)
	if err != nil {
		return nil, err
	}
	return &Entry{
		InodeBlockOffset: raw.Offset,
		Type:             raw.Type,
		Name:             string(tmp),
	}, nil
}

//NewDirectory reads the directory from rdr
func NewDirectory(base io.Reader, size uint32) (entries []*Entry, err error) {
	tmp := make([]byte, size)
	base.Read(tmp)
	rdr := bytes.NewBuffer(tmp)
	for {
		var hdr Header
		err = binary.Read(rdr, binary.LittleEndian, &hdr)
		if err == io.ErrUnexpectedEOF {
			err = nil
			break
		} else if err != nil {
			return nil, err
		}
		hdr.Count++
		headers := hdr.Count / 256
		if hdr.Count%256 > 0 {
			headers++
		}
		for i := uint32(0); i < hdr.Count; i++ {
			if i != 0 && i%256 == 0 {
				err = binary.Read(rdr, binary.LittleEndian, &hdr)
				if err != nil {
					return nil, err
				}
			}
			var ent *Entry
			ent, err = NewEntry(rdr)
			if err != nil {
				return nil, err
			}
			ent.InodeOffset = hdr.InodeOffset
			entries = append(entries, ent)
		}
	}
	return
}
