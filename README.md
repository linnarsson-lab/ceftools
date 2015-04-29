# ceftools

Tools for manipulating cell expression format (CEF) files.

You can download the latest version from the [releases page](https://github.com/linnarsson-lab/ceftools/releases).

## Overview

ceftools is a set of tools designed to manipulate large-scale gene expression data. It aims to be for gene 
expression data what [samtools](http://samtools.github.io) is for sequence data. It was designed to simplify the exchange
and manipulation of very large-scale transcriptomics data, particularly from single-cell RNA-seq. However, it can process
any omics-style dataset that can be represented as an annotated matrix of numbers.

ceftools is implemented as a command-line utility called `cef`, which operates on an annotated matrix of gene-expression 
values. The annotation consists of headers (name-value pairs) as well as row and column attributes. It supports useful operations such as filtering, sorting, splitting, joining and transposing the input. Multiple commands can be chained to perform more complex operations.

ceftools processes files in the text-based CEF format ('cell expression format', `.cef`). CEF files are human-readable, tab-delimited 
text files that can be easily parsed and generated from any scripting language. 

## Synopsis

	cef help			- print help for the cef command
	cef info            - overview of file contents
	cef view			- interactively navigate the matrix
	cef transpose 		- transpose the file
	cef sort			- sort by row attribute or by specific column
	cef sort --spin 	- sort rows by the SPIN algorithm
	cef select			- select rows that match given criteria
	cef join		  	- join two datasets by given attributes
	cef add 			- add attribute or header with constant value 
	cef drop 			- drop attribute(s) or header(s)
	cef rename			- rename attribute
	cef rescale			- rescale rows (rpkm, tpm or log-transformed)
	cef aggregate		- calculate aggregate statistics for every row
	cef import			- import from STRT


## Commands

The examples below make use of the `oligos.cef` sample dataset, which you can download from the [BackSPIN release page](https://github.com/linnarsson-lab/BackSPIN/releases). If you have installed `cef` and have `oligos.cef` in your current working directory, you should be able to run all the examples below without modification.

Commands operate on rows by default. For example `drop` can be used to remove row attributes, but not column attributes. Use the global `--bycol` flag to operate instead on columns. For example, to remove column attribute `Gene` then sort on column attribute `Length`:

```
< infile.cef cef --bycol drop --attrs Gene | cef --bycol sort Length > outfile.cef 
```

Note that since `--bycol` is a global flag it must always be positioned before the command: `cef --bycol <command>`


### Info

Show a summary of the contents of a CEF file.

Synopsis:

	cef info

Example:

	< oligos.cef cef info 

Output:

	          Columns: 820
	             Rows: 19972
	            Flags: 0
	          Headers:
	                   Genome = mm10
	                   Citation = http://www.sciencemag.org/content/347/6226/1138.abstract

	Column attributes: Tissue, Group, Total_mRNA, Well, Sex, Age, Diameter, CellID, Class, Subclass
	   Row attributes: GeneType, Gene, GeneGroup


### View

Interactively view the contents of a CEF file (in the terminal window).

Synopsis:

	cef view

Example: 
	
	< oligos.cef cef view

Ouput:

![Viewer screen](viewer-oligos.png)

The yellow toolbar at the bottom shows the available commands for navigating the file.

Use the 'wasd' keys to scroll the matrix; hold down shift to scroll by a whole screen at a time. Press 'h' to jump to the top-left corner, and 'z' to jump to the bottom-right. 

Use the arrow keys to scroll the entire view (including the row and column attributes).

To sort by an attribute or column, use the arrow keys to scroll the view. Position the column you want to sort by at the left of you screen, then press 'o'. To reverse the sort order, press 'o' again.

To transpose rows and columns, press 't'.

Press 'q' to exit the viewer.


### Transpose

Transpose rows and columns.

Synopsis:

	cef transpose

Example:

	< oligos.cef cef transpose | cef info

Output:

	          Columns: 19972
	             Rows: 820
	            Flags: 0
	          Headers:
	                   Genome = mm10
	                   Citation = http://www.sciencemag.org/content/347/6226/1138.abstract

	Column attributes: GeneType, Gene, GeneGroup
	   Row attributes: Tissue, Group, Total_mRNA, Well, Sex, Age, Diameter, CellID, Class, Subclass

Compare this output to the example given above (*Info* command).


### Sort

Sort the file based on a row attribute or the values in a specific column.

Synopsis:

	cef sort --by "attr=X"		Sort by the column where column attribute 'attr' has value 'X'
	cef sort --by "attr"		Sort by the row attribute 'attr'

	Options:

		--numerical				When sorting by row attribute, sort numerically (default: sort alphabetically)
		--reverse				Sort in reverse order

Example:

	< oligos.cef cef sort --by "CellID=1772067057_G07" --reverse > oligos_sorted.cef

Output:

The output file is sorted by the first column, which has CellID '1772067057_G07', in descending order. You can verify this by doing `< oligos_sorted.cef cef view`.

**Note:** when sorting by column, the "attr=X" clause must be put in double quotes (as above), or bash will interpret the equals sign as a variable assignment.


### Sort (SPIN)

Sort rows based on correlations using the [SPIN algorithm](http://bioinformatics.oxfordjournals.org/content/21/10/2301.full)

Synopsis:

	cef sort --spin 			Sort rows by SPIN

	Options:

		--corrfile "FILE"		Write the correlation matrix to the given file

Example:

	< oligos.cef cef sort --spin --corrfile oligos_spin_corrs.tab > oligos_spin_sorted.cef

Output:

The output file is sorted by correlation similarity, and the row correlation matrix is written to the indicated file in tab-delimited text format.


### Select

Select rows that match a given criterion.

Synopsis:

	cef select --where "attr=X"		Select rows where the given attribute has value 'X'
	cef select --range "10:20"		Select rows 10 through 20
	cef select --range ":20"		Select all rows up to and including row 20
	cef select --range "100:"		Select all rows starting with row 100 and to the last row

	Options:

		--except					Invert the selection

Example:

	< oligos.cef cef select --where "Gene=Actb" | cef view

Output:

The result is a file with a single row, containing the data for *Actb*.


### Join

Join two datasets by matching up rows that have the same value for an attribute.

Synopsis:

	cef join --with <other.cef> --on "attr1=attr2"

This command joins two CEF files, one from standard input (STDIN) and one given by the `--with <other.cef>` option. The result is a new CEF file which has been extended on the right with all the data from 'other.cef' that matches with the existing data in the input CEF file. 

Column attributes of the same name are merged. For example, if both of the input files have column attribute `Age`, the resulting file will have a single column attribute `Age`. But if one file has `Sex` and the other has `Gender`, the output will have two attributes (`Sex` and `Gender`), and missing values will be blank.

Row attributes are merged if they contain identical values for all the retained rows. For example, if both input files have row attributes `Gene` and all retained rows have the same values in the same order, then the output will have only a single row attribute `Gene`. But if there are any differences, then the output will have two row attributes both called `Gene`.

Rows that do not match are silently dropped. 


### Add

Add a header or a constant attribute.

Synopsis:

	cef add --attr "attr=X"		Add a row attribute 'attr' with constant value 'X' in every row
	cef add --header "hdr=X"	Add a new header 'hdr' with value 'X'

Example:

	< oligos.cef cef --bycol add "File=oligos.cef" | cef view

Output:

The resulting file will have a new column attribute (since we specified `--bycol`; note that this option goes *before* the command), with the same value for every column.

This can be useful when joining two datasets, to keep track of which columns came from which file originally.


### Drop

Remove an attribute or header.

Synopsis:

	cef drop --attrs "att1,att2"	Remove attributes 'att1' and 'att2' (comma-separated, case sensitive list)
	cef drop --headers "hdr1,hdr2"	Remove headers 'hdr1' and 'hdr2' (comma-separated, case sensitive list)

Example:

	< oligos.cef cef drop --headers "Genome,Citation"

Output:

	          Columns: 820
	             Rows: 19972
	            Flags: 0
	          Headers:

	Column attributes: Tissue, Group, Total_mRNA, Well, Sex, Age, Diameter, CellID, Class, Subclass
	   Row attributes: GeneType, Gene, GeneGroup

The two headers were dropped.


### Rename

Rename an attribute.

Synopsis:

	cef rename --attr "old=new"		Rename the attribute 'old' to 'new'

Example:

	< oligos.cef cef rename --attr "Gene=Symbol" | cef info

Output:

	          Columns: 820
	             Rows: 19972
	            Flags: 0
	          Headers:
	                   Genome = mm10
	                   Citation = http://www.sciencemag.org/content/347/6226/1138.abstract

	Column attributes: Tissue, Group, Total_mRNA, Well, Sex, Age, Diameter, CellID, Class, Subclass
	   Row attributes: GeneType, Symbol, GeneGroup

Notice the renamed row attribute.


### Rescale

Scale or normalize the main matrix using one of several common methods.

Synopsis:

	cef rescale --method [log|tmp|rpkm]		Rescale using the given method

	Options:

		--length <attr>		Gives the name of the row attribute that gives the gene length (for rpkm)

Example:

	< oligos.cef cef rescale --method log | cef view

Output:

Notice that the values have been rescaled from X to log(X+1).

The 'tpm' option normalizes each row by dividing by the row sum and multiplying by 1000000.

The 'rpkm' option normalizes each row by dividing by the row sum and by the *length*, and multiplying by 1000. The *length* must be given as a row attribute, indicated using the `--length` option. The length is normally given in basepairs.


### Aggregate

Compute aggregated values for each row.

Synopsis:

	cef aggregate [options]

	Options:

		--mean				Compute mean
		--stdev				Compute standard deviation
		--cv 				Compute CV (standard deviation divided by the mean)
		--max 				Compute max value
		--min 				Compute minimum value
		--noise <method>	Compute noise as offset from CV-vs-mean fit

Example:

	< oligos.cef cef aggregate --mean --stdev | cef view

Output:

Two new row attributes are added, named "Mean" and "Stdev". This makes it possible to sort by mean and standard deviation in the viewer.

##### Calculating noise

To calculate noise using the standard CV-vs-mean fit, you must pass a *method* parameter:

`cef aggregate --noise std`

The following values are supported:

|Method | Algorithm|
|-------|----------|
|std    | Fit *log(CV) = log(mean<sup>k<sub>0</sub></sup> + k<sub>1</sub>)* using least absolute deviation |
|bands  | Fit *log(CV) = log(mean<sup>0.52</sup> + k<sub>1</sub>)* by bisection |

The most reliable and well-tested method is `std`, but `bands` can sometimes give a better fit. In both cases, the fitting algorithms 
have been tested only on data represented as absolute mRNA molecule counts per cell. If you have RPKM or TPM data, the fits may or may not 
converge and may be completely off (let us know what you find!).


### Import

Import from other file formats to CEF.

Synopsis:

	cef import --format "format"	Import a file expected to be in 'format'

Currently, this command allows a single format "strt", which can be used to import a Linnarsson lab legacy file format ("_expression.tab").



## CEF file format

CEF files are tab-delimited text files in [UTF-8](http://en.wikipedia.org/wiki/UTF-8) encoding, no [BOM](http://en.wikipedia.org/wiki/Byte_order_mark). The first four characters are 'CEF\t' (that's a single tab character at the end), equivalent to the hexadecimal 4-byte number 0x09464543 in [little-endian](http://en.wikipedia.org/wiki/Endianness) order. CEF files are guaranteed to always begin with these four bytes, which can be used to identify the file format in the absence of a file name extension.

Each row has the same number of tab-separated fields, equal to `max(7, column count + row attribute count + 1)`. In other words, the entire file is a rectangular tab-delimited matrix, with at least seven columns. CEF file *readers* should accept CEF files that have less than the required number of fields in any row, and the missing fields should be interpreted as empty strings (but empty strings should not be interpreted as zeros; thus zeros must always be explicitly represented as '0'). CEF file *writers* should always generate a rectangular tab-delimited matrix.

Each line is terminated by a single newline character: `\n`. CEF file *writers* should always generate lines ending in a single `\n`, but *readers* should silently ignore any number of adjacent newline `\n` and carriage return `\r` characters. This makes it a little easier to generate CEF files manually, e.g. in Excel.

The first row of values defines the file structure. It begins 'CEF', followed by header count, row attribute count, column attribute count, row count, column count, and the `Flags` value. 

This is followed by header lines, which are name-value pairs, with the name in the first column and the value in the second. There are no restrictions on either the names or the values. The order of headers is not necessarily preserved when CEF files are read and written. There can be multiple headers with the same name.

Next, the column attributes are given, each in a single row with an offset of `(row attribute count)`. Finally, the rows are given, starting with row attributes, and followed by the values of the main matrix. Values are represented in text as decimal floating-point numbers in scientific notation, with optional exponent (e.g. `-142.03939`, `-1.4203939e2` or `-1.4203939E+2`; the regex is `[-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?`). Values must fit in a 32-bit IEEE-754 floating point number. 

The order of headers, row and column attributes is not significant and need not be preserved. Of course, the order of the *values* of row and column attributes is significant, and must be preserved.

Example of a file with 1 header, 4 Row Attributes, 2 Column Attributes, 345 Rows, 123 Columns. The last number (0) in the first row is the `Flags` value, currently unused.

|   |   |   |   |    |    |    |
|---|---|---|---|----|----|----|
|CEF| 1 | 4 | 2 |345 |123 |  0 |
|Header name|Header value| | | | | |
|	|	|	|   |**Sex** |Male|Female|
|	|	|	|   |**Age** |P22|P28|
|**Gene**|**Chromosome**|**Position**|**Strand**|    |    |    |
|Actb|2|12837184|+|    |11 |24 |
|Nkx2-1|17|33432|-|    |0 |41 |
|   |   |   |   |    |    | ...|

Note that a CEF file can have zero row attributes, zero column attributes, and even zero rows or columns (in any combination). A CEF file without data, but with only row attributes, can be a useful way of storing annotations. Such a file can be joined to a data file to add the annotation to the data file.


## To-do list

	Speed up CEF writer
	Tutorials for common tasks
	Rescale by given column attribute (mean centered)
	Import simple tables
	Aggregate maxcor, mincorr
	SPIN
	Left, right joins
	Select by regex
	Select by < and >
	Parsers and generators for R, Python, MATLAB, Mathematica, Java, 
	Test suite for parsers and generators
	Validator for CEF files
	Fetch dataset from GEO SOFT
	Repository
