package ceftools

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

func Write(cef *Cef, f *os.File, transposed bool) error {
	w := csv.NewWriter(f)
	w.Comma = '\t'
	write := func(row []string) {
		w.Write(row)
		for i := 0; i < len(row); i++ {
			row[i] = ""
		}
	}

	// Make a vector to hold each line (we'll reuse it)
	var row []string
	if transposed {
		row = make([]string, int(math.Max(7, float64(cef.NumRows+len(cef.ColumnAttributes)+1))))
	} else {
		row = make([]string, int(math.Max(7, float64(cef.NumColumns+len(cef.RowAttributes)+1))))
	}

	// Write the header line
	row[0] = "CEF"
	row[1] = strconv.Itoa(len(cef.Headers))
	row[2] = strconv.Itoa(cef.NumColumns)
	row[3] = strconv.Itoa(cef.NumRows)
	row[4] = strconv.Itoa(len(cef.ColumnAttributes))
	row[5] = strconv.Itoa(len(cef.RowAttributes))
	row[6] = strconv.Itoa(cef.Flags)
	if transposed {
		row[2] = strconv.Itoa(cef.NumRows)
		row[3] = strconv.Itoa(cef.NumColumns)
		row[4] = strconv.Itoa(len(cef.RowAttributes))
		row[5] = strconv.Itoa(len(cef.ColumnAttributes))
	}
	write(row)

	// Write the headers
	for i := 0; i < len(cef.Headers); i++ {
		row[0] = cef.Headers[i].Name
		row[1] = cef.Headers[i].Value
		write(row)
	}

	ralen := len(cef.RowAttributes)
	calen := len(cef.ColumnAttributes)

	if transposed {
		// Write the column attributes (from row attrs)
		for i := 0; i < ralen; i++ {
			row[calen] = cef.RowAttributes[i].Name
			for j := 0; j < cef.NumRows; j++ {
				row[j+calen+1] = cef.RowAttributes[i].Values[j]
			}
			write(row)
		}

		// Write the row attributes and matrix
		for i := 0; i < calen; i++ {
			row[i] = cef.ColumnAttributes[i].Name
		}
		write(row)
		for i := 0; i < cef.NumColumns; i++ {
			for j := 0; j < calen; j++ {
				row[j] = cef.ColumnAttributes[j].Values[i]
				for k := 0; k < cef.NumRows; k++ {
					row[k+ralen+1] = strconv.FormatFloat(float64(cef.Get(i, k)), 'f', -1, 64)
				}
			}
			write(row)
		}
	} else {
		// Write the column attributes
		for i := 0; i < calen; i++ {
			row[ralen] = cef.ColumnAttributes[i].Name
			for j := 0; j < cef.NumColumns; j++ {
				row[j+ralen+1] = cef.ColumnAttributes[i].Values[j]
			}
			write(row)
		}

		// Write the row attributes and matrix
		for i := 0; i < ralen; i++ {
			row[i] = cef.RowAttributes[i].Name
		}
		write(row)
		for i := 0; i < cef.NumRows; i++ {
			for j := 0; j < ralen; j++ {
				row[j] = cef.RowAttributes[j].Values[i]
				for k := 0; k < cef.NumColumns; k++ {
					row[k+ralen+1] = strconv.FormatFloat(float64(cef.Get(k, i)), 'f', -1, 64)
				}
			}
			write(row)
		}
	}
	w.Flush()
	return nil
}

func ReadStrt(f *os.File, transposed bool) (*Cef, error) {
	var r = csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1

	cef := new(Cef)
	cef.Headers = make([]Header, 0)

	// Make a vector to hold each line (we'll reuse it)
	row, err := r.Read()
	if err != nil {
		return nil, err
	}
	if row[0][0] == '#' {
		cef.Headers = append(cef.Headers, Header{"Comment", row[0][1:]})
		row, err = r.Read()
		if err != nil {
			return nil, err
		}
	}

	// Count the row attributes by counting leading empty columns
	nRowAttrs := 0
	for i := 0; i < len(row); i++ {
		nRowAttrs++
		if row[i] != "" {
			break
		}
	}

	// Parse the column attributes
	cef.ColumnAttributes = make([]Attribute, 0)
	cef.NumColumns = len(row) - nRowAttrs
	parsingColAttrs := true
	for parsingColAttrs {
		attr := Attribute{row[nRowAttrs-1], make([]string, cef.NumColumns)}
		if attr.Name[len(attr.Name)-1] == ':' {
			attr.Name = attr.Name[:len(attr.Name)-1]
		}
		for i := nRowAttrs - 1; i < cef.NumColumns; i++ {
			attr.Values[i-nRowAttrs+1] = row[i]
		}
		cef.ColumnAttributes = append(cef.ColumnAttributes, attr)

		row, err = r.Read()
		if err != nil {
			return nil, err
		}
		if row[0] != "" {
			parsingColAttrs = false
		}
	}

	// Read the row attribute names
	cef.RowAttributes = make([]Attribute, nRowAttrs)
	for i := 0; i < nRowAttrs; i++ {
		cef.RowAttributes[i] = Attribute{row[i], make([]string, 0)}
	}

	// Read the rows and row attributes
	cef.Matrix = make([]float32, 0)
	nRows := 0
	for true {
		row, err = r.Read()
		if err != nil {
			if err == io.EOF {
				cef.NumRows = nRows
				return cef, nil
			}
			return nil, err
		}
		nRows++

		// Row attributes
		for i := 0; i < nRowAttrs; i++ {
			cef.RowAttributes[i].Values = append(cef.RowAttributes[i].Values, row[i])
		}
		values := make([]float32, cef.NumColumns)
		for i := 0; i < cef.NumColumns; i++ {
			value, err := strconv.ParseFloat(row[i+nRowAttrs], 32)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Error parsing value in row %v, column %v", nRows, i+1))
			}
			values[i] = float32(value)
		}
		cef.Matrix = append(cef.Matrix, values...)
	}
	return cef, nil
}

func Read(f *os.File, transposed bool) (*Cef, error) {
	var r = csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1

	cef := new(Cef)

	// Make a vector to hold each line (we'll reuse it)
	row, err := r.Read()
	if err != nil {
		return nil, err
	}

	if row[0] != "CEF" {
		return nil, errors.New("Unknown file format")
	}

	// Parse the header line (the first field, 'CEF' has already been consumed)
	nHeaders, err := strconv.Atoi(row[1])
	if err != nil {
		return nil, errors.New("Header count (row 1, column 2) is not a valid integer")
	}
	nColumns, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, errors.New("Column count (row 1, column 3) is not a valid integer")
	}
	nRows, err := strconv.Atoi(row[3])
	if err != nil {
		return nil, errors.New("Row count (row 1, column 4) is not a valid integer")
	}
	nColumnAttrs, err := strconv.Atoi(row[4])
	if err != nil {
		return nil, errors.New("Column attribute count (row 1, column 5) is not a valid integer")
	}
	nRowAttrs, err := strconv.Atoi(row[5])
	if err != nil {
		return nil, errors.New("Row attribute count (row 1, column 6) is not a valid integer")
	}
	flags, err := strconv.Atoi(row[6])
	if err != nil {
		return nil, errors.New("Flags value (row 1, column 7) is not a valid integer")
	}
	cef.NumRows = nRows
	cef.NumColumns = nColumns
	cef.Flags = flags

	// Read the headers
	cef.Headers = make([]Header, nHeaders)
	for i := 0; i < len(cef.Headers); i++ {
		row, err := r.Read()
		if err != nil {
			return nil, err
		}
		if len(row) < 2 || row[0] == "" || row[1] == "" {
			return nil, errors.New(fmt.Sprintf("Invalid header in row %v: name and/or value missing", i+2))
		}
		cef.Headers[i].Name = row[0]
		cef.Headers[i].Value = row[1]
	}

	// Read the column attributes
	cef.ColumnAttributes = make([]Attribute, nColumnAttrs)
	for i := 0; i < nColumnAttrs; i++ {
		row, err := r.Read()
		if err != nil {
			return nil, err
		}
		if len(row) != nRowAttrs+int(cef.NumColumns)+1 {
			return nil, errors.New(fmt.Sprintf("Invalid column attribute in row %v: wrong number of values", len(cef.Headers)+2+i))
		}
		cef.ColumnAttributes[i] = Attribute{row[nRowAttrs], row[nRowAttrs+1:]}
	}

	// Read the row attribute names and create row attributes
	cef.RowAttributes = make([]Attribute, nRowAttrs)
	row, err = r.Read()
	if err != nil {
		return nil, err
	}
	if len(row) < nRowAttrs {
		return nil, errors.New(fmt.Sprintf("Number of row attribute names (%v) is less than number indicated in header (%v)", len(row), nRowAttrs))
	}
	for i := 0; i < nRowAttrs; i++ {
		if row[i] == "" {
			return nil, errors.New(fmt.Sprintf("Row attribute name cannot be empty (name missing in column %v)", i+1))
		}
		cef.RowAttributes[i] = Attribute{row[i], make([]string, cef.NumRows)}
	}

	// Read the rows, with row attribute values
	cef.Matrix = make([]float32, cef.NumColumns*cef.NumRows)
	for i := 0; i < cef.NumRows; i++ {
		row, err := r.Read()
		if err != nil {
			return nil, err
		}
		if len(row) != nRowAttrs+int(cef.NumColumns)+1 {
			return nil, errors.New(fmt.Sprintf("Row number %v is not the right length (number of columns is wrong)", len(cef.Headers)+3+nColumnAttrs+i))
		}
		for j := 0; j < nRowAttrs; j++ {
			cef.RowAttributes[j].Values[i] = row[j]
		}
		for j := 0; j < int(cef.NumColumns); j++ {
			val, err := strconv.ParseFloat(row[j+nRowAttrs+1], 32)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Invalid float32 value in column %v, row %v of the main matrix: %v", j+1, i+1, row[j+nRowAttrs]))
			}
			if transposed {
				cef.Matrix[j*cef.NumRows+i] = float32(val)
			} else {
				cef.Matrix[j+i*cef.NumColumns] = float32(val)
			}
		}
	}
	// Exchange the rows and columns
	if transposed {
		temp1 := cef.NumRows
		cef.NumRows = cef.NumColumns
		cef.NumColumns = temp1
		temp2 := cef.RowAttributes
		cef.RowAttributes = cef.ColumnAttributes
		cef.ColumnAttributes = temp2
	}
	return cef, nil
}
