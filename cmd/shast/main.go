package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"mvdan.cc/sh/v3/fileutil"
	"mvdan.cc/sh/v3/syntax"
)

func main() {

	isShFileOpt := pflag.BoolP("issh", "i", false, "Check if file is sh file")
	pflag.Parse()

	args := os.Args
	if *isShFileOpt {
		dirpath := args[2]
		filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
			fmt.Printf("In %s\n", path)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return err
			}
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			if fileutil.CouldBeScript2(d) == fileutil.ConfIfShebang {
				fmt.Printf("%s could be sh file\n", path)
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				data, err := io.ReadAll(file)
				if err != nil {
					return err
				}
				if fileutil.HasShebang(data) {
					fmt.Printf("%s has sheband\n", d.Name())
				}
			}
			return nil
		})
		os.Exit(0)
	}

	var script string
	if len(args) == 1 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: Cannot read stdin")
			os.Exit(1)
		}
		script = string(data)
	} else {
		script = args[1]
	}
	fmt.Println("script: ", script)

	reader := strings.NewReader(script)
	parser := syntax.NewParser()
	file, err := parser.Parse(reader, "test")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Cannot parse script")
		os.Exit(1)
	}

	syntax.DebugPrint(os.Stdout, file)
}
