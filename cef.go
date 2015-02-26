package main

import (
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

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Transpose file
	case transpose.FullCommand():
		println("Transposing the CEF/CEB file...")
		if *app_input == "" {
			println("Reading from STDIN")
		} else {
			println("Reading from " + *app_input)
		}
		if *app_output == "" {
			println("Writing to STDOUT")
		} else {
			println("Writing to " + *app_output)
		}
		if *app_cef {
			println("Generating CEF file instead of CEB")
		}
	}
}
