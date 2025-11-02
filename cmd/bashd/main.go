package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/server"
	"github.com/matkrin/bashd/internal/shellcheck"
	"github.com/spf13/pflag"
)

// Set at compile time
var VERSION string

func main() {
	name := "bashd"

	logFileOpt := pflag.StringP("logfile", "l", "", "Log to FILE instead of stderr")
	jsonOpt := pflag.BoolP("json", "j", false, "Log in JSON format")
	verbosityOpt := pflag.CountP("verbose", "v", "Increase log message verbosity")
	versionOpt := pflag.BoolP("version", "V", false, "Print version")

	severityOpt := pflag.StringP("severity", "S", "style", "Minimum severity of errors to consider")

	shellcheckIncludeOpt := pflag.StringSlice("shellcheck-include", []string{}, "Only include ShellCheck lints")
	shellcheckExcludeOpt := pflag.StringSlice("shellcheck-exclude", []string{}, "Exclude ShellCheck lints")
	shellcheckEnableOpt := pflag.StringSlice("shellcheck-enable", []string{}, "Enable ShellCheck optional lints")

	fmtBinaryNextLineOpt := pflag.Bool("fmt-binary-next-line", false, "Binary ops start a line")
	fmtCaseIndentOpt := pflag.Bool("fmt-case-indent", false, "Switch cases will be indented")
	fmtSpaceRedirectsOpt := pflag.Bool("fmt-space-redirects", false, "Redirect operators will be followed by a space")
	fmtFuncNextLineOpt := pflag.Bool("fmt-func-next-line", false, "Function opening braces are placed on a separate line")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "bashd - Bash language server\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", name)
		fmt.Fprintf(os.Stderr, "Options:\n")
		pflag.PrintDefaults()
	}

	pflag.Parse()

	if !slices.Contains([]string{"error", "warning", "info", "style"}, *severityOpt) {
		fmt.Fprintf(os.Stderr, "illegal severity '%s'\n", *severityOpt)
		os.Exit(1)
	}

	if *versionOpt {
		fmt.Printf("%s %s\n", name, VERSION)
		os.Exit(0)
	}

	logFile, err := initLogging(*verbosityOpt, *logFileOpt, *jsonOpt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}
	if logFile != nil {
		defer logFile.Close()
	}

	shellcheckOptions := shellcheck.Options{
		Include:  *shellcheckIncludeOpt,
		Exclude:  *shellcheckExcludeOpt,
		Enable:   *shellcheckEnableOpt,
		Severity: *severityOpt,
	}

	formatOptions := server.FormatOptions{
		BinaryNextLine: *fmtBinaryNextLineOpt,
		CaseIndent:     *fmtCaseIndentOpt,
		SpaceRedirects: *fmtSpaceRedirectsOpt,
		FuncNextLine:   *fmtFuncNextLineOpt,
	}

	config := server.Config{
		ExcludeDirs:            []string{".git", ".venv", "node_modules"},
		DiagnosticDebounceTime: 200 * time.Millisecond,
		ShellCheckOptions:      shellcheckOptions,
		FormatOptions:          formatOptions,
	}

	state := server.NewState(config)
	writer := os.Stdout
	server := server.NewServer(name, VERSION, state, writer)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(lsp.Split)

	for scanner.Scan() {
		msg := make([]byte, len(scanner.Bytes()))
		copy(msg, scanner.Bytes())

		method, contents, err := lsp.DecodeMessage(msg)
		if err != nil {
			slog.Error("ERROR decoding message", "err", err)
		}
		server.HandleMessage(method, contents)
	}
}

func initLogging(verbosity int, filename string, json bool) (*os.File, error) {
	level := new(slog.LevelVar)
	var l slog.Level
	switch verbosity {
	case 3:
		l = slog.LevelDebug
	case 2:
		l = slog.LevelInfo
	case 1:
		l = slog.LevelWarn
	default:
		l = slog.LevelError
	}

	level.Set(l)
	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	var logfile *os.File
	var err error

	if filename != "" {
		logfile, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			return nil, fmt.Errorf("ERROR: failed to open log file: %w", err)
		}
		if json {
			handler = slog.NewJSONHandler(logfile, handlerOptions)
		} else {
			handler = slog.NewTextHandler(logfile, handlerOptions)
		}
	} else {
		if json {
			handler = slog.NewJSONHandler(os.Stderr, handlerOptions)
		} else {
			handler = slog.NewTextHandler(os.Stderr, handlerOptions)
		}
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logfile, nil
}
