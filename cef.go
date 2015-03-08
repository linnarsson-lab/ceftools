package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/slinnarsson/ceftools/cef"
	"os"
)

func main() {
	// Define the command-line structure using Kingpin
	var app = kingpin.New("cef", "ceftools v0.1 (C) 2015 Sten Linnarsson <http://linnarssonlab.org/>")
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
		var cf = new(cef.CefFile)
		cf.NumColumns = 10
		cf.NumRows = 10
		cf.MajorVersion = 0
		cf.MinorVersion = 1
		cf.Headers = make([]cef.CefHeader, 2)
		cf.Headers[0].Name = "Header 1"
		cf.Headers[0].Value = "Header value 1"
		cf.Headers[1].Name = "Header 2"
		cf.Headers[1].Value = "Header value 2"
		cf.ColumnAttributes = make([]cef.CefAttribute, 2)
		cf.ColumnAttributes[0].Name = "CellID"
		cf.ColumnAttributes[0].Values = make([]string, 10)
		cf.ColumnAttributes[1].Name = "Well"
		cf.ColumnAttributes[1].Values = make([]string, 10)
		cf.RowAttributes = make([]cef.CefAttribute, 2)
		cf.RowAttributes[0].Name = "Gene"
		cf.RowAttributes[0].Values = make([]string, 10)
		cf.RowAttributes[1].Name = "Chromosome"
		cf.RowAttributes[1].Values = make([]string, 10)
		cf.Matrix = make([]float32, 10*10)
		cef.WriteAsCEF(cf, os.Stdout, false)
		return

	// Show info
	case info.FullCommand():
		var cf, err = cef.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"), true)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		fmt.Fprintln(os.Stderr, "Version: %v.%v", cf.MajorVersion, cf.MinorVersion)
		fmt.Fprintln(os.Stderr, "Columns: %v", cf.NumColumns)
		fmt.Fprintln(os.Stderr, "Rows: %v", cf.NumRows)
		fmt.Fprintln(os.Stderr, "Flags: %v", cf.Flags)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Headers:")
		for i := 0; i < len(cf.Headers); i++ {
			fmt.Fprint(os.Stderr, "  ")
			fmt.Fprint(os.Stderr, cf.Headers[i].Name)
			fmt.Fprint(os.Stderr, " = ")
			fmt.Fprintln(os.Stderr, cf.Headers[i].Value)
		}
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, "Column attributes: ")
		for i := 0; i < len(cf.ColumnAttributes); i++ {
			fmt.Fprint(os.Stderr, cf.ColumnAttributes[i].Name)
			if i != (len(cf.ColumnAttributes) - 1) {
				fmt.Fprint(os.Stderr, ", ")
			}
		}
		fmt.Fprint(os.Stderr, "Row attributes: ")
		for i := 0; i < len(cf.RowAttributes); i++ {
			fmt.Fprint(os.Stderr, cf.RowAttributes[i].Name)
			if i != (len(cf.RowAttributes) - 1) {
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
