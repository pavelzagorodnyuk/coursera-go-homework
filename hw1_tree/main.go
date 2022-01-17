package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	// checking the parameters
	if out == nil {
		return errors.New("invalid parameters: out contains nil")
	}

	if path == "" {
		return errors.New("invalid parameters: empty path")
	}

	// printing a tree
	return scanLevel(out, path, printFiles, "")
}

func scanLevel(out io.Writer, path string, printFiles bool, prefix string) error {

	// reading the contents of the path
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// if the files do not need to be printed, then delete them from the list
	if !printFiles {
		entries = cleanFromFiles(entries)
	}

	// printing the contents of the list
	for index, entry := range entries {
		if entry.IsDir() {

			// defining a new prefix
			var newPrefix string
			var branch rune
			if index < len(entries)-1 {
				branch = '├'
				newPrefix = prefix + "│\t"
			} else {
				branch = '└'
				newPrefix = prefix + "\t"
			}

			fmt.Fprintf(out, "%s%c───%s\n", prefix, branch, entry.Name())

			// getting ready to print a subdirectory
			newPath := strings.Join([]string{path, entry.Name()}, string(os.PathSeparator))

			if err = scanLevel(out, newPath, printFiles, newPrefix); err != nil {
				return err
			}

		} else {

			// determining the file size
			info, err := entry.Info()
			var size string
			if err != nil {
				size = "unknown"
			} else if info.Size() == 0 {
				size = "empty"
			} else {
				size = strconv.Itoa(int(info.Size())) + "b"
			}

			// defining the branch symbol
			var branch rune
			if index < len(entries)-1 {
				branch = '├'
			} else {
				branch = '└'
			}

			fmt.Fprintf(out, "%s%c───%s (%s)\n", prefix, branch, entry.Name(), size)
		}
	}

	return nil
}

func cleanFromFiles(entries []os.DirEntry) []os.DirEntry {

	// counting the number of directories
	var dirCount int
	for _, entry := range entries {
		if entry.IsDir() {
			dirCount++
		}
	}

	// making a list of directories
	var directories = make([]os.DirEntry, 0, dirCount)

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry)
		}
	}

	return directories
}
