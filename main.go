// Fatbin
// Rémy Mathieu © 2016
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// redefine the flags function
	flag.Usage = func() {
		// print defaults of fatbin only if it's named fatbin
		if filepath.Base(os.Args[0]) == "fatbin" {
			flag.PrintDefaults()
			return
		}
	}

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
	if err := RunFatbin(os.Args[1:]...); err != nil {
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
}
