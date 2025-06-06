package inode

import (
	"encoding/binary"
	"io"
)

type Device struct {
	LinkCount uint32
	Dev       uint32
}

func ReadDevice(r io.Reader) (d Device, err error) {
	dat := make([]byte, 8)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	d.LinkCount = binary.LittleEndian.Uint32(dat)
	d.Dev = binary.LittleEndian.Uint32(dat[4:])
	return
}

type EDevice struct {
	Device
	XattrInd uint32
}

func ReadEDevice(r io.Reader) (d EDevice, err error) {
	dat := make([]byte, 12)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	d.LinkCount = binary.LittleEndian.Uint32(dat)
	d.Dev = binary.LittleEndian.Uint32(dat[4:])
	d.XattrInd = binary.LittleEndian.Uint32(dat[8:])
	return
}

type IPC struct {
	LinkCount uint32
}

func ReadIPC(r io.Reader) (i IPC, err error) {
	dat := make([]byte, 4)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	i.LinkCount = binary.LittleEndian.Uint32(dat)
	return
}

type EIPC struct {
	IPC
	XattrInd uint32
}

func ReadEIPC(r io.Reader) (i EIPC, err error) {
	dat := make([]byte, 8)
	_, err = r.Read(dat)
	if err != nil {
		return
	}
	i.LinkCount = binary.LittleEndian.Uint32(dat)
	i.XattrInd = binary.LittleEndian.Uint32(dat[4:])
	return
}
