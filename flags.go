package main

import (
	"flag"
	"fmt"
	"strings"
)

type Flags struct {
	Directory  string // the directory to "fatbinarize"
	Executable string // the file to start on execution of the fatbin
}

var flags Flags

func parseFlags() error {
	var dir, exe string

	flag.StringVar(&dir, "dir", "", "the directory to fatbinerize")
	flag.StringVar(&exe, "exe", "", "the file inside the directory to execute at startup")

	flag.Parse()

	f := Flags{
		Directory:  dir,
		Executable: exe,
	}

	if len(dir) != 0 {
		if !strings.HasPrefix(dir, "/") {
			dir += "/"
			f.Directory += "/"
		}

		if len(exe) == 0 {
			return fmt.Errorf("You must provide an executable when compressing a directory. See flag -exe.")
		}
	}

	flags = f
	return nil
}
