// Todo: Use goroutines for copy contents of files 

package main

import (
	"path/filepath"
	"strings"
	"io"
	"os"
	"flag"
	"fmt"
)

var backupDst string
var backupSrc string
var backupName string

// Called by filepath.Walk() whenever it comes accross a file, 
// determine its new path at the destination and copy over it over
func handle(path string, f os.FileInfo, err error) error {
	var dst = strings.Replace(path, backupSrc, "", -1)
	dst = filepath.Join(backupDst, dst)
	
	fmt.Printf("Copying: %s -> %s\n", path, dst)
	err = copyFile(path, dst)
	if err != nil {
		fmt.Printf("CopyFile failed %q\n", err)
	} else {
		fmt.Printf("CopyFile succeeded\n")
	}
	
	return nil
}

// First check the source file is non regular (directory, symlink etc)
// all these checks are performed using os.Stat which returns fileinfo.
// Secondly check the destination file is non regular AND that it isnt the
// same in as the source file
// Else copy the contents of the source file to the new destination file
// Source: https://stackoverflow.com/a/21067803
func copyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}

	err = copyFileContents(src, dst)
	return
}

// Open up the source file and create and copy the contents over to the new
// destination file
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	path, _ := filepath.Split(dst)
	if  _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func init() {
	// Setup arguments/flags
	flag.StringVar(&backupSrc, "source", "", "Path to backup")
	flag.StringVar(&backupSrc, "s", "",  "Path to backup (shorthand)")
	flag.StringVar(&backupDst, "destination", "", "Path to save backup to")
	flag.StringVar(&backupDst, "d", "", "Path to save backup to (shorthand)")
	flag.StringVar(&backupName, "name", "", "Name of folder to save backup to (optional)")
	flag.StringVar(&backupName, "n", "", "Name of folder to save backup to (optional) (shorthand)")
}

func main() {
	flag.Parse()

	// Check arguments are set
	if backupSrc == "" || backupDst == "" {
		fmt.Fprint(os.Stderr, "Please specify a source and a destination path for the backup\n")
		flag.Usage()
		os.Exit(1)
	}
	
	err := filepath.Walk(backupSrc, handle)
	fmt.Printf("filepath.Walk() return %v\n", err)
}
