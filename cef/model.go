package cef

type CefAttribute struct {
	Name   string
	Values []string
}

type CefHeader struct {
	Name  string
	Value string
}

type CefFile struct {
	NumRows          int64
	NumColumns       int64
	Headers          []CefHeader
	RowAttributes    []CefAttribute
	ColumnAttributes []CefAttribute
	Matrix           []float32
}
