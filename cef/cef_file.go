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
	if magic == 0x43454609 { // "CF\t"
		return readCf(f)
	}
	return nil, errors.New("Unknown file format")
}

func readCf(f *os.File) (*CefFile, error) {
	var cf = csv.NewReader(f)
	cf.Comma = '\t'
	cf.FieldsPerRecord = -1
	//	var row, err = cf.Read()

	return nil, nil
}

func readCeb(f *os.File) (*CefFile, error) {
	// Allocate a CF file struct
	var cf CefFile

	// Ensure we're dealing with the correct version of the CEB file format
	var version int32
	err := binary.Read(f, binary.LittleEndian, &version)
	if err != nil || version != 0x76302E31 {
		return nil, errors.New("This CEB file version is not supported by this version of Cellophane")
	}

	// Read the column and row counts
	if err = binary.Read(f, binary.LittleEndian, &cf.NumColumns); err != nil {
		return nil, err
	}
	if err = binary.Read(f, binary.LittleEndian, &cf.NumRows); err != nil {
		return nil, err
	}

	// Read the column attributes
	var nColAttrs int32
	if err = binary.Read(f, binary.LittleEndian, &nColAttrs); err != nil {
		return nil, err
	}
	cf.ColumnAttributes = make([]CefAttribute, nColAttrs)
	for i := int32(0); i < nColAttrs; i++ {
		colAttrName, err := readString(f)
		if err != nil {
			return nil, err
		}
		cf.ColumnAttributes[i] = CefAttribute{colAttrName, make([]string, cf.NumColumns)}
		for j := int64(0); j < cf.NumColumns; j++ {
			if cf.ColumnAttributes[i].Values[j], err = readString(f); err != nil {
				return nil, err
			}
		}
	}

	// Read the row attributes
	var nRowAttrs int32
	if err = binary.Read(f, binary.LittleEndian, &nRowAttrs); err != nil {
		return nil, err
	}
	cf.RowAttributes = make([]CefAttribute, nRowAttrs)
	for i := int32(0); i < nRowAttrs; i++ {
		rowAttrName, err := readString(f)
		if err != nil {
			return nil, err
		}
		cf.RowAttributes[i] = CefAttribute{rowAttrName, make([]string, cf.NumRows)}
		for j := int64(0); j < cf.NumRows; j++ {
			if cf.RowAttributes[i].Values[j], err = readString(f); err != nil {
				return nil, err
			}
		}
	}

	// Read the matrix
	cf.Matrix = make([]float32, cf.NumColumns*cf.NumRows)
	for i := int64(0); i < cf.NumColumns; i++ {
		for j := int64(0); j < cf.NumRows; j++ {
			var value float32
			if err = binary.Read(f, binary.LittleEndian, &value); err != nil {
				return nil, err
			}
			cf.Matrix[i+j*cf.NumColumns] = value
		}
	}

	return &cf, nil
}

func readString(f *os.File) (string, error) {
	var length int32
	if err := binary.Read(f, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	var buffer = make([]byte, length)
	if err := binary.Read(f, binary.LittleEndian, &buffer); err != nil {
		return "", err
	}

	return string(buffer), nil
}
