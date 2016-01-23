package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	Directory  string // the directory to "fatbinarize"
	Executable string // the file to start on execution of the fatbin
}

var flags Flags

func parseFlags() error {
	var dir, exec string

	flag.StringVar(&dir, "dir", "", "the directory to fatbinerize")
	flag.StringVar(&exec, "exec", "", "the file inside the directory to execute at startup")

	flag.Parse()

	f := Flags{
		Directory:  dir,
		Executable: exec,
	}

	if len(dir) == 0 {
		return fmt.Errorf("You must provide a directory. See flag -dir")
	}

	if len(exec) == 0 {
		return fmt.Errorf("You must provide an executable to run. See flag -exec")
	}

	flags = f
	return nil
}
