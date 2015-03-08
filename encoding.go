package cef

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"math"
	"os"
	"strconv"
)

func Read(f *os.File, transposed bool, skipMatrix bool) (*CefFile, error) {
	var magic int32
	binary.Read(f, binary.LittleEndian, &magic)

	if magic == 0x43454209 { // "CEB\t"
		return readCEB(f, transposed, skipMatrix)
	}
	if magic == 0x43454609 { // "CEF\t"
		return readCEF(f)
	}
	return nil, errors.New("Unknown file format")
}

func WriteAsCEB(cf *CefFile, f *os.File, transposed bool) error {
	return nil
}

func WriteAsCEF(cf *CefFile, f *os.File, transposed bool) error {
	w := csv.NewWriter(f)
	w.Comma = '\t'
	write := func(row []string) {
		w.Write(row)
		w.Flush()
		for i := 0; i < len(row); i++ {
			row[i] = ""
		}
	}

	// Make a vector to hold each line (we'll reuse it)
	row := make([]string, int(math.Max(7, float64(cf.NumColumns+int64(len(cf.RowAttributes)+1)))))

	// Write the header line
	row[0] = "CEF"
	row[1] = strconv.Itoa(len(cf.Headers))
	row[2] = strconv.Itoa(int(cf.NumColumns))
	row[3] = strconv.Itoa(int(cf.NumRows))
	row[4] = strconv.Itoa(len(cf.ColumnAttributes))
	row[5] = strconv.Itoa(len(cf.RowAttributes))
	row[6] = strconv.Itoa(int(cf.Flags))
	write(row)

	// Write the headers
	for i := 0; i < len(cf.Headers); i++ {
		row[0] = cf.Headers[0].Name
		row[1] = cf.Headers[1].Value
		write(row)
	}

	// Write the column attributes
	ralen := int64(len(cf.RowAttributes))
	calen := int64(len(cf.ColumnAttributes))
	for i := int64(0); i < calen; i++ {
		row[ralen] = cf.ColumnAttributes[i].Name
		for j := int64(0); j < cf.NumColumns; j++ {
			row[j+ralen+1] = cf.ColumnAttributes[i].Values[j]
		}
		write(row)
	}

	// Write the row attributes and matrix
	for i := int64(0); i < ralen; i++ {
		row[i] = cf.RowAttributes[i].Name
	}
	write(row)
	for i := int64(0); i < cf.NumRows; i++ {
		for j := int64(0); j < ralen; j++ {
			row[j] = cf.RowAttributes[j].Values[i]
			for k := int64(0); k < cf.NumColumns; k++ {
				row[k+ralen+1] = strconv.FormatFloat(float64(cf.Get(k, i)), 'f', -1, 64)
			}
		}
		write(row)
	}
	return nil
}

func readCEF(f *os.File) (*CefFile, error) {
	var cf = csv.NewReader(f)
	cf.Comma = '\t'
	cf.FieldsPerRecord = -1
	//	var row, err = cf.Read()

	return nil, nil
}

func readCEB(f *os.File, transposed bool, skipMatrix bool) (*CefFile, error) {
	// Allocate a CF file struct
	var cf CefFile

	// Ensure we're dealing with the correct version of the CEB file format
	err := binary.Read(f, binary.LittleEndian, &cf.MajorVersion)
	if err != nil || cf.MajorVersion > 0 {
		return nil, errors.New("This CEB file version is not supported by this version of Cellophane")
	}
	// The minor version is ignored (given that the major version was ok); changes should be backward compatible
	err = binary.Read(f, binary.LittleEndian, &cf.MinorVersion)
	if err != nil {
		return nil, err
	}

	// Read the column and row counts
	if err = binary.Read(f, binary.LittleEndian, &cf.NumColumns); err != nil {
		return nil, err
	}
	if err = binary.Read(f, binary.LittleEndian, &cf.NumRows); err != nil {
		return nil, err
	}
	// Read the flags
	if err = binary.Read(f, binary.LittleEndian, &cf.Flags); err != nil {
		return nil, err
	}

	// Maybe we can skip the matrix?
	if skipMatrix {
		f.Seek(cf.NumColumns*cf.NumRows*4, 1)
	} else {
		// Read the matrix
		cf.Matrix = make([]float32, cf.NumColumns*cf.NumRows)
		for i := int64(0); i < cf.NumColumns; i++ {
			for j := int64(0); j < cf.NumRows; j++ {
				var value float32
				if err = binary.Read(f, binary.LittleEndian, &value); err != nil {
					return nil, err
				}
				if transposed {
					cf.Matrix[i*cf.NumRows+j] = value // TODO: verify this
				} else {
					cf.Matrix[i+j*cf.NumColumns] = value
				}
			}
		}
	}

	// Exchange the row column counts
	if transposed {
		temp := cf.NumRows
		cf.NumRows = cf.NumColumns
		cf.NumColumns = temp
	}

	// Skip some bytes
	var nSkip int64
	if err = binary.Read(f, binary.LittleEndian, &nSkip); err != nil {
		return nil, err
	}
	if nSkip > 0 {
		if _, err = f.Seek(nSkip, 1); err != nil {
			return nil, err
		}
	}

	// Read the headers
	var nHeaders int32
	if err = binary.Read(f, binary.LittleEndian, &nHeaders); err != nil {
		return nil, err
	}
	cf.Headers = make([]CefHeader, nHeaders)
	for i := int32(0); i < nHeaders; i++ {
		hdrName, err := readString(f)
		if err != nil {
			return nil, err
		}
		hdrValue, err := readString(f)
		if err != nil {
			return nil, err
		}
		cf.Headers[i] = CefHeader{hdrName, hdrValue}
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
	}
	for i := int32(0); i < nColAttrs; i++ {
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
	}
	for i := int32(0); i < nRowAttrs; i++ {
		for j := int64(0); j < cf.NumRows; j++ {
			if cf.RowAttributes[i].Values[j], err = readString(f); err != nil {
				return nil, err
			}
		}
	}

	// Exchange the row column attributes
	if transposed {
		temp := cf.RowAttributes
		cf.RowAttributes = cf.ColumnAttributes
		cf.ColumnAttributes = temp
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
