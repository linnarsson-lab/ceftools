package main

import (
	"fmt"
	"github.com/slinnarsson/ceftools/cef"
	"gopkg.in/alecthomas/kingpin.v1"
	"os"
)

func main() {
	// Define the command-line structure using Kingpin
	var app = kingpin.New("cef", "ceftools v0.1 (C) 2015 Sten Linnarsson <http://linnarssonlab.org/>")
	var app_cef = app.Flag("cef", "Generate CEF as output, instead of CEB").Bool()
	var app_transpose = app.Flag("transpose", "Transpose input before, and/or output after processing").Short('t').Default("none").Enum("none", "in", "out", "inout")
	var info = app.Command("info", "Show a summary of the file contents")
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

	// Transpose file
	case info.FullCommand():
		var cf, err = cef.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"))
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

		// if *app_cef {
		// 	cef.WriteAsCEF(cf, os.Stdout, true)
		// } else {
		// 	cef.WriteAsCEB(cf, os.Stdout, true)
		// }
	}
}
