package cef

type CefAttribute struct {
	Name   string
	Values []string
}

type CefFile struct {
	NumRows          int64
	NumColumns       int64
	RowAttributes    []CefAttribute
	ColumnAttributes []CefAttribute
	Matrix           []float32
}
