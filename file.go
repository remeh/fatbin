package main

// File is a file on the FS which is embedded
// in the Fatbin. It'll be extracted before
// launch.
type FileInfo struct {
	Tree []string // directory path
	Name string   // file name
	Perm string   // TODO(remy): permission
}
