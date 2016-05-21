package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	return err
}

func main() {
	var fromDirectory = flag.String("from", ".", "the directory from where files will be moved")
	var destinationDirectory = flag.String("to", "",
		"the base directory where files will be moved into subfolders")
	flag.Parse()
	if len(*destinationDirectory) == 0 {
		destinationDirectory = fromDirectory
	}

	fromDirectoryInfo, err := os.Stat(*fromDirectory)
	if err != nil {
		log.Fatal("missing source directory", *fromDirectory)
	}

	files, err := ioutil.ReadDir(*fromDirectory)
	if err != nil {
		log.Fatal(err)
		return
	}

	dirlist := make(map[string][]os.FileInfo)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		directory := filepath.FromSlash(fmt.Sprintf("%d/%d-%02d-%02d", file.ModTime().Year(), file.ModTime().Year(), file.ModTime().Month(), file.ModTime().Day()))
		dirlist[directory] = append(dirlist[directory], file)
	}

	for dir, files := range dirlist {
		fmt.Printf("Dir %s contains %d files\n", dir, len(files))
		targetDir := filepath.Join(*destinationDirectory, dir)
		if err = os.MkdirAll(targetDir, fromDirectoryInfo.Mode()); err != nil {
			log.Fatal("Cannot create directory ", targetDir)
		}

		for _, file := range files {
			var finalPath string
			for i := 0; i < 100; i++ {
				if i == 0 {
					finalPath = fmt.Sprintf("%s/%s/%s", *destinationDirectory, dir, file.Name())
				} else {
					finalPath = fmt.Sprintf("%s/%s/%d_%s", *destinationDirectory, dir, i, file.Name())
				}
				if _, err := os.Stat(finalPath); err == nil {
					// path already exists
					continue
				} else {
					err = moveFile(filepath.Join(*fromDirectory, file.Name()), finalPath)
					if err != nil {
						log.Fatal("cannot move from %s to %s", fromDirectory, finalPath)
					}
					fmt.Printf("\t- %s\n", finalPath)
					break
				}

			}
		}
	}
}
