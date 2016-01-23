package main

import (
	"flag"
	"fmt"
)

func main() {
	// read the flags

	err := parseFlags()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
	}

	tree, err := BuildTree(flags.Directory)
	if err != nil {
		fmt.Println(err)
		return
	}

	tree.print()
}
