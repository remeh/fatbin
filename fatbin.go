// Fatbin
// Rémy Mathieu © 2016
package main

import (
	"bufio"
	"compress/gzip"
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

var dataSeparator = fmt.Sprintf("%s%s%s\n", "<", "fatbin", ">")

// RunFatbin starts the given fatbin archive.
func RunFatbin(args ...string) error {
	var err error

	// create a tmp directory where everything will go
	dir, err := ioutil.TempDir("", "fatbin")
	if err != nil {
		return err
	}

	// we will write the whole data after <fatbin> into a tmp file
	var archive string
	if archive, err = extractData(os.Args[0]); err != nil {
		return err
	}

	if fatbin, err := readFatbin(archive, dir); err != nil {
		return err
	} else {
		defer func() {
			if len(dir) > 0 && dir != "/" {
				if err := os.Remove(archive); err != nil {
					fmt.Println("Can't remove the temporary archive:", err.Error())
				}
				if err := os.RemoveAll(dir); err != nil {
					fmt.Println("Can't remove the temporary dir:", err.Error())
				}
			}
		}()

		if err := os.Chmod(dir+"/"+fatbin.Executable, 0755); err != nil {
			return err
		}

		// execute the binary
		cmd := exec.Command(dir+"/"+fatbin.Executable, args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// extractData extracts the binary archive from the binary executable.
func extractData(filename string) (string, error) {
	fbin, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	defer fbin.Close()

	tmpArch, err := ioutil.TempFile("", "fatarch")
	if err != nil {
		return "", err
	}

	defer tmpArch.Close()

	reader := bufio.NewReader(fbin)
	for {
		d, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		// found the data part, write it to a temporary archive.
		if equals(d, []byte(dataSeparator)) {
			if _, err := io.Copy(tmpArch, reader); err != nil {
				return "", err
			}

			return tmpArch.Name(), nil
		}
	}

	return "", fmt.Errorf("Can't find the data in the file.")
}

func readFatbin(filename, dstDir string) (Fatbin, error) {
	rv := Fatbin{}

	src, err := os.Open(filename)
	if err != nil {
		return rv, err
	}

	return Parse(src, dstDir)
}

func BuildFatbin(directory Directory, executable string) (Fatbin, error) {
	rv := Fatbin{}

	// look for the executable in the tree
	if len(directory.Files[executable].Name) == 0 {
		return rv, fmt.Errorf("Can't find the executable in the root directory.")
	}

	// first we copy the fatbin binary to a new file

	dst, err := os.Create(flags.Output)
	if err != nil {
		return rv, fmt.Errorf("Can't write the final archive: %s", err.Error())
	}

	fbin, err := os.Open(os.Args[0])
	if err != nil {
		return rv, err
	}

	if _, err := io.Copy(dst, fbin); err != nil {
		return rv, err
	}

	// write the separator
	// NOTE(remy): I use Sprintf because otherwise <fatbin> appears in the sources
	// and is parsed at runtime.
	if _, err := dst.Write([]byte("\n" + dataSeparator)); err != nil {
		return rv, err
	}

	// open the tmp file for the archive
	f, err := ioutil.TempFile("", "fatbin")
	if err != nil {
		return rv, err
	}

	gz := gzip.NewWriter(f)

	defer f.Close()

	fatbin := Fatbin{
		Version:    "1",
		Executable: executable,
		Directory:  directory,
	}

	// we must first write the header files
	headers, err := json.Marshal(fatbin)
	if err != nil {
		return rv, err
	}

	gz.Write(TOKEN_HEADER_START)
	gz.Write(headers)
	gz.Write([]byte("\n"))
	gz.Write(TOKEN_HEADER_END)

	gz.Write(TOKEN_DATA_START)

	// write each files
	if err := write(fatbin, gz); err != nil {
		return fatbin, err
	}

	gz.Write(TOKEN_DATA_END)
	gz.Close()

	// rewind the temp file
	f.Seek(0, 0)

	defer func() {
		dst.Sync()
		dst.Close()
	}()

	// NOTE(remy): I don't use os.Rename because it has many
	// limitations (e.g. no rename between partitions).
	if _, err := io.Copy(dst, f); err != nil {
		return rv, err
	}

	fmt.Printf("Archive created into : %s\n", flags.Output)

	return rv, nil
}

func write(fatbin Fatbin, dst io.Writer) error {
	return writeDirectory(fatbin.Directory, dst)
}

func writeDirectory(dir Directory, dst io.Writer) error {
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

func writeFile(info FileInfo, dst io.Writer) error {
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

	//fmt.Printf("Extracted %s\n", dstDir+fileInfo.Name)

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
