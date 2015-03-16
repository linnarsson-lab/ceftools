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
	var app_transpose = app.Flag("transpose", "Transpose matrix (in, out, inout or none)").Short('t').Default("none").Enum("none", "in", "out", "inout")

	var info = app.Command("info", "Show a summary of the file contents")
	var test = app.Command("test", "Perform an internal test")
	var export = app.Command("export", "Export the input (CEF or CEB) as text-based CEF")
	var cmdimport = app.Command("import", "Import from a legacy format (default: CEF or CEB) and generate a CEB file output")
	var import_strt = cmdimport.Flag("strt", "Import a STRT '_expression.tab' file").Bool()

	var drop = app.Command("drop", "Remove attributes")
	var drop_attrs = drop.Flag("attrs", "Row attribute(s) to remove (case-sensitive, comma-separated)").Short('a').Required().String()
	var drop_except = drop.Flag("except", "Keep the given attributes instead of dropping them ").Bool()
	var cmdselect = app.Command("select", "Select rows that match criteria (and drop the rest)")
	var select_rows = cmdselect.Flag("range", "Select a range of rows (colon-separated, 1-based)").String()
	var select_where = cmdselect.Flag("where", "Select rows with specific value for attribute ('attr=value')").String()
	var select_except = cmdselect.Flag("except", "Invert selection").Bool()

	var rescale = app.Command("rescale", "Rescale values by rows")
	var rescale_method = rescale.Flag("method", "Method to use (log, tpm or rpkm)").Short('m').Required().Enum("log", "tpm", "rpkm")
	var rescale_length = rescale.Flag("length", "Indicate the name of the attribute that gives gene length (for RPKM)").String()

	var join = app.Command("join", "Join two files based on given attributes")
	var join_other = join.Flag("with", "The file to which the input should be joined").Required().String()
	var join_on = join.Flag("on", "The attributes on which to join, of form 'attr1=attr2'").Required().String()

	var sort = app.Command("sort", "Sort by row attribute or by specific column")
	var sort_by = sort.Flag("by", "The attribute or column ('column=value') to sort by").Required().String()
	var sort_reverse = sort.Flag("reverse", "Sort in reverse order").Short('r').Bool()
	var sort_numerical = sort.Flag("numerical", "Numerical sort (default: alphabetical)").Short('n').Bool()
	//	var sort_bycvmean = sort.Flag("cvmean", "Sort by offset to a CV-vs-mean least-squares fit").Bool()

	// Parse the command line
	var parsed, err = app.Parse(os.Args[1:])
	if err != nil {
		app.Usage(os.Stderr)
		return
	}

	// Handle the sub-commands
	switch kingpin.MustParse(parsed, nil) {
	case sort.FullCommand():
		if err = ceftools.CmdSort(*sort_by, *sort_numerical, *sort_reverse, *app_transpose); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	case join.FullCommand():
		if err = ceftools.CmdJoin(*join_other, *join_on, *app_transpose); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	case cmdimport.FullCommand():
		if *import_strt {
			if err = ceftools.CmdImportStrt(*app_transpose); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			if err = ceftools.CmdImport(*app_transpose); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		return
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
		if err = ceftools.CmdDrop(*drop_attrs, *drop_except, *app_transpose); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	case cmdselect.FullCommand():
		print(select_except)
		print(select_rows)
		print(select_where)
		return
	case rescale.FullCommand():
		if err = ceftools.CmdRescale(*rescale_method, *rescale_length, *app_transpose); err != nil {
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
		cef.ColumnAttributes[0].Values = []string{"A", "B", "C", "D", "E"}
		cef.ColumnAttributes[1].Name = "Well"
		cef.ColumnAttributes[1].Values = []string{"A01", "B03", "C09", "D12", "E21"}
		cef.RowAttributes = make([]ceftools.Attribute, 3)
		cef.RowAttributes[0].Name = "Gene"
		cef.RowAttributes[0].Values = []string{"Actb", "Gapdh", "Synpr", "Pou3f2", "Bdnf", "Ngf", "Sox11", "Olig1", "Olig2", "Foxj1"}
		cef.RowAttributes[1].Name = "Chromosome"
		cef.RowAttributes[1].Values = []string{"Chr1", "Chr2", "Chr3", "Chr4", "Chr5", "Chr1", "Chr2", "Chr3", "Chr4", "Chr5"}
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
