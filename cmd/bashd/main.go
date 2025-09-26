package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/server"
)

// Set at compile time
var VERSION string

func main() {
	fmt.Println(VERSION)
	name := "bashd"

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
	server := server.NewServer(name, VERSION, state, writer)

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
