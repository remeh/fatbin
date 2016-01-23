package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Fatbin struct {
	Version    string    `json:"version"`
	Executable string    `json:"executable"`
	Directory  Directory `json:"dir"`
}

func CreateFatbin(directory Directory, executable string) (Fatbin, *os.File, error) {
	rv := Fatbin{}

	// look for the executable in the tree
	if len(directory.Files[executable].Name) == 0 {
		return rv, nil, fmt.Errorf("Can't find the executable in the root directory.")
	}

	// open the target file
	f, err := ioutil.TempFile("", "fatbin")
	if err != nil {
		return rv, nil, err
	}

	fatbin := Fatbin{
		Version:    "1",
		Executable: executable,
		Directory:  directory,
	}

	// we must first write the header files
	headers, err := json.Marshal(fatbin)
	if err != nil {
		return rv, nil, err
	}

	f.Write(headers)
	f.Write([]byte("\n"))

	f.Write([]byte("<fatbin-data>\n"))

	// write each files
	if err := write(fatbin, f); err != nil {
		return fatbin, f, err
	}

	f.Write([]byte("</fatbin-data>\n"))

	return rv, f, nil
}

func write(fatbin Fatbin, dst *os.File) error {
	return writeDirectory(fatbin.Directory, dst)
}

func writeDirectory(dir Directory, dst *os.File) error {
	for _, fi := range dir.Files {
		if err := writeFile(fi, dst); err != nil {
			return err
		}
	}

	// write sub directories
	for _, subdir := range dir.Directories {
		if err := writeDirectory(subdir, dst); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(info FileInfo, dst *os.File) error {
	dst.Write([]byte("<fatbin-file>\n"))
	dst.Write([]byte("<fatbin-file-header>\n"))
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	dst.Write(data)
	dst.Write([]byte("\n"))
	dst.Write([]byte("</fatbin-file-header>\n"))
	dst.Write([]byte("<fatbin-file-data>\n"))
	f, err := os.Open(flags.Directory + info.Name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, f); err != nil {
		return err
	}
	dst.Write([]byte("\n"))
	dst.Write([]byte("</fatbin-file-data>\n"))
	dst.Write([]byte("</fatbin-file>\n"))
	return nil
}

/*
	TODO(remy): implements Reader/Writer
	func (f Fatbin) Read(p []byte) (int, error) {
	}

	func (f Fatbin) Write(p []byte) (int, error) {
	}
*/
