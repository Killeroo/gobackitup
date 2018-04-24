// TODO: Use goroutines for copy contents of files (No point limited by usb read write anyway)
// TODO: change to capitals first (https://talks.golang.org/2014/names.slide#8) 
// TODO: Add some colours and maybe file sizes// TODO: Autocreate backup director
package main

import (
	"path/filepath"
	"archive/zip"
	"path"
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

func FileSize(size float64) (result string) {
	// TODO: return kb mb or gb as string using 
}

// TODO: Add a description
func CopyFolder(src, dst string) (err error) {

	filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
		var dst = strings.Replace(path, data.src, "", -1)
		dst = filepath.Join(data.dst, dst)

		// TODO: dont print when its agolder
		
		fmt.Printf("Copying: %s [%dMB] -> ", path, (f.Size()/1024))
		
		err = copyFile(path, dst)
		if err != nil {
			fmt.Printf("CopyFile failed %q\n", err)
		} else {
			fmt.Printf("CopyFile succeeded\n")
		}		
		return nil // TODO: really?
	})	
	return err
}

// Zips up folder to destination path, new zip file is names after
// base file of the source directory.
// First we create a zip folder at the destination path named after
// the base directory of the source. Then we create an archive object
// for our new zip file.
// Next walk through the src directory and for each file we create a
// zip info header. If its a directory add a path seperator, otherwise
// if its a file add a deflate method then create the header in the archive.
// Lastly if its a file we copy over the file to the archive using an
// writer from the archive.
func ZipFolder(src, dst string) (err error) {
	filename := filepath.Join(data.dst, path.Base(data.src) + ".zip") // Windows bug here
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
			return err
		}
		header.Name = filepath.Join(base, strings.TrimPrefix(path, src))
		fmt.Printf("Zipping: %s -> %s\n", path, header.Name)

		if f.IsDir() {
			header.Name += "/" //os.PathSeparator // TODO: Bug here
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
		

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		} 
		defer srcFile.Close()
		_, err = io.Copy(writer, srcFile)
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
	flag.BoolVar(&data.zip, "z", false, "\n(optional) (shorthand) Compress the backup")
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

	var err error

	if data.zip {
		err = ZipFolder(data.src, data.dst)	
	} else {
		err = CopyFolder(data.src, data.dst)
		//err = filepath.Walk(data.src, handle)
	}
	
	if err != nil {
		fmt.Fprint(os.Stderr, "Oh no! gobackitup encountered an error: %v\n", err)
		os.Exit(3)
	}
}
