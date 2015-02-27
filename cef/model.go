package cef

type Annotation struct {
	Name   string
	Values []string
}

type CefFile struct {
	RowAnnotations    []Annotation
	ColumnAnnotations []Annotation
	Rows              int
	Columns           int
	Matrix            []float32
}
