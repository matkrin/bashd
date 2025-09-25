package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/server"
	"mvdan.cc/sh/v3/syntax"
)

func testParser() {
	script := `f () {
	local a b c
}`

	// b="hello"
	// c=(1 2 3)
	// declare -i d
	// e=$((5 + 3))
	// declare -A f
	// f[red]=apple
	// readonly g=3.14
	// export h="var"
	// i=$(date)
	// j=""
	// echo ${k:-"default_value"}
	// echo 'hello world' | grep wo
	// cat file.txt
	// echo a

	reader := strings.NewReader(script)
	parser := syntax.NewParser()
	file, err := parser.Parse(reader, "test")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	syntax.DebugPrint(os.Stdout, file)
	// syntax.Walk(file, func(node syntax.Node) bool {
	// 	return true
	// })
}

func main() {
	// testParser()
	name := "bashd"
	version := "0.1.0a1"

	logLevel := "debug"
	logFile := filepath.Join(os.Getenv("HOME"), "Developer", "bashd", "bashd.log")
	initLogging(logLevel, logFile)
	slog.Info("Logging initialized", "level", logLevel)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(lsp.Split)

	config := server.Config{
		ExcludeDirs: []string{".git", ".venv", "node_modules"},
	}

	state := server.NewState(config)
	writer := os.Stdout
	server := server.NewServer(name, version, state, writer)

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, contents, err := lsp.DecodeMessage(msg)
		if err != nil {
			slog.Error("ERROR decoding message", "err", err)
		}
		server.HandleMessage(method, contents)
	}
}

func initLogging(levelStr string, filename string) {

	var logger *slog.Logger
	level := new(slog.LevelVar)

	var l slog.Level
	switch levelStr {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	level.Set(l)

	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("No log file")
	}

	handler := slog.NewTextHandler(logfile, &slog.HandlerOptions{
		Level: level,
	})

	logger = slog.New(handler)
	slog.SetDefault(logger)
}
