package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type Fatbin struct {
	Version    string    `json:"version"`
	Executable string    `json:"executable"`
	Directory  Directory `json:"dir"`
}

func RunFatbin(filename string) error {
	// create a tmp directory where everything will go
	dir, err := ioutil.TempDir("", "fatbin")
	if err != nil {
		return err
	}

	if fatbin, err := readFatbin(filename, dir); err != nil {
		return err
	} else {
		defer func() {
			if len(dir) > 0 && dir != "/" {
				if err := os.RemoveAll(dir); err != nil {
					fmt.Println("Can't remove the temporary dir:", err.Error())
				}
			}
		}()

		if err := os.Chmod(dir+"/"+fatbin.Executable, 0755); err != nil {
			return err
		}

		// execute the binary
		cmd := exec.Command(dir + "/" + fatbin.Executable)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

func readFatbin(filename, dstDir string) (Fatbin, error) {
	rv := Fatbin{}

	src, err := os.Open(filename)
	if err != nil {
		return rv, err
	}

	return Parse(src, dstDir)
}

// TODO(remy): finish this method
// TODO(remy): should we close the file here or outside ?
func BuildFatbin(directory Directory, executable string) (Fatbin, *os.File, error) {
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

	f.Write(TOKEN_HEADER_START)
	f.Write(headers)
	f.Write([]byte("\n"))
	f.Write(TOKEN_HEADER_END)

	f.Write(TOKEN_DATA_START)

	// write each files
	if err := write(fatbin, f); err != nil {
		return fatbin, f, err
	}

	f.Write(TOKEN_DATA_END)

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
	dst.Write(TOKEN_FILE_START)
	dst.Write(TOKEN_FILE_HEADER_START)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	dst.Write(data)
	dst.Write([]byte("\n"))
	dst.Write(TOKEN_FILE_HEADER_END)
	dst.Write(TOKEN_FILE_DATA_START)
	f, err := os.Open(flags.Directory + info.Name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, f); err != nil {
		return err
	}
	dst.Write([]byte("\n"))
	dst.Write(TOKEN_FILE_DATA_END)
	dst.Write(TOKEN_FILE_END)
	return nil
}

func createFatbinDirectories(fatbin Fatbin, dstDir string) error {
	return createDirectories(fatbin.Directory, dstDir)
}

func createDirectories(directory Directory, dstDir string) error {
	for _, d := range directory.Directories {
		println(d.Name)
		if err := os.MkdirAll(dstDir+d.Name, 0755); err != nil {
			return err
		}

		// sub dir
		createDirectories(d, dstDir+"/")
	}
	return nil
}

func extractFile(dstDir string, fileInfo FileInfo, data []byte) error {
	// avoid a disaster
	if len(dstDir) == 0 || dstDir == "/" {
		return fmt.Errorf("Error in the dst dir: %s\n", dstDir)
	}
	file, err := os.Create(dstDir + fileInfo.Name)
	if err != nil {
		return err
	}

	defer file.Close()

	fmt.Printf("Extracted %s\n", dstDir+fileInfo.Name)

	_, err = file.Write(data)
	return err
}

/*
	TODO(remy): implements Reader/Writer
	func (f Fatbin) Read(p []byte) (int, error) {
	}

	func (f Fatbin) Write(p []byte) (int, error) {
	}
*/
