package main

import (
	"flag"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

func main() {
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		println("Cellophane v0.1 (C) 2015 Sten Linnarsson <http://linnarssonlab.org/>")
		println()
	}

	// Define command-line flags
	var flHelp = flag.Bool("help", false, "Show this help message and quit")
	flag.Parse()

	if *flHelp {
		println("Usage:")
		flag.PrintDefaults()
		return
	}
}
