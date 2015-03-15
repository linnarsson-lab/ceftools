package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/slinnarsson/ceftools"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Define the command-line structure using Kingpin
	var versionString = fmt.Sprintf("ceftools v%v.%v (C) 2015 Sten Linnarsson <http://linnarssonlab.org/>", ceftools.MajorVersion, ceftools.MinorVersion)

	var app = kingpin.New("cef", versionString)
	var app_transpose = app.Flag("transpose", "Transpose matrix (in, out, inout or none)").Short('t').Default("none").Enum("none", "in", "out", "inout")

	var info = app.Command("info", "Show a summary of the file contents")
	var test = app.Command("test", "Perform an internal test")
	var export = app.Command("export", "Export the file as text-based CEF")

	var drop = app.Command("drop", "Remove attributes")
	var drop_attrs = drop.Flag("attrs", "Row attribute(s) to remove (case-sensitive, comma-separated)").Short('a').Required().String()
	var drop_except = drop.Flag("except", "Keep the given attributes instead of dropping them ").Bool()
	var cmdselect = app.Command("select", "Select rows that match criteria (and drop the rest)")
	var select_rows = cmdselect.Flag("range", "Select a range of rows (colon-separated, 1-based)").String()
	var select_where = cmdselect.Flag("where", "Select rows with specific value for attribute ('attr=value')").String()
	var select_not = cmdselect.Flag("not", "Invert selection").Bool()

	var rescale = app.Command("rescale", "Rescale values by rows")
	var rescale_method = rescale.Flag("method", "Method to use (log, tpm or rpkm)").Short('m').Required().Enum("log", "tpm", "rpkm")
	var rescale_length = rescale.Flag("length", "Indicate the name of the attribute that gives gene length (for RPKM)").String()

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
	case export.FullCommand():
		// Read the input
		var cef, err = ceftools.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		// Write the CEF file
		if err := ceftools.WriteAsCEF(cef, os.Stdout, (*app_transpose == "inout") || (*app_transpose == "out")); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	case drop.FullCommand():
		// Read the input
		var cef, err = ceftools.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		contains := func(s []string, e string) bool {
			for _, a := range s {
				if a == e {
					return true
				}
			}
			return false
		}

		// Drop the attributes
		todrop := strings.Split(*drop_attrs, ",")
		temp := cef.RowAttributes[:0]
		for _, att := range cef.RowAttributes {
			if contains(todrop, att.Name) == *drop_except {
				temp = append(temp, att)
			}
		}
		cef.RowAttributes = temp

		// Write the result
		if err := ceftools.WriteAsCEB(cef, os.Stdout, (*app_transpose == "inout") || (*app_transpose == "out")); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	case cmdselect.FullCommand():
		print(select_not)
		print(select_rows)
		print(select_where)
		return
	case rescale.FullCommand():
		// Read the input
		var cef, err = ceftools.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		log_rescale := func(vals []float32) {
			for i := 0; i < len(vals); i++ {
				vals[i] = float32(math.Log10(float64(vals[i] + 1)))
			}
		}
		tpm_rescale := func(vals []float32) {
			sum := float32(0)
			for i := 0; i < len(vals); i++ {
				sum += vals[i]
			}
			if sum != 0 {
				for i := 0; i < len(vals); i++ {
					vals[i] = vals[i] * 1000000 / sum
				}
			}
		}
		rpkm_rescale := func(vals []float32, length float32) {
			sum := float32(0)
			for i := 0; i < len(vals); i++ {
				sum += vals[i]
			}
			if sum != 0 {
				for i := 0; i < len(vals); i++ {
					vals[i] = vals[i] * 1000000 / sum / length
				}
			}
		}
		var length []string
		if *rescale_length != "" {
			for i := 0; i < len(cef.RowAttributes); i++ {
				if cef.RowAttributes[i].Name == *rescale_length {
					length = cef.RowAttributes[i].Values
				}
			}
			if length == nil {
				panic("Length attribute not found when attempting to rescale by rpkm")
			}
		}
		for i := int64(0); i < cef.NumRows; i++ {
			switch *rescale_method {
			case "log":
				log_rescale(cef.GetRow(i))
				break
			case "tpm":
				tpm_rescale(cef.GetRow(i))
			case "rpkm":
				bp, err := strconv.Atoi(length[i])
				if err != nil {
					panic("Length attribute was not a valid integer (when attempting to rescale by rpkm)")
				}
				rpkm_rescale(cef.GetRow(i), float32(bp)/1000)
			}
		}

		// Write the result
		if err := ceftools.WriteAsCEB(cef, os.Stdout, (*app_transpose == "inout") || (*app_transpose == "out")); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return

	// Perform test
	case test.FullCommand():
		var cef = new(ceftools.Cef)
		cef.NumColumns = 5
		cef.NumRows = 10
		cef.MajorVersion = 0
		cef.MinorVersion = 1
		cef.Headers = make([]ceftools.Header, 2)
		cef.Headers[0].Name = "Tissue"
		cef.Headers[0].Value = "Amygdala"
		cef.Headers[1].Name = "Species"
		cef.Headers[1].Value = "Mouse"
		cef.ColumnAttributes = make([]ceftools.Attribute, 2)
		cef.ColumnAttributes[0].Name = "CellID"
		cef.ColumnAttributes[0].Values = make([]string, 5)
		cef.ColumnAttributes[1].Name = "Well"
		cef.ColumnAttributes[1].Values = make([]string, 5)
		cef.RowAttributes = make([]ceftools.Attribute, 3)
		cef.RowAttributes[0].Name = "Gene"
		cef.RowAttributes[0].Values = make([]string, 10)
		cef.RowAttributes[1].Name = "Chromosome"
		cef.RowAttributes[1].Values = make([]string, 10)
		cef.RowAttributes[2].Name = "Length"
		cef.RowAttributes[2].Values = []string{"1200", "1300", "1400", "1700", "1920", "130", "800", "7800", "1100", "200"}
		cef.Matrix = make([]float32, 10*5)
		cef.Set(0, 0, 1)
		cef.Set(0, 1, 2)
		cef.Set(0, 2, 3)
		if err := ceftools.WriteAsCEB(cef, os.Stdout, (*app_transpose == "inout") || (*app_transpose == "out")); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return

	// Show info
	case info.FullCommand():
		var cef, err = ceftools.Read(os.Stdin, (*app_transpose == "inout") || (*app_transpose == "in"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		fmt.Fprintf(os.Stderr, "          Version: %v.%v\n", cef.MajorVersion, cef.MinorVersion)
		fmt.Fprintf(os.Stderr, "          Columns: %v\n", cef.NumColumns)
		fmt.Fprintf(os.Stderr, "             Rows: %v\n", cef.NumRows)
		fmt.Fprintf(os.Stderr, "            Flags: %v\n", cef.Flags)
		fmt.Fprintln(os.Stderr, "          Headers:")
		for i := 0; i < len(cef.Headers); i++ {
			fmt.Fprint(os.Stderr, "                   ")
			fmt.Fprint(os.Stderr, cef.Headers[i].Name)
			fmt.Fprint(os.Stderr, " = ")
			fmt.Fprintf(os.Stderr, cef.Headers[i].Value)
			fmt.Fprint(os.Stderr, "\n")
		}
		fmt.Fprint(os.Stderr, "\n")
		fmt.Fprint(os.Stderr, "Column attributes: ")
		for i := 0; i < len(cef.ColumnAttributes); i++ {
			fmt.Fprint(os.Stderr, cef.ColumnAttributes[i].Name)
			if i != (len(cef.ColumnAttributes) - 1) {
				fmt.Fprint(os.Stderr, ", ")
			}
		}
		fmt.Fprint(os.Stderr, "\n")
		fmt.Fprint(os.Stderr, "   Row attributes: ")
		for i := 0; i < len(cef.RowAttributes); i++ {
			fmt.Fprint(os.Stderr, cef.RowAttributes[i].Name)
			if i != (len(cef.RowAttributes) - 1) {
				fmt.Fprint(os.Stderr, ", ")
			}
		}
		fmt.Fprintln(os.Stderr, "")
	}
}
