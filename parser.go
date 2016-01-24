// Fatbin
// Rémy Mathieu © 2016
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

var (
	TOKEN_HEADER_START      = []byte("<fatbin-header>\n")
	TOKEN_HEADER_END        = []byte("</fatbin-header>\n")
	TOKEN_DATA_START        = []byte("<fatbin-data>\n")
	TOKEN_FILE_START        = []byte("<fatbin-file>\n")
	TOKEN_FILE_END          = []byte("</fatbin-file>\n")
	TOKEN_FILE_HEADER_START = []byte("<fatbin-file-header>\n")
	TOKEN_FILE_HEADER_END   = []byte("</fatbin-file-header>\n")
	TOKEN_FILE_DATA_START   = []byte("<fatbin-file-data>\n")
	TOKEN_FILE_DATA_END     = []byte("</fatbin-file-data>\n")
	TOKEN_DATA_END          = []byte("</fatbin-data>\n")
)

func Parse(src *os.File, dstDir string) (Fatbin, error) {
	var err error
	var rv Fatbin

	reader := bufio.NewReader(src)

	// the very first line must be the TOKEN_HEADER_START

	if err := nextLineExpected(reader, TOKEN_HEADER_START); err != nil {
		return rv, err
	}

	// read until the TOKEN_HEADER_END token
	// it is the header file

	header, err := readUntil(reader, TOKEN_HEADER_END)
	if err != nil {
		return rv, err
	}

	// unmarshal the header
	if err := json.Unmarshal(header, &rv); err != nil {
		return rv, err
	}

	// next is the TOKEN_DATA_START

	if err := nextLineExpected(reader, TOKEN_DATA_START); err != nil {
		return rv, err
	}

	// we're ready to parse the whole fatbin but before, create each needed
	// directories
	if err := createFatbinDirectories(rv, dstDir); err != nil {
		return rv, err
	}

	// now, we will read files until we met a TOKEN_DATA_END
	if err := readFiles(reader, dstDir); err != nil {
		return rv, err
	}

	return rv, nil
}

func readFiles(reader *bufio.Reader, dstDir string) error {

	// NOTE(remy): I directly re-use the TOKEN for the parser
	// state, which is not every time a good idea, but I definitely
	// know it'll work.
	parserState := string(TOKEN_DATA_START)

	var fileInfo FileInfo

	for {
		switch parserState {
		// we've just started the parsing or we've finished a file
		// we must met either an FILE_START or a DATA_END
		case string(TOKEN_DATA_START):
			line, err := nextLine(reader)
			if err != nil {
				return unexpectedParsingError(err)
			}

			if equals(line, TOKEN_DATA_END) {
				// no files, end of the parsing
				break
			}

			if !equals(line, TOKEN_FILE_START) {
				return unexpectedToken(TOKEN_FILE_START, line)
			}

			parserState = string(TOKEN_FILE_START)

		// we're entering a file, we must met a
		// FILE_HEADER_START token
		case string(TOKEN_FILE_START):
			line, err := nextLine(reader)
			if err != nil {
				return unexpectedParsingError(err)
			}

			if !equals(line, TOKEN_FILE_HEADER_START) {
				return unexpectedToken(TOKEN_FILE_HEADER_START, line)
			}

			parserState = string(TOKEN_FILE_HEADER_START)

		// we will read some file's header.
		case string(TOKEN_FILE_HEADER_START):
			headers, err := readUntil(reader, TOKEN_FILE_HEADER_END)
			if err != nil {
				return unexpectedParsingError(err)
			}

			if err := json.Unmarshal(headers, &fileInfo); err != nil {
				return unexpectedParsingError(err)
			}

			parserState = string(TOKEN_FILE_HEADER_END)

		// header of a file read, we must now find a TOKEN_FILE_DATA_START
		case string(TOKEN_FILE_HEADER_END):
			line, err := nextLine(reader)
			if err != nil {
				return unexpectedParsingError(err)
			}

			if !equals(line, TOKEN_FILE_DATA_START) {
				return unexpectedToken(TOKEN_FILE_DATA_START, line)
			}

			parserState = string(TOKEN_FILE_DATA_START)

		// we're now reading file binary data until we met TOKEN_FILE_DATA_END
		// NOTE(remy): it could be very RAM costly because the whole file is
		// written in RAM by readUntil. A better thing would be to streamly read
		// and write the data until we met the end token.
		case string(TOKEN_FILE_DATA_START):
			data, err := readUntil(reader, TOKEN_FILE_DATA_END)
			if err != nil {
				unexpectedParsingError(err)
			}

			// the last char MUST be a \n (because we wrote it during serialization)
			// so we remove it here
			if data[len(data)-1] == '\n' {
				data = data[:len(data)-1]
			}

			if err := extractFile(dstDir, fileInfo, data); err != nil {
				return err
			}

			parserState = string(TOKEN_FILE_DATA_END)

		// end of a file data, we must met a FILE_END
		case string(TOKEN_FILE_DATA_END):
			if err := nextLineExpected(reader, TOKEN_FILE_END); err != nil {
				return err
			}

			parserState = string(TOKEN_FILE_END)

		// this file is finished, we will either met a start of
		// a new file or the end of the parsing
		case string(TOKEN_FILE_END):
			line, err := nextLine(reader)
			if err != nil {
				return unexpectedParsingError(err)
			}

			if equals(line, TOKEN_FILE_START) {
				parserState = string(TOKEN_FILE_START)
				continue
			}

			if equals(line, TOKEN_DATA_END) {
				return nil
			}

			return unexpectedToken([]byte(fmt.Sprintf("%s or %s", TOKEN_FILE_START, TOKEN_DATA_END)), line)

		default:
			return fmt.Errorf("Unexpected parser state. Can't parse the file.")
		}
	}
}

func unexpectedParsingError(err error) error {
	return fmt.Errorf("Unexpected error while parsing: %s", err.Error())
}

func nextLineExpected(reader *bufio.Reader, expecting []byte) error {
	var line []byte
	var err error

	if line, err = nextLine(reader); err != nil {
		return err
	}

	if !equals(line, expecting) {
		return unexpectedToken(expecting, line)
	}

	return nil
}

func unexpectedToken(expecting []byte, line []byte) error {
	limit := 32
	complete := ""

	if len(line) >= limit {
		complete = "..."
	}

	if len(line) <= limit {
		limit = len(line)
	}

	return fmt.Errorf("Unexpected token.\nExpected: %s\nHad: %s%s\n", expecting, line[:limit], complete)
}

func readUntil(reader *bufio.Reader, endToken []byte) ([]byte, error) {
	data := bytes.NewBuffer(nil)

	for {
		b, err := nextLine(reader)
		if err != nil {
			return nil, fmt.Errorf("Unexpected error in readUntil: %s", err.Error())
		}

		if equals(b, endToken) {
			break
		}

		data.Write(b)
	}

	return data.Bytes(), nil
}

func nextLine(reader *bufio.Reader) ([]byte, error) {
	return reader.ReadBytes('\n')
}

func equals(a, b []byte) bool {
	return bytes.Equal(a, b)
}
