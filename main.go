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
		return
	}

	if len(flags.Directory) > 0 && len(flags.Executable) > 0 {
		// compress mode
		tree, err := BuildTree(flags.Directory)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, file, err := CreateFatbin(tree, flags.Executable)
		if file != nil {
			defer file.Close()
		}

		if err != nil {
			fmt.Println(err)
		}
	} else {
		// run mode
		println("run mode")
	}
}
