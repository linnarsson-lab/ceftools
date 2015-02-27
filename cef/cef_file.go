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
	// Ensure we're dealing with the correct version of the CEB file format
	var version int32
	var err := binary.Read(f, binary.LittleEndian, &version)
	if err != nil || version != 0x76302E31 {
		return nil, errors.New("This CEB file version is not supported by this version of Cellophane")
	}

	// Read the column and row counts
	var nCols int64
	if err = binary.Read(f, binary.LittleEndian, &nCols); err != nil { return nil, err }
	var nRows int64
	if err = binary.Read(f, binary.LittleEndian, &nRows); err != nil { return nil, err }

	// Read the column attributes
	var nColAttrs int32
	if err = binary.Read(f, binary.LittleEndian, &nColAttrs); err != nil { return nil, err }
	var colAttrNames [nColAttrs]string
	var colAttrValues [nCols][nColAttrs]string
	for c := 0; c < nColAttrs; c++ {
		if colAttrNames[c], err = readString(f); err != nil { return nil, err }

		for cv := 0; cv < nCols; cv++ {
			if colAttrValues[cv][c], err = readString(f); err != nil { return nil, err }
		}
	}

	// Read the row attributes
	var nRowAttrs int32
	if err = binary.Read(f, binary.LittleEndian, &nRowAttrs); err != nil { return nil, err }
	var rowAttrNames [nRowAttrs]string
	var rowAttrValues [nRows][nRowAttrs]string
	for r := 0; r < nRowAttrs; r++ {
		if rowAttrNames[r], err = readString(f); err != nil { return nil, err }

		for rv := 0; rv < nRows; rv++ {
			if rowAttrValues[rv][r], err = readString(f); err != nil { return nil, err }
		}
	}

	// Read the matrix
	var matrix [nCols][nRows]float32
	for c = 0; c < nCols; c++ {
		for r = 0; r < nRows; r++ {
			if err = binary.Read(f, binary.LittleEndian, &matrix[c][r]); err != nil { return nil, err }
		}
	}

	// Assemble the CEB structure and return
	return nil, nil
}

func readString(f, *os.File) (string, error) {
	var length int32
	if err = binary.Read(f, binary.LittleEndian, &length); err != nil { return "", err }

	var buffer [length]byte
	if err = binary.Read(f, binary.LittleEndian, &buffer); err != nil { return "", err }

	return string(buffer), nil
}
