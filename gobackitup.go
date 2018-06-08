// TODO: Use goroutines for copy contents of files (No point limited by usb read write anyway) 
// TODO: Autocreate backup directory
// TODO: Use of pass data object instead of seperate src, dst variables

package main

import (
	"path/filepath"
	"archive/zip"
	"strings"
	"io"
	"os"
	"flag"
	"fmt"
	"strconv"
	"github.com/daviddengcn/go-colortext"
)

type backupInfo struct {
	dst string
	src string
	name string
	zip bool
}

var data backupInfo

// Results
var errors, files int
var bytesCopied uint64

// Coverts bytes to string of appropriate size and denomination (rounds up) 
// Eg '1500' bytes -> '1KB'
func fileSize(size uint64) (result string) {
	if size < 1024 {
		return strconv.Itoa(int(size)) + "B"
	} else if conv := size / 1024; conv < 1024 {
		return strconv.Itoa(int(conv)) + "KB"
	} else if  conv := (size/1024)/1024; conv < 1024 {
		return strconv.Itoa(int(conv)) + "MB"
	} else {
		return strconv.Itoa(int(((size/1024)/1024)/1024)) + "GB"
	} 
}

// Displays and counts error messages
func ErrorMsg(err error) {
	ct.Foreground(ct.Red, false)
	fmt.Printf("Error: %q\n", err)
	ct.ResetColor()

	errors++
}

// Writes some output to console about a given file when copying or zipping
func DeclareFile(f os.FileInfo, path string) {
	if data.zip {
		fmt.Printf("Zipping:")
	} else {
		fmt.Printf("Copying:")
	}
	fmt.Printf(" %s ", path)

	size := fileSize(uint64(f.Size()))
	if strings.ContainsAny(size, "G") {
		ct.Foreground(ct.Magenta, false)
	} else if strings.ContainsAny(size, "M") {
		ct.Foreground(ct.Red, false)
	} else if strings.ContainsAny(size, "K") {
		ct.Foreground(ct.Yellow, false)
	} else if strings.ContainsAny(size, "B") {
		ct.Foreground(ct.Green, false)
	}
			
	fmt.Printf("[%s]", size)
	ct.ResetColor()
	fmt.Printf(" -> ")

	bytesCopied += uint64(f.Size())
	files++
}

// Zips up folder to destination path, new zip file is names after
// base file of the source directory.
// First we determine the name of the zip file, we account for a few 
// edge cases here (mainly on windows). Then we create a zip file 
// at the destination path named after the base directory of the source.
// Then we create an archive object for our new zip file.
// Next walk through the src directory and for each file we create a
// zip info header. If its a directory add a path separator, otherwise
// if its a file add a deflate method then create the header in the archive.
// Lastly if its a file we copy over the file to the archive using an
// writer from the archive.
func zipFolder(src string) (err error) {
	var zipname = ""
	if data.name != "" {
		zipname = data.name	
	} else {
		zipname = filepath.Base(data.src)
		if zipname == "\\" {

			zipname = strings.Trim(filepath.VolumeName(data.src), ":") + "_backup" 
		}
	}
	
	filename := filepath.Join(data.dst, zipname + ".zip") // Windows bug here
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	defer archive.Close()

	base := filepath.Base(src)

	filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
		header, err := zip.FileInfoHeader(f)
		if err != nil {
			ErrorMsg(err)
			return err
		}
		header.Name = filepath.Join(base, strings.TrimPrefix(path, src))
		
		// Only print out info on files
		if !f.Mode().IsDir() {
			DeclareFile(f, path)
		}
		
		if f.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			ErrorMsg(err)
			return err
		}

		if f.IsDir() {
			return nil
		} 
		
		srcFile, err := os.Open(path)
		if err != nil {
			ErrorMsg(err)
			return err
		} 
		defer srcFile.Close()
		
		_, err = io.Copy(writer, srcFile)
		if err != nil {
			ErrorMsg(err)
		} else {
			if !f.Mode().IsDir() {
				ct.Foreground(ct.Green, false)
				fmt.Printf("File archieved\n")
				ct.ResetColor()
			}
		}
		return err
	})

	return err
}

// Copies a folder to a new location, base folder of the
// source is used at the destination path
func copyFolder(src, dst string) (err error) {
	data.dst = filepath.Join(dst, filepath.Base(src))
	
	err = filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
		var dst = strings.Replace(path, data.src, "", -1)
		dst = filepath.Join(data.dst, dst)

		if !f.Mode().IsDir() {
			DeclareFile(f, path)
		}
		err = copyFile(path, dst)
		if err != nil {
			ErrorMsg(err)
		} else {
			if !f.Mode().IsDir() {
				ct.Foreground(ct.Green, false)
				fmt.Printf("File copied\n")
				ct.ResetColor()
			}
		}
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

func usage() {
	fmt.Println("gobackitup [options] --source <PATH> --destination <PATH>")
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(0)
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
	flag.BoolVar(&data.zip, "z", false, "\n(optional) (shorthand) Compress the backup")
}

func main() {
	flag.Usage = usage
	flag.Parse()
	
	// Check arguments are set
	if data.src == "" || data.dst == "" {
		fmt.Fprint(os.Stderr, "Please specify a source and a destination path for the backup\n")
		fmt.Fprint(os.Stderr, "Use gobackitup --help for options")
		flag.Usage()
		os.Exit(1)
	}

	// Check paths exist
	if _, err := os.Stat(data.src); os.IsNotExist(err) {
		fmt.Fprint(os.Stderr, "Source or destination path not found. Please check they exist and try again")
		fmt.Fprint(os.Stderr, "Use gobackitup --help for options")
		os.Exit(1)		
	}

	// TODO: drive folder and backup folder creation here
	if data.name != "" && !data.zip {
		data.dst = filepath.Join(data.dst, data.name)
		if err := os.MkdirAll(data.dst, os.ModePerm); err != nil {
			fmt.Fprint(os.Stderr, "Could not create folder %s", data.name)
			os.Exit(1)
		}
	}

	var err error

	if data.zip {
		err = zipFolder(data.src)
	} else {
		err = copyFolder(data.src, data.dst)
	}
	
	if err != nil {
		fmt.Fprint(os.Stderr, "Oh no! gobackitup encountered an error: %v\n", err)
		os.Exit(3)
	} else {
		fmt.Printf("Backup complete.\n")
		fmt.Printf("Saved to %s\n", data.dst)
		fmt.Printf("Processed %d files (%s total)\n", files, fileSize(bytesCopied))
		fmt.Printf("Encountered %d errors\n", errors)
	}
}
