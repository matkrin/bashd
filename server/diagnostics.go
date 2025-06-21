package server

import (
	"fmt"
	"os"

	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func checkFile(uri string, state State) []lsp.Diagnostic {
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

	return diagnostics
}

func diagnosticParseError(err error) lsp.Diagnostic {
	line := 0
	col := 0
	message := ""

	switch e := err.(type) {
	case syntax.ParseError:
		logger.Info("PARSEERROR")
		line = int(e.Pos.Line())
		col = int(e.Pos.Col())
		message = e.Text
	case syntax.LangError:
		logger.Info("LANGERROR")
		line = int(e.Pos.Line())
		col = int(e.Pos.Col())
		message = e.Feature
	case syntax.QuoteError:
		logger.Info("LANGERROR")
		message = e.Message
	default:
		logger.Infof("OTHER ERROR OF TYPE: %T", err)

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
				Line:      int(file.Start.Line()) - 1,
				Character: int(file.Start.Col()) - 1,
			},
			End: lsp.Position{
				Line:      int(file.End.Line()) - 1,
				Character: int(file.End.Col()) - 1,
			},
		},
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  fmt.Sprintf("File `%s` does not exist", file.Name),
	}
}
