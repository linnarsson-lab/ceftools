# ceftools

Tools for manipulating cell expression format (CEF) files.

**Note: ceftools is not yet ready for use, nor do I know when it will be**

You can download pre-release alpha versions on the [Releases page](https://github.com/linnarsson-lab/ceftools/releases).

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

Commands that have been implemented so far 

	cef help			- print help for the cef command
	cef info            - overview of file contents
	cef drop 			- drop attribute(s) or header(s)
	cef import			- import from STRT
	cef rescale			- rescale rows (rpkm, tpm or log-transformed)
	cef join		  	- join two datasets by given attributes
	cef sort			- sort by row attribute or by specific column
	cef add 			- add attribute or header with constant value 
	cef transpose 		- transpose the file
	cef select			- select rows that match given criteria

Commands operate on rows by default. For example `drop` can be used to remove row attributes, but not column attributes. Use the global `--bycol` flag to operate instead on columns. For example, to remove column attribute `Gene` then sort on column attribute `Length`:

```
< infile.cef cef --bycol drop Gene | cef --bycol sort Length > outfile.cef 
```



### CEF file format

CEF files are tab-delimited text files in [UTF-8](http://en.wikipedia.org/wiki/UTF-8) encoding, no [BOM](http://en.wikipedia.org/wiki/Byte_order_mark). The first four characters are 'CEF\t' (that's a single tab character at the end), equivalent to the hexadecimal 4-byte number 0x09464543. CEF files are guaranteed to always begin with these four bytes, which can be used to identify the file format in the absence of a file name extension.

Each row has the same number of tab-separated fields, equal to `max(7, column count + row attribute count + 1)`. In other words, the entire file is a rectangular tab-delimited matrix, with at least seven columns. CEF file *readers* should accept CEF files that have less than the required number of fields in any row, and the missing fields should be interpreted as empty strings (but empty strings should not be interpreted as zeros; thus zeros must always be explicitly represented as '0'). CEF file *writers* should always generate a rectangular tab-delimited matrix.

Carriage returns before newline characters are silently removed. Fields may be quoted using double quotes; these are silently removed when fields are read. Tabs and newlines are allowed inside a quoted field.

The first line defines the file structure. It begins 'CEF', followed by header count, row attribute count, column attribute count, row count, column count, and the `Flags` value. 

This is followed by header lines, which are name-value pairs, with the name in the first column and the value in the second. There are no restrictions on either the names or the values. The order of headers is not necessarily preserved when CEF files are read and written. There can be multiple headers with the same name.

Next, the column attributes are given, each in a single row with an offset of `(row attribute count)`. Finally, the rows are given, starting with row attributes, and followed by the values of the main matrix. Values are represented in text as decimal floating-point numbers with no exponent (e.g. `-142.03939`) and must fit in a 32-bit IEEE-754 floating point number. The order of row and column attributes need to be preserved.

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


### To-do list

	Left, right joins
	Sort by specific column
	Sort by cv/mean offset (https://github.com/glycerine/zettalm)
	Parsers and generators for R, Python, MATLAB, Mathematica, Java, 
	Test suite for parsers and generators
	Validator for CEF files
	Cloud-based dataset manager

Future commands

	cef aggregate		- calculate aggregate statistics for every row
	cef groupby			- group rows that share a row attribute, and aggregate values
	cef validate		- verify that the input file conforms to the CEB or CEF standard
	cef view			- interactively navigate the matrix
	cef browse			- web interface to cef ? (https://github.com/jteeuwen/go-bindata)

Future repo tools

	cef repo --create <name>			- Create a repository
	cef repo --put "repo/dataset"		- Upload a dataset
	cef repo --delete "repo/dataset"	- Remove a dataset
	cef repo --get "repo/dataset"		- Get a dataset from given repo
	cef repo --list						- List repositories
	cef repo --content <repo>			- List the datasets in a given repo
	cef repo --about <repo>				- Show information about given repo (owner, description, ...)

	cef repo --put "slinnarsson/cortex" --desc "Data from Zeisel et al. Science 2015"


