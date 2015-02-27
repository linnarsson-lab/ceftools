package main

import (
	"fmt"
	"github.com/slinnarsson/Cellophane/cef"
	"gopkg.in/alecthomas/kingpin.v1"
	"os"
)

func main() {
	// Define the command-line structure using Kingpin
	var app = kingpin.New("cef", "Cellophane v0.1 (C) 2015 Sten Linnarsson <http://linnarssonlab.org/>")
	var app_input = app.Flag("input", "Name of file to use as input (default: use STDIN)").String()
	var app_output = app.Flag("output", "Name of file to use as output (default: use STDOUT)").String()
	var app_cef = app.Flag("cef", "Generate CEF as output (default: generate CEB)").Bool()
	var transpose = app.Command("transpose", "Transpose the file")
	//	var info = app.Command("info", "Show a summary of the file contents")
	//	var join = app.Command("join", "Join two files based on an attribute used as key")
	//	var join_other = join.Flag("other", "The file to which <STDIN> should be joined").Required().String()

	// Parse the command line
	var parsed, err = app.Parse(os.Args[1:])
	if err != nil {
		app.Usage(os.Stderr)
		return
	}

	// Open the input and output streams
	var input = os.Stdin
	var output = os.Stdout
	if *app_input != "" {
		input, err = os.Open(*app_input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		defer input.Close()
	}
	if *app_output != "" {
		output, err = os.Create(*app_output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		defer output.Close()
	}

	// Handle the sub-commands
	switch kingpin.MustParse(parsed, nil) {

	// Transpose file
	case transpose.FullCommand():
		fmt.Fprintln(os.Stderr, "Transposing the CEF/CEB file...")
		if *app_cef {
			fmt.Fprintln(os.Stderr, "Generating CEF file instead of CEB")
		}
		var cef, err = cef.Read(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		output.WriteString(cef.ColumnAnnotations[0].Name)
	}
}
