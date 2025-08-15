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
		slog.Error("ERROR running shellcheck", "err", err)
		return diagnostics
	}
	diagnostics = append(diagnostics, shellcheck.ToDiagnostics()...)

	return diagnostics
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
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      line,
				Character: col,
			},
			End: lsp.Position{
				Line:      line,
				Character: col,
			},
		},
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  message,
	}
}

func fileNotExistent(file SourcedFile) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      file.Start.Line() - 1,
				Character: file.Start.Col() - 1,
			},
			End: lsp.Position{
				Line:      file.End.Line() - 1,
				Character: file.End.Col() - 1,
			},
		},
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  fmt.Sprintf("File `%s` does not exist", file.Name),
	}
}
