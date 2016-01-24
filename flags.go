// Fatbin
// Rémy Mathieu © 2016
package main

import (
	"flag"
	"fmt"
	"strings"
)

type Flags struct {
	Directory  string // the directory to "fatbinarize"
	Executable string // the file to start on execution of the fatbin
	Output     string // the archive file to create.
}

var flags Flags

func parseFlags() error {
	var dir, exe, out string

	flag.StringVar(&dir, "f.dir", "", "the directory to fatbinerize")
	flag.StringVar(&exe, "f.exe", "", "the file inside the fatbin archive to execute at startup")
	flag.StringVar(&out, "f.out", "archive.fbin", "the archive file to create.")

	flag.Parse()

	f := Flags{
		Directory:  dir,
		Executable: exe,
		Output:     out,
	}

	if len(dir) != 0 {
		if !strings.HasPrefix(dir, "/") {
			dir += "/"
			f.Directory += "/"
		}

		if len(exe) == 0 {
			return fmt.Errorf("You must provide an executable when compressing a directory. See flag -f.exe.")
		}

		if len(out) == 0 {
			return fmt.Errorf("The output file can't be empty in creation mode.")
		}
	}

	flags = f
	return nil
}
