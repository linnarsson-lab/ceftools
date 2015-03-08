package cef

type CefAttribute struct {
	Name   string
	Values []string
}

type CefHeader struct {
	Name  string
	Value string
}

const (
	Transposed = 1 << iota
)

type CefFile struct {
	MajorVersion     int32
	MinorVersion     int32
	NumRows          int64
	NumColumns       int64
	Headers          []CefHeader
	Flags            int64
	RowAttributes    []CefAttribute
	ColumnAttributes []CefAttribute
	Matrix           []float32
}

func (cf CefFile) Get(col int64, row int64) float32 {
	return cf.Matrix[col+row*cf.NumColumns]
}

func (cf CefFile) Set(col int64, row int64, val float32) {
	cf.Matrix[col+row*cf.NumColumns] = val
}
