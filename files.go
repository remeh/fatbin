// Fatbin
// Rémy Mathieu © 2016
package main

import (
	"fmt"
	"os"
	"strings"
)

type Directory struct {
	Name        string               `json:"name"`        // relative directory path
	Files       map[string]FileInfo  `json:"files"`       // files contained inside the directory
	Directories map[string]Directory `json:"directories"` // sub-directories
	Perm        string               `json:"perm"`        // TODO(remy): permission
}

// File is a file on the FS which is embedded
// in the Fatbin. It'll be extracted before
// launch.
type FileInfo struct {
	Name string `json:"name"` // file name
	Perm string `json:"perm"` // TODO(remy): permission
}

func (d Directory) print() {
	fmt.Printf("%s:\n", d.Name)
	for f := range d.Files {
		fmt.Printf("- %s\n", f)
	}

	for _, subdir := range d.Directories {
		subdir.print()
	}
}

func BuildTree(path string) (Directory, error) {
	return parseDirectory(path)
}

func parseDirectory(path string) (Directory, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	rv := Directory{
		Name:        relative(path),
		Files:       make(map[string]FileInfo, 0),
		Directories: make(map[string]Directory, 0),
	}

	d, err := os.Open(path)
	if err != nil {
		return rv, err
	}

	info, err := d.Stat()
	if err != nil {
		return rv, err
	}

	if !info.IsDir() {
		return rv, fmt.Errorf("The given path is not a directory.")
	}

	// read all the sub directories

	subdirs, err := d.Readdir(0)
	if err != nil {
		return rv, err
	}

	// handle files / dir in this directory
	for _, file := range subdirs {
		if file.IsDir() {
			// embed this sub-directory
			subpath := fmt.Sprintf("%s%s", path, file.Name())
			dir, err := parseDirectory(subpath)
			if err != nil {
				return rv, err
			}

			rv.Directories[relative(dir.Name)] = dir
		} else {
			rv.Files[file.Name()] = FileInfo{
				Name: relative(path) + file.Name(),
			}
		}
	}

	return rv, nil
}

func relative(path string) string {
	if strings.HasPrefix(path, flags.Directory) {
		return strings.Replace(path, flags.Directory, "", -1)
	}
	return path
}
