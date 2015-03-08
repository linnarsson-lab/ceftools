package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/slinnarsson/ceftools"
	"os"
)

func main() {
	// Define the command-line structure using Kingpin
	var versionString = fmt.Sprintf("ceftools v%v.%v (C) 2015 Sten Linnarsson <http://linnarssonlab.org/>", ceftools.MajorVersion, ceftools.MinorVersion)
	var app = kingpin.New("cef", versionString)
	var app_cef = app.Flag("cef", "Generate CEF as output, instead of CEB").Bool()
	var app_transpose = app.Flag("transpose", "Transpose input before, and/or output after processing").Short('t').Default("none").Enum("none", "in", "out", "inout")
	var info = app.Command("info", "Show a summary of the file contents")
	var test = app.Command("test", "Perform an internal test")
	//	var join = app.Command("join", "Join two files based on an attribute used as key")
	//	var join_other = join.Flag("other", "The file to which <STDIN> should be joined").Required().String()

	// Parse the command line
	var parsed, err = app.Parse(os.Args[1:])
	if err != nil {
		app.Usage(os.Stderr)
		return
	}

	// Handle the sub-commands
	switch kingpin.MustParse(parsed, nil) {

	// Perform test
	case test.FullCommand():
		var cef = new(ceftools.Cef)
		cef.NumColumns = 5
		cef.NumRows = 10
		cef.MajorVersion = 0
		cef.MinorVersion = 1
		cef.Headers = make([]ceftools.Header, 2)
		cef.Headers[0].Name = "Header 1"
		cef.Headers[0].Value = "Header value 1"
		cef.Headers[1].Name = "Header 2"
		cef.Headers[1].Value = "Header value 2"
		cef.ColumnAttributes = make([]ceftools.Attribute, 2)
		cef.ColumnAttributes[0].Name = "CellID"
		cef.ColumnAttributes[0].Values = make([]string, 5)
		cef.ColumnAttributes[1].Name = "Well"
		cef.ColumnAttributes[1].Values = make([]string, 5)
		cef.RowAttributes = make([]ceftools.Attribute, 2)
		cef.RowAttributes[0].Name = "Gene"
		cef.RowAttributes[0].Values = make([]string, 10)
		cef.RowAttributes[1].Name = "Chromosome"
		cef.RowAttributes[1].Values = make([]string, 10)
		cef.Matrix = make([]float32, 10*5)
		ceftools.WriteAsCEF(cef, os.Stdout, (*app_transpose == "inout") || (*app_transpose == "out"))
		return

	// Show info
	case info.FullCommand():
		var cef, err = ceftools.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"), true)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		fmt.Fprintln(os.Stderr, "Version: %v.%v", cef.MajorVersion, cef.MinorVersion)
		fmt.Fprintln(os.Stderr, "Columns: %v", cef.NumColumns)
		fmt.Fprintln(os.Stderr, "Rows: %v", cef.NumRows)
		fmt.Fprintln(os.Stderr, "Flags: %v", cef.Flags)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Headers:")
		for i := 0; i < len(cef.Headers); i++ {
			fmt.Fprint(os.Stderr, "  ")
			fmt.Fprint(os.Stderr, cef.Headers[i].Name)
			fmt.Fprint(os.Stderr, " = ")
			fmt.Fprintln(os.Stderr, cef.Headers[i].Value)
		}
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, "Column attributes: ")
		for i := 0; i < len(cef.ColumnAttributes); i++ {
			fmt.Fprint(os.Stderr, cef.ColumnAttributes[i].Name)
			if i != (len(cef.ColumnAttributes) - 1) {
				fmt.Fprint(os.Stderr, ", ")
			}
		}
		fmt.Fprint(os.Stderr, "Row attributes: ")
		for i := 0; i < len(cef.RowAttributes); i++ {
			fmt.Fprint(os.Stderr, cef.RowAttributes[i].Name)
			if i != (len(cef.RowAttributes) - 1) {
				fmt.Fprint(os.Stderr, ", ")
			}
		}

		if *app_cef {
		}
		// 	cef.WriteAsCEF(cf, os.Stdout, true)
		// } else {
		// 	cef.WriteAsCEB(cf, os.Stdout, true)
		// }
	}
}