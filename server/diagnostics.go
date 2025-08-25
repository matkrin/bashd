package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/matkrin/bashd/lsp"
	"github.com/matkrin/bashd/shellcheck"
	"mvdan.cc/sh/v3/syntax"
)

func checkDiagnostics(uri string, state *State) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	document := state.Documents[uri]
	fileAst, err := parseDocument(document.Text, uri)
	if err != nil {
		diagnostics = append(diagnostics, diagnosticParseError(err))
		return diagnostics
	}

	sourcedFiles := findSourceStatments(fileAst, state.EnvVars)
	for _, sourcedFile := range sourcedFiles {
		if _, err := os.Stat(sourcedFile.Name); err != nil {
			diagnostics = append(diagnostics, fileNotExistent(sourcedFile))
		}
	}

	shellcheck, err := shellcheck.Run(document.Text)
	if err != nil {
		return diagnostics
	}
	diagnostics = append(diagnostics, shellcheck.ToDiagnostics()...)

	return diagnostics
}

func checkDiagnosticsWorkspace(state *State) map[string][]lsp.Diagnostic {
	workspaceDiagnostics := map[string][]lsp.Diagnostic{}

	for _, shFile := range state.WorkspaceShFiles() {
		diagnostics := []lsp.Diagnostic{}
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			diagnostics = append(diagnostics, diagnosticParseError(err))
		}

		fileAst, err := parseDocument(string(fileContent), shFile)
		if err != nil {
			slog.Error("ERROR parsing", "file", shFile)
			continue
		}
		sourcedFiles := findSourceStatments(fileAst, state.EnvVars)
		for _, sourcedFile := range sourcedFiles {
			if _, err := os.Stat(sourcedFile.Name); err != nil {
				diagnostics = append(diagnostics, fileNotExistent(sourcedFile))
			}
		}

		shellcheck, err := shellcheck.Run(string(fileContent))
		if err != nil {
			slog.Error("ERROR running shellcheck", "err", err)
		} else {
			diagnostics = append(diagnostics, shellcheck.ToDiagnostics()...)
		}

		uri := pathToURI(shFile)
		workspaceDiagnostics[uri] = diagnostics
	}

	return workspaceDiagnostics
}

func diagnosticParseError(err error) lsp.Diagnostic {
	line := uint(0)
	col := uint(0)
	message := ""

	switch e := err.(type) {
	case syntax.ParseError:
		line = e.Pos.Line()
		col = e.Pos.Col()
		message = e.Text
	case syntax.LangError:
		line = e.Pos.Line()
		col = e.Pos.Col()
		message = e.Feature
	case syntax.QuoteError:
		message = e.Message
	default:
		slog.Info("Unknown parser error", "err", err)

	}

	return lsp.Diagnostic{
		Range:    lsp.NewRange(line, col, line, col),
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  message,
	}
}

func fileNotExistent(file SourcedFile) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.NewRange(file.Start.Line(),
			file.Start.Col(),
			file.End.Line(),
			file.End.Col(),
		),
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  fmt.Sprintf("File `%s` does not exist", file.Name),
	}
}
