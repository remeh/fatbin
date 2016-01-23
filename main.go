package main

import (
	"flag"
	"fmt"
)

func main() {
	// read the flags

	flags, err := parseFlags()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
	}

}
