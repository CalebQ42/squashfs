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
	*Header
	Name string
	EntryRaw
}

//NewEntry creates a new directory entry
func NewEntry(rdr io.Reader) (Entry, error) {
	var entry Entry
	err := binary.Read(rdr, binary.LittleEndian, &entry.EntryRaw)
	if err != nil {
		return Entry{}, err
	}
	tmp := make([]byte, entry.EntryRaw.NameSize+1)
	err = binary.Read(rdr, binary.LittleEndian, &tmp)
	if err != nil {
		return Entry{}, err
	}
	entry.Name = string(tmp)
	return entry, err
}

//Directory is an entry in the directory table of a squashfs.
//Will only have multiple headers if there are more then 256 entries
type Directory struct {
	Headers []Header
	Entries []Entry
}

//NewDirectory reads the directory from rdr
func NewDirectory(base io.Reader, size uint32) (*Directory, error) {
	var dir Directory
	var err error
	tmp := make([]byte, size)
	base.Read(tmp)
	rdr := bytes.NewBuffer(tmp)
	for {
		var hdr Header
		err = binary.Read(rdr, binary.LittleEndian, &hdr)
		if err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			return nil, err
		}
		hdr.Count++
		headers := hdr.Count / 256
		if hdr.Count%256 > 0 {
			headers++
		}
		dir.Headers = append(dir.Headers, hdr)
		for i := uint32(0); i < hdr.Count; i++ {
			if i != 0 && i%256 == 0 {
				var newHdr Header
				err = binary.Read(rdr, binary.LittleEndian, &newHdr)
				if err != nil {
					return nil, err
				}
				dir.Headers = append(dir.Headers, newHdr)
			}
			var ent Entry
			ent, err = NewEntry(rdr)
			if err != nil {
				return nil, err
			}
			ent.Header = &dir.Headers[len(dir.Headers)-1]
			dir.Entries = append(dir.Entries, ent)
		}
	}
	return &dir, nil
}
