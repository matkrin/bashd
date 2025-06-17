package main

import (
	"bufio"
	"os"

	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
	"github.com/matkrin/bashd/server"
)

// func testParser() []*syntax.Lit {
// 	script := `#!/bin/bash
// a=5
// echo "$a"
// `
// 	// b="hello"
// 	// c=(1 2 3)
// 	// declare -i d
// 	// e=$((5 + 3))
// 	// declare -A f
// 	// f[red]=apple
// 	// readonly g=3.14
// 	// export h="var"
// 	// i=$(date)
// 	// j=""
// 	// echo ${k:-"default_value"}
// 	// echo 'hello world' | grep wo
// 	// cat file.txt
// 	// echo a
//
// 	reader := strings.NewReader(script)
// 	parser := syntax.NewParser()
// 	file, err := parser.Parse(reader, "test")
// 	if err != nil {
// 		fmt.Printf("%v\n", err)
// 	}
//
// 	fmt.Printf("%v\n", file.Name)
// 	fmt.Printf("%v\n", file.Last)
//
// 	// variables := ExtractVariableNames(file)
// 	// for i, v := range variables {
// 	// 	fmt.Printf("%v: ", i)
// 	// 	syntax.DebugPrint(os.Stdout, v)
// 	// }
// 	//
// 	return variables
// }

func main() {
	// testParser()

	logLevel := "debug"
	logger.Init("debug", "/Users/matthias/Developer/bashd/bashd.log")
	logger.Infof("Logging initialized with level: %s", logLevel)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(lsp.Split)

	state := server.NewState()
	writer := os.Stdout

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, contents, err := lsp.DecodeMessage(msg)
		if err != nil {
			logger.Error("Got an error: %s", err)
		}
		server.HandleMessage(writer, state, method, contents)
	}
}
