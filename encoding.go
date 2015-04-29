package ceftools

import (
	"bufio"
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
		row = make([]string, int(math.Max(7, float64(cef.Rows+len(cef.ColumnAttributes)+1))))
	} else {
		row = make([]string, int(math.Max(7, float64(cef.Columns+len(cef.RowAttributes)+1))))
	}

	// Write the header line
	row[0] = "CEF"
	row[1] = strconv.Itoa(len(cef.Headers))
	row[2] = strconv.Itoa(len(cef.RowAttributes))
	row[3] = strconv.Itoa(len(cef.ColumnAttributes))
	row[4] = strconv.Itoa(cef.Rows)
	row[5] = strconv.Itoa(cef.Columns)
	row[6] = strconv.Itoa(cef.Flags)
	if transposed {
		row[2] = strconv.Itoa(len(cef.ColumnAttributes))
		row[3] = strconv.Itoa(len(cef.RowAttributes))
		row[4] = strconv.Itoa(cef.Columns)
		row[5] = strconv.Itoa(cef.Rows)
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
			for j := 0; j < cef.Rows; j++ {
				row[j+calen+1] = cef.RowAttributes[i].Values[j]
			}
			write(row)
		}

		// Write the row attributes and matrix
		for i := 0; i < calen; i++ {
			row[i] = cef.ColumnAttributes[i].Name
		}
		write(row)
		for i := 0; i < cef.Columns; i++ {
			for j := 0; j < calen; j++ {
				row[j] = cef.ColumnAttributes[j].Values[i]
				for k := 0; k < cef.Rows; k++ {
					row[k+calen+1] = strconv.FormatFloat(float64(cef.Get(k, i)), 'f', -1, 64)
				}
			}
			write(row)
		}
	} else {
		// Write the column attributes
		for i := 0; i < calen; i++ {
			row[ralen] = cef.ColumnAttributes[i].Name
			for j := 0; j < cef.Columns; j++ {
				row[j+ralen+1] = cef.ColumnAttributes[i].Values[j]
			}
			write(row)
		}

		// Write the row attributes and matrix
		for i := 0; i < ralen; i++ {
			row[i] = cef.RowAttributes[i].Name
		}
		write(row)
		for i := 0; i < cef.Rows; i++ {
			for j := 0; j < ralen; j++ {
				row[j] = cef.RowAttributes[j].Values[i]
				for k := 0; k < cef.Columns; k++ {
					row[k+ralen+1] = strconv.FormatFloat(float64(cef.Get(i, k)), 'f', -1, 64)
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
	cef.Columns = len(row) - nRowAttrs
	parsingColAttrs := true
	for parsingColAttrs {
		attr := Attribute{row[nRowAttrs-1], make([]string, cef.Columns)}
		if attr.Name[len(attr.Name)-1] == ':' {
			attr.Name = attr.Name[:len(attr.Name)-1]
		}
		for i := nRowAttrs; i < cef.Columns+nRowAttrs; i++ {
			attr.Values[i-nRowAttrs] = row[i]
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
				cef.Rows = nRows
				return cef, nil
			}
			return nil, err
		}
		nRows++

		// Row attributes
		for i := 0; i < nRowAttrs; i++ {
			cef.RowAttributes[i].Values = append(cef.RowAttributes[i].Values, row[i])
		}
		values := make([]float32, cef.Columns)
		for i := 0; i < cef.Columns; i++ {
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

func Read_old(f *os.File, transposed bool) (*Cef, error) {
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
	nRowAttrs, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, errors.New("Row attribute count (row 1, column 6) is not a valid integer")
	}
	nColumnAttrs, err := strconv.Atoi(row[3])
	if err != nil {
		return nil, errors.New("Column attribute count (row 1, column 5) is not a valid integer")
	}
	nRows, err := strconv.Atoi(row[4])
	if err != nil {
		return nil, errors.New("Row count (row 1, column 4) is not a valid integer")
	}
	nColumns, err := strconv.Atoi(row[5])
	if err != nil {
		return nil, errors.New("Column count (row 1, column 3) is not a valid integer")
	}
	flags, err := strconv.Atoi(row[6])
	if err != nil {
		return nil, errors.New("Flags value (row 1, column 7) is not a valid integer")
	}
	cef.Rows = nRows
	cef.Columns = nColumns
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
		if len(row) < nRowAttrs+int(cef.Columns)+1 {
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
		cef.RowAttributes[i] = Attribute{row[i], make([]string, cef.Rows)}
	}

	// Read the rows, with row attribute values
	cef.Matrix = make([]float32, cef.Columns*cef.Rows)
	for i := 0; i < cef.Rows; i++ {
		row, err := r.Read()
		if err != nil {
			return nil, err
		}
		if len(row) < nRowAttrs+int(cef.Columns)+1 {
			return nil, errors.New(fmt.Sprintf("Row number %v is not the right length (number of columns is wrong)", len(cef.Headers)+3+nColumnAttrs+i))
		}
		for j := 0; j < nRowAttrs; j++ {
			cef.RowAttributes[j].Values[i] = row[j]
		}
		for j := 0; j < int(cef.Columns); j++ {
			val, err := strconv.ParseFloat(row[j+nRowAttrs+1], 32)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Invalid float32 value in column %v, row %v of the main matrix: %v", j+1, i+1, row[j+nRowAttrs]))
			}
			if transposed {
				cef.Matrix[j*cef.Rows+i] = float32(val)
			} else {
				cef.Matrix[j+i*cef.Columns] = float32(val)
			}
		}
	}

	// Exchange the rows and columns
	if transposed {
		temp1 := cef.Rows
		cef.Rows = cef.Columns
		cef.Columns = temp1
		temp2 := cef.RowAttributes
		cef.RowAttributes = cef.ColumnAttributes
		cef.ColumnAttributes = temp2
	}
	return cef, nil
}

func nextString(f *bufio.Reader) string {
	result := make([]rune, 0, 10)
	for {
		r, _, err := f.ReadRune()
		if err != nil {
			panic(err.Error())
		}
		if r == '\t' {
			return string(result)
		}
		if r == '\r' || r == '\n' {
			f.UnreadRune()
			return string(result)
		}
		result = append(result, r)
	}
}

func readStrings(f *bufio.Reader, n int) []string {
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = nextString(f)
	}
	return result
}

func skipFields(f *bufio.Reader, n int) {
	for i := 0; i < n; i++ {
		nextString(f)
	}
}

func nextLine(f *bufio.Reader) {
	// Consume whitespace
	for {
		r, _, err := f.ReadRune()
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err.Error())
		}
		if r != ' ' && r != '\t' {
			f.UnreadRune()
			break
		}
	}
	// Consume any number of endline runes
	for {
		r, _, err := f.ReadRune()
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err.Error())
		}
		if r != '\r' && r != '\n' {
			f.UnreadRune()
			return
		}
	}
}

func Read(f *os.File, transposed bool) (*Cef, error) {
	var r = bufio.NewReader(f)
	cef := new(Cef)

	if nextString(r) != "CEF" {
		return nil, errors.New("Unknown file format")
	}

	// Parse the header line (the first field, 'CEF' has already been consumed)
	nHeaders, err := strconv.Atoi(nextString(r))
	if err != nil {
		return nil, errors.New("Header count (row 1, column 2) is not a valid integer")
	}
	nRowAttrs, err := strconv.Atoi(nextString(r))
	if err != nil {
		return nil, errors.New("Row attribute count (row 1, column 6) is not a valid integer")
	}
	nColumnAttrs, err := strconv.Atoi(nextString(r))
	if err != nil {
		return nil, errors.New("Column attribute count (row 1, column 5) is not a valid integer")
	}
	nRows, err := strconv.Atoi(nextString(r))
	if err != nil {
		return nil, errors.New("Row count (row 1, column 4) is not a valid integer")
	}
	nColumns, err := strconv.Atoi(nextString(r))
	if err != nil {
		return nil, errors.New("Column count (row 1, column 3) is not a valid integer")
	}
	flags, err := strconv.Atoi(nextString(r))
	if err != nil {
		return nil, errors.New("Flags value (row 1, column 7) is not a valid integer")
	}
	nextLine(r)
	cef.Rows = nRows
	cef.Columns = nColumns
	cef.Flags = flags

	// Read the headers
	cef.Headers = make([]Header, nHeaders)
	for i := 0; i < len(cef.Headers); i++ {
		cef.Headers[i].Name = nextString(r)
		cef.Headers[i].Value = nextString(r)
		nextLine(r)
	}

	// Read the column attributes
	cef.ColumnAttributes = make([]Attribute, nColumnAttrs)
	for i := 0; i < nColumnAttrs; i++ {
		skipFields(r, nRowAttrs)
		cef.ColumnAttributes[i] = Attribute{nextString(r), readStrings(r, nColumns)}
		nextLine(r)
	}

	// Read the row attribute names and create row attributes
	cef.RowAttributes = make([]Attribute, nRowAttrs)
	for i := 0; i < nRowAttrs; i++ {
		ra := nextString(r)
		if ra == "" {
			return nil, errors.New(fmt.Sprintf("Row attribute name cannot be empty (name missing in column %v)", i+1))
		}
		cef.RowAttributes[i] = Attribute{ra, make([]string, cef.Rows)}
	}
	nextLine(r)

	// Read the rows, with row attribute values
	cef.Matrix = make([]float32, cef.Columns*cef.Rows)
	for i := 0; i < cef.Rows; i++ {
		for j := 0; j < nRowAttrs; j++ {
			cef.RowAttributes[j].Values[i] = nextString(r)
		}
		skipFields(r, 1)
		for j := 0; j < int(cef.Columns); j++ {
			val, err := strconv.ParseFloat(nextString(r), 32)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Invalid float32 value in column %v, row %v of the main matrix", j+1, i+1))
			}
			if transposed {
				cef.Matrix[j*cef.Rows+i] = float32(val)
			} else {
				cef.Matrix[j+i*cef.Columns] = float32(val)
			}
		}
		nextLine(r)
	}

	// Exchange the rows and columns
	if transposed {
		temp1 := cef.Rows
		cef.Rows = cef.Columns
		cef.Columns = temp1
		temp2 := cef.RowAttributes
		cef.RowAttributes = cef.ColumnAttributes
		cef.ColumnAttributes = temp2
	}
	return cef, nil
}
