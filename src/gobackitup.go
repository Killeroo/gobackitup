// Todo: Use goroutines for copy contents of files 
// Todo: Move backup details to a struct?
// Zipping:
// http://blog.ralch.com/tutorial/golang-working-with-zip/
// https://gist.github.com/svett/424e6784facc0ba907ae
// https://golang.org/src/archive/zip/example_test.go

package main

import (
	"path/filepath"
	"archive/zip"
	"strings"
	"io"
	"os"
	"flag"
	"fmt"
)

type backupInfo struct {
	dst string
	src string
	name string
	zip bool
}

var data backupInfo

// Called by filepath.Walk() whenever it comes accross a file, 
// determine its new path at the destination and copy over it over
func handle(path string, f os.FileInfo, err error) error {
	var dst = strings.Replace(path, data.src, "", -1)
	dst = filepath.Join(data.dst, dst)
	
	fmt.Printf("Copying: %s -> %s\n", path, dst)
	err = copyFile(path, dst)
	if err != nil {
		fmt.Printf("CopyFile failed %q\n", err)
	} else {
		fmt.Printf("CopyFile succeeded\n")
	}
	
	return nil
}



// Src: https://gist.github.com/svett/424e6784facc0ba907ae
func zipfolder (src, dst string) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	var base string // RENAME TO ROOT
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.IsDir() { // !REMOVE? CHECK FOR DIR EARLIER?
		base = filepath.Base(src)
	}

	// SEPERATE INTO callback function and zip function (zipFile) just like we did before
	filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
		if err != nil { // !!REMOVE? UNESSARY
			return err
		}
		
		// Create zip header of file for archive
		header, err := zip.FileInfoHeader(f)
		if err != nil {
			return err
		}

		if base != "" { // REMOVE? Look at top messag
			header.Name = filepath.Join(base, strings.TrimPrefix(path, src))
		}

		fmt.Printf("Zipping: %s -> %s\n", path, header.Name)
		
		if f.IsDir() {
			header.Name += "/" // CHANGE THIS TO BE OPERATING SYSTEM INDEPENDANT 
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}			
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// First check the source file is non regular (directory, symlink etc)
// all these checks are performed using os.Stat which returns fileinfo.
// If the source is a folder then a corresponding folder is created in
// the destination.
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
		if sfi.Mode().IsDir() {
			os.MkdirAll(dst, os.ModePerm)
			return
		} else {
			return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())		
		}
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
	flag.StringVar(&data.src, "source", "", "Path to backup")
	flag.StringVar(&data.src, "s", "",  "(shorthand) Path to backup")
	flag.StringVar(&data.dst, "destination", "", "Path to save backup to")
	flag.StringVar(&data.dst, "d", "", "(shorthand) Path to save backup to")
	flag.StringVar(&data.name, "name", "", "(optional) Name of folder to save backup to ")
	flag.StringVar(&data.name, "n", "", "(optional) (shorthand) Name of folder to save backup to")
	flag.BoolVar(&data.zip, "zip", false, "(optional) Compress the backup")
	flag.BoolVar(&data.zip, "z", false, "(optional) (shorthand) Compress the backup")
}

func main() {
	flag.Parse()

	// Check arguments are set
	if data.src == "" || data.dst == "" {
		fmt.Fprint(os.Stderr, "Please specify a source and a destination path for the backup\n")
		flag.Usage()
		os.Exit(1)
	}

	if data.name != "" {
		data.dst = filepath.Join(data.dst, data.name)
		if err := os.MkdirAll(data.dst, os.ModePerm); err != nil {
			fmt.Fprint(os.Stderr, "Could not create folder %s", data.name)
			os.Exit(1)
		}
	}

	err := zipfolder(data.src, data.dst)
	//err := filepath.Walk(data.src, handle)
	if err != nil {
		fmt.Fprint(os.Stderr, "Oh no! gobackitup returned an error: %v\n", err)
		os.Exit(3)
	}
}
