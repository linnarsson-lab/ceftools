# ceftools

Tools for manipulating cell expression binary (CEB) files, and for importing and exporting to text-based CEF files.

**Note: ceftools is not yet ready for use, nor do I know when it will be**

You can download the most recent [Mac OS X binary](cef/cef), named `cef`. To use it, click **View Raw** and save it to your computer. In your Terminal, navigate to the file, make it executable (`chmod uga+x ./cef`) and then run `./cef help`.

## Overview

ceftools is a set of tools designed to manipulate large-scale gene expression data. It aims to be for gene 
expression data what [samtools](http://samtools.github.io) is for sequence data. It was designed to simplify the exchange
and manipulation of very large-scale transcriptomics data, particularly from single-cell RNA-seq. However, it can process
any omics dataset that can be represented as an annotated matrix of numbers.

ceftools is implemented as a command-line utility called `cef`, which operates on an annotated matrix of gene-expression 
values. It supports useful operations such as filtering, sorting, splitting, joining and transposing the input. Multiple 
commands can be chained to perform more complex operations.

ceftools processes binary files in CEB format ('cell expression binary', `.ceb`), which are compact but not human-readable. CEB files can be
converted to and from the text-based CEF format ('cell expression format', `.cef`) suitable for human consumption. CEF files are tab-delimited 
text files that can be easily parsed or imported into e.g. Excel. Not all features of CEB files are guaranteed to be faithfully retained in CEF, so CEB should be considered the canonical reference format.

## Synopsis

Commands that have been implemented so far 

	cef help			- print help for the cef command
	cef info            - overview of file contents
	cef drop 			- drop attribute(s) or header(s)
	cef import			- import from STRT, CEF or CEB
	cef export			- export as CEF
	cef rescale			- rescale rows (rpkm, tpm or log-transformed)
	cef join		  	- join two datasets by given attributes
	cef sort			- sort by row attribute or by specific column
	cef add 			- add attribute or header with constant value 

Commands operate on rows by default. For example `drop` can be used to remove row attributes, but not column attributes. Every command accepts a `--transpose none|in|out|inout` parameter, which causes the CEB to be transposed before and/or after the operation is applied. This can be used to operate on columns. For example, to remove column attribute `Gene` then sort on column attribute `Length`:

```
< infile.ceb cef --transpose in drop Gene | cef --transpose out sort Length > outfile.ceb 
```


## CEF and CEB file formats

Why do we need two file formats? CEF files are easy to read, easily imported into many applications, and easy to parse and generate. However, being text-based they can be slow to read. This is because the file must be read line-by-line, split up into tab-delimited fields, and then each field must be parsed e.g. to determine an expression value. For a large expression matrix, this can translate into a lot of slow parsing. In contrast, CEB files are binary, unreadable for humans, still easy to read and generate from software. Since CEB files store the expression matrix in a predictable binary form, it can simply be loaded into memory and used directly; this is potentially much faster than parsing the equivalent CEF file.


### Detecting the file format

ceftools transparently reads and distinguishes CEB and CEF files without any further specification of the input format. Thus, the input can be in either format, and it will just work.

CEF files are tab-delimited text files in [UTF-8](http://en.wikipedia.org/wiki/UTF-8) encoding. The first four characters are 'CEF\t' (that's a single tab character at the end), equivalent to the hexadecimal 4-byte number 0x09464543. There is no byte order mark (BOM).

CEB files are binary. The first four bytes are [UTF-8](http://en.wikipedia.org/wiki/UTF-8) encoded 'CEB\t' (that's a single tab character at the end), equivalent to the hexadecimal 4-byte number 0x09424543.

The first four bytes of a file therefore unambiguously indicate if it's a CEF or CEB file (assuming it's known to be one of the two).


### CEF file format

Tab-delimited file with newline endings, [UTF-8](http://en.wikipedia.org/wiki/UTF-8) encoding, no [BOM](http://en.wikipedia.org/wiki/Byte_order_mark). Carriage returns before newline characters are silently removed. Each row has the same number of tab-separated fields, equal to `max(7, column count + row attribute count + 1)`. In other words, the entire file is a rectangular tab-delimited matrix, with at least seven columns. 

The first line defines the file structure. It begins 'CEF', followed by header count, column count, row count, column attribute count, row attribute count, and the `Flags` value. 

This is followed by header lines, which are name-value pairs, with the name in the first column and the value in the second. There are no restrictions on either the names or the values, except that they cannot contain tabs, newlines or carriage returns.

Next, the column attributes are given, each in a single row with an offset of 

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



### CEB file format

CEB files are binary files that are intended to be unambiguously specified, easy to parse and easy to generate. 

#### Version

CEB files contain a minor/major version indicator. Major version changes are only backward compatible (newer parsers can read older files). Minor version changes are both forward and backward compatible (older parsers can read newer files). A compliant CEB parser should refuse to read a CEB file with a larger major version, but should ignore the minor version. 

#### Skipped bytes and flags

There is a section in the file, following the main matrix of values, that should simply be skipped. The purpose of this section is to make room for future file format extensions, while maintaining backward compatibility. A future v0.2 file format might store some data in the skipped section, and compliant v0.1 parsers will simply ignore it and still be able to read the file. 

There is also a `Flags` field, currently unused

#### Data types

`string` is a [UTF-8](http://en.wikipedia.org/wiki/UTF-8) encoded string with a leading int32 (signed 32-bit integer) length indicator, and no terminator. For example, the string "Hello, world!" is stored as [12, 'H','e','l','l','o',' ','w','o','r','l',d','!']. The length indicator gives the total number of bytes in the string, excluding the length indicator (it does not count the number of characters in the string, since some UTF-8 characters are stored as multi-byte sequences)

`int32` is a signed 32-bit integer, 4 bytes, [little endian](http://en.wikipedia.org/wiki/Endianness)

`int64` is a signed 64-bit integer, 8 bytes, [little endian](http://en.wikipedia.org/wiki/Endianness)

`byte` is a single byte, 8 bits

`float32` is a 32-bit floating point value, 4 bytes, [IEEE-754](http://en.wikipedia.org/wiki/IEEE_floating_point) encoded

`T[]` is an array of values of type T


#### File structure

	byte[4]					'CEB\t' magic word
	int32					Major version (backward incompatible)
	int32					Minor version (backward compatible)

	int64					Column count (nCols)
	int64					Row count (nRows)
	int64 					Flags

	float32[nRows x nCols]	Main matrix (by rows)

	int64					Skip length (nSkip)
	byte[nSkip]				Skipped bytes (ignored)

	int32					Header entries count
	string					Header #1 name
	string 					Header #1 value
	...						...
	string 					Header #n name
	string 					Header #n value

	int32					Column attribute count (nColAttrs)
	string[nColAttrs]		Column attribute names 
	string[nCols] 			Column attribute #1 values
	...						...
	string[nCols] 			Column attribute #nColAttrs values

	int32					Row attribute count (nRowAttrs)
	string[nRowAttrs]		Row attribute names
	string[nRows] 			Row attribute #1 values (total equal to row count)
	...						...
	string[nRows] 			Row attribute #nRowAttrs values

### To-do list

	Merge headers when joining
	Left, right joins
	Sort by specific column
	Sort by cv/mean offset (https://github.com/glycerine/zettalm)
	Parsers and generators for R, Python, MATLAB, Mathematica, Java, 
	Test suite for parsers and generators
	Validator for CEF/CEB files
	Cloud-based dataset manager
	Internal pipes
	Memory-mapped files

Future commands

	cef select			- select rows that match given criteria
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


