package cef

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"os"
)

func Read(f *os.File) (*CefFile, error) {
	var magic int32
	binary.Read(f, binary.LittleEndian, &magic)

	if magic == 0x43454209 { // "CEB\t"
		return readCeb(f)
	}
	if magic == 0x43454609 { // "CEF\t"
		return readCef(f)
	}
	return nil, errors.New("Unknown file format")
}

func readCef(f *os.File) (*CefFile, error) {
	var cef = csv.NewReader(f)
	cef.Comma = '\t'
	cef.FieldsPerRecord = -1
	//	var row, err = cef.Read()

	return nil, nil
}

func readCeb(f *os.File) (*CefFile, error) {
	return nil, nil
}
