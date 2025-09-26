package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func main() {
	args := os.Args
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
