package ceftools

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Attribute struct {
	Name   string
	Values []string
}

type Header struct {
	Name  string
	Value string
}

const (
	Transposed = 1 << iota
)

const (
	MajorVersion = 0
	MinorVersion = 1
	MagicCEB     = 0x09424543
	MagicCEF     = 0x09464543
)

type Cef struct {
	NumRows          int
	NumColumns       int
	Headers          []Header
	Flags            int
	RowAttributes    []Attribute
	ColumnAttributes []Attribute
	Matrix           []float32
}

func (cef Cef) Get(col int, row int) float32 {
	return cef.Matrix[col+row*cef.NumColumns]
}

func (cef Cef) Set(col int, row int, val float32) {
	cef.Matrix[col+row*cef.NumColumns] = val
}

func (cef Cef) GetRow(row int) []float32 {
	return cef.Matrix[row*cef.NumColumns : (row+1)*cef.NumColumns]
}

type stringRec struct {
	value string
	index int
}
type indexedStrings []stringRec

func (a indexedStrings) Len() int           { return len(a) }
func (a indexedStrings) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a indexedStrings) Less(i, j int) bool { return a[i].value < a[j].value }

func (cef Cef) SortByRowAttribute(attr string, reverse bool) (*Cef, error) {
	// Find the indexes
	var index []string
	for i := 0; i < len(cef.RowAttributes); i++ {
		if cef.RowAttributes[i].Name == attr {
			index = cef.RowAttributes[i].Values
		}
	}
	if index == nil {
		return nil, errors.New("Attribute not found when attempting to sort: " + attr)
	}

	// Collect the values to be sorted
	recs := make([]stringRec, len(index))
	for i := 0; i < len(index); i++ {
		recs[i] = stringRec{index[i], i}
	}
	// Sort them
	sort.Sort(indexedStrings(recs))

	// Make the resulting Cef
	result := new(Cef)
	result.NumColumns = cef.NumColumns
	result.NumRows = cef.NumRows
	result.Headers = cef.Headers
	result.Flags = cef.Flags
	result.Matrix = make([]float32, 0)
	result.ColumnAttributes = cef.ColumnAttributes
	result.RowAttributes = make([]Attribute, len(cef.RowAttributes))
	for i := 0; i < len(cef.RowAttributes); i++ {
		result.RowAttributes[i].Name = cef.RowAttributes[i].Name
		result.RowAttributes[i].Values = make([]string, len(result.RowAttributes[i].Values))
	}
	if reverse {
		for i := cef.NumRows - 1; i >= 0; i-- {
			from := recs[i].index
			result.Matrix = append(result.Matrix, cef.GetRow(from)...)
			for j := 0; j < len(cef.RowAttributes); j++ {
				result.RowAttributes[j].Values = append(result.RowAttributes[j].Values, cef.RowAttributes[j].Values[from])
			}
		}
	} else {
		for i := 0; i < cef.NumRows; i++ {
			from := recs[i].index
			result.Matrix = append(result.Matrix, cef.GetRow(from)...)
			for j := 0; j < len(cef.RowAttributes); j++ {
				result.RowAttributes[j].Values = append(result.RowAttributes[j].Values, cef.RowAttributes[j].Values[from])
			}
		}
	}
	return result, nil
}

type numberRec struct {
	value float32
	index int
}
type indexedNumbers []numberRec

func (a indexedNumbers) Len() int           { return len(a) }
func (a indexedNumbers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a indexedNumbers) Less(i, j int) bool { return a[i].value < a[j].value }

func (cef Cef) SortNumerical(by string, reverse bool) (*Cef, error) {
	recs := make([]numberRec, cef.NumRows)

	temp := strings.Split(by, "=")
	if len(temp) == 2 {
		// Find the column attribute
		colAttr := -1
		for i := 0; i < len(cef.ColumnAttributes); i++ {
			if cef.ColumnAttributes[i].Name == temp[0] {
				colAttr = i
				break
			}
		}
		if colAttr == -1 {
			return nil, errors.New("Column attribute not found when attempting to select: " + temp[0])
		}

		// Find the column that matches the value
		col := -1
		for i := 0; i < cef.NumColumns; i++ {
			if cef.ColumnAttributes[colAttr].Values[i] == temp[1] {
				col = i
				break
			}
		}
		if col == -1 {
			return nil, errors.New("Column attribute value not found when attempting to select: " + temp[1])
		}

		// Make the list of values
		for i := 0; i < cef.NumRows; i++ {
			recs[i] = numberRec{cef.Get(col, i), i}
		}
	} else {
		// Find the indexes
		var index []string
		for i := 0; i < len(cef.RowAttributes); i++ {
			if cef.RowAttributes[i].Name == by {
				index = cef.RowAttributes[i].Values
			}
		}
		if index == nil {
			return nil, errors.New("Attribute not found when attempting to sort: " + by)
		}

		// Collect the values to be sorted
		for i := 0; i < len(index); i++ {
			value, err := strconv.ParseFloat(index[i], 32)
			if err != nil {
				value = 0
			}
			recs[i] = numberRec{float32(value), i}
		}
	}

	// Sort them
	sort.Sort(indexedNumbers(recs))

	// Make the resulting Cef
	result := new(Cef)
	result.NumColumns = cef.NumColumns
	result.NumRows = cef.NumRows
	result.Headers = cef.Headers
	result.Flags = cef.Flags
	result.Matrix = make([]float32, 0)
	result.ColumnAttributes = cef.ColumnAttributes
	result.RowAttributes = make([]Attribute, len(cef.RowAttributes))
	for i := 0; i < len(cef.RowAttributes); i++ {
		result.RowAttributes[i].Name = cef.RowAttributes[i].Name
		result.RowAttributes[i].Values = make([]string, len(result.RowAttributes[i].Values))
	}
	if reverse {
		for i := cef.NumRows - 1; i >= 0; i-- {
			from := recs[i].index
			result.Matrix = append(result.Matrix, cef.GetRow(from)...)
			for j := 0; j < len(cef.RowAttributes); j++ {
				result.RowAttributes[j].Values = append(result.RowAttributes[j].Values, cef.RowAttributes[j].Values[from])
			}
		}
	} else {
		for i := 0; i < cef.NumRows; i++ {
			from := recs[i].index
			result.Matrix = append(result.Matrix, cef.GetRow(from)...)
			for j := 0; j < len(cef.RowAttributes); j++ {
				result.RowAttributes[j].Values = append(result.RowAttributes[j].Values, cef.RowAttributes[j].Values[from])
			}
		}
	}
	return result, nil
}

// Join performs a database-style join of two Cef instances, by
// lining up rows that have the same value for the given attributes.
// The 'mode' parameter determines the type of join performed: left join (mode "left"),
// right join (mode "right") or inner join (mode "inner")
func (left Cef) Join(right *Cef, leftAttr string, rightAttr string) (*Cef, error) {
	// Find the indexes
	var leftIndex []string
	for i := 0; i < len(left.RowAttributes); i++ {
		if left.RowAttributes[i].Name == leftAttr {
			leftIndex = left.RowAttributes[i].Values
		}
	}
	var rightIndex []string
	for i := 0; i < len(right.RowAttributes); i++ {
		if right.RowAttributes[i].Name == rightAttr {
			rightIndex = right.RowAttributes[i].Values
		}
	}
	if rightIndex == nil || leftIndex == nil {
		return nil, errors.New("Index not found when attempting to join " + leftAttr + " " + rightAttr)
	}

	leftKeys := map[string]int{}

	// Hash the keys of the left table, pointing to the corresponding row index
	for i := 0; i < len(leftIndex); i++ {
		if leftKeys[leftIndex[i]] == 0 { // Don't add keys twice, to ensure the join will prefer earlier rows
			leftKeys[leftIndex[i]] = i + 1 // Store as index + 1, so that we can distinguish the zero value
		}
	}

	// Prepare the result
	result := new(Cef)
	result.NumColumns = left.NumColumns + right.NumColumns
	result.Headers = left.Headers
	result.Flags = left.Flags
	result.Matrix = make([]float32, 0)
	result.ColumnAttributes = make([]Attribute, len(left.ColumnAttributes)+len(right.ColumnAttributes))
	// Make empty column attributes
	for i := 0; i < len(left.ColumnAttributes); i++ {
		result.ColumnAttributes[i].Name = left.ColumnAttributes[i].Name
		result.ColumnAttributes[i].Values = make([]string, 0)
	}
	for i := 0; i < len(right.ColumnAttributes); i++ {
		result.ColumnAttributes[i+len(left.ColumnAttributes)].Name = right.ColumnAttributes[i].Name
		result.ColumnAttributes[i+len(left.ColumnAttributes)].Values = make([]string, 0)
	}
	// Make empty row attributes
	result.RowAttributes = make([]Attribute, len(left.RowAttributes)+len(right.RowAttributes))
	for i := 0; i < len(left.RowAttributes); i++ {
		result.RowAttributes[i].Name = left.RowAttributes[i].Name
		result.RowAttributes[i].Values = make([]string, 0)
	}
	for i := 0; i < len(right.RowAttributes); i++ {
		result.RowAttributes[i+len(left.RowAttributes)].Name = right.RowAttributes[i].Name
		result.RowAttributes[i+len(left.RowAttributes)].Values = make([]string, 0)
	}

	// Join the column attributes
	for j := 0; j < len(left.ColumnAttributes); j++ {
		result.ColumnAttributes[j].Values = append(result.ColumnAttributes[j].Values, left.ColumnAttributes[j].Values...)
		result.ColumnAttributes[j].Values = append(result.ColumnAttributes[j].Values, make([]string, right.NumColumns)...)
	}
	for j := 0; j < len(right.ColumnAttributes); j++ {
		result.ColumnAttributes[j+len(left.ColumnAttributes)].Values = append(result.ColumnAttributes[j+len(left.ColumnAttributes)].Values, make([]string, left.NumColumns)...)
		result.ColumnAttributes[j+len(left.ColumnAttributes)].Values = append(result.ColumnAttributes[j+len(left.ColumnAttributes)].Values, right.ColumnAttributes[j].Values...)
	}

	// For each row of the right table, look it up in the hash
	numRows := 0
	for i := 0; i < len(rightIndex); i++ {
		ix := leftKeys[rightIndex[i]]
		if ix != 0 {
			// We have a match; append one row to the result
			numRows++
			leftKeys[rightIndex[i]] = 0 // Delete the key to prevent future matches (skip if doing right join?)
			result.Matrix = append(result.Matrix, left.GetRow(ix-1)...)
			result.Matrix = append(result.Matrix, right.GetRow(i)...)

			// Append to the row attributes
			for j := 0; j < len(left.RowAttributes); j++ {
				result.RowAttributes[j].Values = append(result.RowAttributes[j].Values, left.RowAttributes[j].Values[ix-1])
			}
			for j := 0; j < len(right.RowAttributes); j++ {
				result.RowAttributes[j+len(left.RowAttributes)].Values = append(result.RowAttributes[j+len(left.RowAttributes)].Values, right.RowAttributes[j].Values[i])
			}
		} else {
			fmt.Fprintf(os.Stderr, "Dropped %v", rightIndex[i])
		}
	}
	result.NumRows = numRows

	// Merge duplicate column attributes
	temp := make([]Attribute, 0)
	for i := 0; i < len(result.ColumnAttributes); i++ {
		// Check if this attribute has already been appended
		found := false
		for j := 0; j < i; j++ {
			if result.ColumnAttributes[i].Name == result.ColumnAttributes[j].Name {
				found = true
				// Merge values
				for k := 0; k < len(result.ColumnAttributes[j].Values); k++ {
					if result.ColumnAttributes[j].Values[k] == "" {
						result.ColumnAttributes[j].Values[k] = result.ColumnAttributes[i].Values[k]
					}
				}
				break
			}
		}
		if !found {
			temp = append(temp, result.ColumnAttributes[i])
		}
	}
	result.ColumnAttributes = temp

	// Drop duplicate row attributes
	temp = make([]Attribute, 0)
	for i := 0; i < len(result.RowAttributes); i++ {
		// Check if this attribute has already been appended
		found := false
		for j := 0; j < i; j++ {
			if result.RowAttributes[i].Name == result.RowAttributes[j].Name {
				found = true
				break
			}
		}
		if !found {
			temp = append(temp, result.RowAttributes[i])
		}
	}
	result.RowAttributes = temp

	// Merge headers
	for _, hdr := range right.Headers {
		var found bool
		for _, existing := range left.Headers {
			if existing.Name == hdr.Name && existing.Value == hdr.Value {
				found = true
				break
			}
		}
		if !found {
			result.Headers = append(result.Headers, hdr)
		}
	}

	return result, nil
}
