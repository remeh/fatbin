package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// read the flags

	err := parseFlags()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		return
	}

	// compress mode
	if len(flags.Directory) > 0 && len(flags.Executable) > 0 {
		build()
		return
	}

	run()
}

func run() {
	var run string
	if len(os.Args) == 1 {
		run = "archive.fbin"
	} else if len(os.Args) > 1 {
		run = os.Args[1]
	}

	if err := RunFatbin(run); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func build() {
	tree, err := BuildTree(flags.Directory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err := BuildFatbin(tree, flags.Executable); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// TODO(finish):
}
