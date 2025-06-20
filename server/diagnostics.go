package server

import (
	"fmt"
	"os"

	"github.com/matkrin/bashd/lsp"
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
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
			End: lsp.Position{
				Line:      0,
				Character: 1,
			},
		},
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  err.Error(),
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
