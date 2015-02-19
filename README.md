# Cellophane
Tools for manipulating cell expression format (CEF and CEB) files

## Overview

Cellophane is a set of tools designed to manipulate large-scale gene expression data. It aims to be for gene 
expression data what [samtools](http://samtools.github.io) is for sequence data. It was designed to simplify the exchange
and manipulation of very large-scale transcriptomics data, particularly from single-cell RNA-seq. However, it can process
any omics dataset that can be represented as an annotated matrix of numbers.

Cellophane is implemented as a command-line utility called `cef`, which operates on an annotated matrix of gene-expression 
values. It supports useful operations such as filtering, sorting, splitting, joining and transposing the input. Multiple 
commands can be chained to perform more complex operations.

Cellophane works with binary files in CEB format ('cell expression binary', `.ceb`), which are compact but not human-readable. CEB files can be
converted to and from the text-based CEF format ('cell expression format', `.cef`) suitable for human consumption. CEF files are tab-delimited 
text files that can be easily parsed or imported into e.g. Excel. 

## Synposis

```
cef info            - overview of file contents
cef transpose	  	- transpose rows and columns
cef T               - transpose rows and columns (synonym for 'transpose')
cef join		  	- join two datasets by given identifier
cef remove 			- remove attribute
cef annotate        - insert attribute from file
cef filter			- filter rows by given criteria
cef normalize		- normalize rows
cef sort			- sort by attribute or column value, or by correlation
cef aggregate		- calculate aggregate statistics for every row
cef view			- print parts of the matrix
cef import			- import a .cef file and output a .ceb file
cef export			- read a .ceb file and export a .cef
cef validate		- validate a .cef or .ceb file 

cef A : B           - pipe command A into command B (internal pipe)
```
