package ceftools

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"math"
	"os"
	"strconv"
)

func Read(f *os.File, transposed bool, skipMatrix bool) (*Cef, error) {
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

func WriteAsCEB(cef *Cef, f *os.File, transposed bool) error {
	writeString := func(s string) error {
		length := int32(len(s))
		if err := binary.Write(f, binary.LittleEndian, &length); err != nil {
			return err
		}
		if _, err := f.WriteString(s); err != nil {
			return err
		}
		return nil
	}

	magic := int32(0x43454209)
	if err := binary.Write(f, binary.LittleEndian, &magic); err != nil {
		return err
	}

	majorVersion := int32(MajorVersion)
	if err := binary.Write(f, binary.LittleEndian, &majorVersion); err != nil {
		return err
	}
	minorVersion := int32(MinorVersion)
	if err := binary.Read(f, binary.LittleEndian, &minorVersion); err != nil {
		return err
	}

	// Write the column and row counts
	if err := binary.Write(f, binary.LittleEndian, &cef.NumColumns); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, &cef.NumRows); err != nil {
		return err
	}
	// Write the flags
	if err := binary.Write(f, binary.LittleEndian, &cef.Flags); err != nil {
		return err
	}

	// Write the matrix
	if transposed {
		for j := int64(0); j < cef.NumRows; j++ {
			for i := int64(0); i < cef.NumColumns; i++ {
				value := cef.Get(i, j)
				if err := binary.Write(f, binary.LittleEndian, &value); err != nil {
					return err
				}
			}
		}
	} else {
		for i := int64(0); i < cef.NumColumns; i++ {
			for j := int64(0); j < cef.NumRows; j++ {
				value := cef.Get(i, j)
				if err := binary.Write(f, binary.LittleEndian, &value); err != nil {
					return err
				}
			}
		}
	}

	// Currently the skip section is unused
	nSkip := int64(0)
	if err := binary.Write(f, binary.LittleEndian, &nSkip); err != nil {
		return err
	}

	// Write the headers
	nHeaders := int32(len(cef.Headers))
	if err := binary.Write(f, binary.LittleEndian, &nHeaders); err != nil {
		return err
	}

	for i := int32(0); i < nHeaders; i++ {
		if err := writeString(cef.Headers[i].Name); err != nil {
			return err
		}
		if err := writeString(cef.Headers[i].Value); err != nil {
			return err
		}
	}

	// Helper to write attributes
	writeAttrs := func(attrs []Attribute) error {
		var nAttrs = int32(len(attrs))
		if err := binary.Write(f, binary.LittleEndian, &nAttrs); err != nil {
			return err
		}
		for i := int32(0); i < nAttrs; i++ {
			if err := writeString(attrs[i].Name); err != nil {
				return err
			}
		}
		for i := int32(0); i < nAttrs; i++ {
			for j := 0; j < len(attrs[0].Values); j++ {
				if err := writeString(attrs[i].Values[j]); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if transposed {
		writeAttrs(cef.RowAttributes)
		writeAttrs(cef.ColumnAttributes)
	} else {
		writeAttrs(cef.ColumnAttributes)
		writeAttrs(cef.RowAttributes)
	}

	return nil
}

func WriteAsCEF(cef *Cef, f *os.File, transposed bool) error {
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
	var row []string
	if transposed {
		row = make([]string, int(math.Max(7, float64(cef.NumRows+int64(len(cef.ColumnAttributes)+1)))))
	} else {
		row = make([]string, int(math.Max(7, float64(cef.NumColumns+int64(len(cef.RowAttributes)+1)))))
	}

	// Write the header line
	row[0] = "CEF"
	row[1] = strconv.Itoa(len(cef.Headers))
	row[2] = strconv.Itoa(int(cef.NumColumns))
	row[3] = strconv.Itoa(int(cef.NumRows))
	row[4] = strconv.Itoa(len(cef.ColumnAttributes))
	row[5] = strconv.Itoa(len(cef.RowAttributes))
	row[6] = strconv.Itoa(int(cef.Flags))
	if transposed {
		row[2] = strconv.Itoa(int(cef.NumRows))
		row[3] = strconv.Itoa(int(cef.NumColumns))
		row[4] = strconv.Itoa(len(cef.RowAttributes))
		row[5] = strconv.Itoa(len(cef.ColumnAttributes))
	}
	write(row)

	// Write the headers
	for i := 0; i < len(cef.Headers); i++ {
		row[0] = cef.Headers[0].Name
		row[1] = cef.Headers[1].Value
		write(row)
	}

	ralen := int64(len(cef.RowAttributes))
	calen := int64(len(cef.ColumnAttributes))

	if transposed {
		// Write the column attributes (from row attrs)
		for i := int64(0); i < ralen; i++ {
			row[calen] = cef.RowAttributes[i].Name
			for j := int64(0); j < cef.NumRows; j++ {
				row[j+calen+1] = cef.RowAttributes[i].Values[j]
			}
			write(row)
		}

		// Write the row attributes and matrix
		for i := int64(0); i < calen; i++ {
			row[i] = cef.ColumnAttributes[i].Name
		}
		write(row)
		for i := int64(0); i < cef.NumColumns; i++ {
			for j := int64(0); j < calen; j++ {
				row[j] = cef.ColumnAttributes[j].Values[i]
				for k := int64(0); k < cef.NumRows; k++ {
					row[k+ralen+1] = strconv.FormatFloat(float64(cef.Get(k, i)), 'f', -1, 64)
				}
			}
			write(row)
		}
	} else {
		// Write the column attributes
		for i := int64(0); i < calen; i++ {
			row[ralen] = cef.ColumnAttributes[i].Name
			for j := int64(0); j < cef.NumColumns; j++ {
				row[j+ralen+1] = cef.ColumnAttributes[i].Values[j]
			}
			write(row)
		}

		// Write the row attributes and matrix
		for i := int64(0); i < ralen; i++ {
			row[i] = cef.RowAttributes[i].Name
		}
		write(row)
		for i := int64(0); i < cef.NumRows; i++ {
			for j := int64(0); j < ralen; j++ {
				row[j] = cef.RowAttributes[j].Values[i]
				for k := int64(0); k < cef.NumColumns; k++ {
					row[k+ralen+1] = strconv.FormatFloat(float64(cef.Get(k, i)), 'f', -1, 64)
				}
			}
			write(row)
		}
	}
	return nil
}

func readCEF(f *os.File) (*Cef, error) {
	var cef = csv.NewReader(f)
	cef.Comma = '\t'
	cef.FieldsPerRecord = -1
	//	var row, err = cf.Read()

	return nil, nil
}

func readCEB(f *os.File, transposed bool, skipMatrix bool) (*Cef, error) {
	// Allocate a CF file struct
	var cef Cef

	// Ensure we're dealing with the correct version of the CEB file format
	err := binary.Read(f, binary.LittleEndian, &cef.MajorVersion)
	if err != nil || cef.MajorVersion > 0 {
		return nil, errors.New("This CEB file version is not supported by this version of Cellophane")
	}
	// The minor version is ignored (given that the major version was ok); changes should be backward compatible
	err = binary.Read(f, binary.LittleEndian, &cef.MinorVersion)
	if err != nil {
		return nil, err
	}

	// Read the column and row counts
	if err = binary.Read(f, binary.LittleEndian, &cef.NumColumns); err != nil {
		return nil, err
	}
	if err = binary.Read(f, binary.LittleEndian, &cef.NumRows); err != nil {
		return nil, err
	}
	// Read the flags
	if err = binary.Read(f, binary.LittleEndian, &cef.Flags); err != nil {
		return nil, err
	}

	// Maybe we can skip the matrix?
	if skipMatrix {
		f.Seek(cef.NumColumns*cef.NumRows*4, 1)
	} else {
		// Read the matrix
		cef.Matrix = make([]float32, cef.NumColumns*cef.NumRows)
		for i := int64(0); i < cef.NumColumns; i++ {
			for j := int64(0); j < cef.NumRows; j++ {
				var value float32
				if err = binary.Read(f, binary.LittleEndian, &value); err != nil {
					return nil, err
				}
				if transposed {
					cef.Matrix[i*cef.NumRows+j] = value // TODO: verify this
				} else {
					cef.Matrix[i+j*cef.NumColumns] = value
				}
			}
		}
	}

	// Exchange the row column counts
	if transposed {
		temp := cef.NumRows
		cef.NumRows = cef.NumColumns
		cef.NumColumns = temp
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
	cef.Headers = make([]Header, nHeaders)
	for i := int32(0); i < nHeaders; i++ {
		hdrName, err := readString(f)
		if err != nil {
			return nil, err
		}
		hdrValue, err := readString(f)
		if err != nil {
			return nil, err
		}
		cef.Headers[i] = Header{hdrName, hdrValue}
	}

	// Read the column attributes
	var nColAttrs int32
	if err = binary.Read(f, binary.LittleEndian, &nColAttrs); err != nil {
		return nil, err
	}
	cef.ColumnAttributes = make([]Attribute, nColAttrs)
	for i := int32(0); i < nColAttrs; i++ {
		colAttrName, err := readString(f)
		if err != nil {
			return nil, err
		}
		cef.ColumnAttributes[i] = Attribute{colAttrName, make([]string, cef.NumColumns)}
	}
	for i := int32(0); i < nColAttrs; i++ {
		for j := int64(0); j < cef.NumColumns; j++ {
			if cef.ColumnAttributes[i].Values[j], err = readString(f); err != nil {
				return nil, err
			}
		}
	}

	// Read the row attributes
	var nRowAttrs int32
	if err = binary.Read(f, binary.LittleEndian, &nRowAttrs); err != nil {
		return nil, err
	}
	cef.RowAttributes = make([]Attribute, nRowAttrs)
	for i := int32(0); i < nRowAttrs; i++ {
		rowAttrName, err := readString(f)
		if err != nil {
			return nil, err
		}
		cef.RowAttributes[i] = Attribute{rowAttrName, make([]string, cef.NumRows)}
	}
	for i := int32(0); i < nRowAttrs; i++ {
		for j := int64(0); j < cef.NumRows; j++ {
			if cef.RowAttributes[i].Values[j], err = readString(f); err != nil {
				return nil, err
			}
		}
	}

	// Exchange the row column attributes
	if transposed {
		temp := cef.RowAttributes
		cef.RowAttributes = cef.ColumnAttributes
		cef.ColumnAttributes = temp
	}

	return &cef, nil
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
