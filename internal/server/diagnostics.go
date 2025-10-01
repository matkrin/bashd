package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/shellcheck"
	"github.com/matkrin/bashd/internal/utils"
	"mvdan.cc/sh/v3/syntax"
)

func findDiagnostics(
	documentText string,
	uri string,
	envVars map[string]string,
) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}

	shellcheck, err := shellcheck.Run(documentText)
	if err != nil {
		slog.Error("ERROR running shellcheck", "err", err)
	} else {
		diagnostics = append(diagnostics, shellcheck.ToDiagnostics()...)
	}

	fileAst, err := ast.ParseDocument(documentText, uri)
	if err != nil {
		diagnostics = append(diagnostics, diagnosticParseError(err))
		return diagnostics
	}

	for _, sourceStatement := range fileAst.FindSourceStatments(envVars) {
		if _, err := os.Stat(sourceStatement.SourcedFile); err != nil {
			diagnostics = append(diagnostics, fileNotExistentError(sourceStatement))
		}
	}

	return diagnostics
}

func findDiagnosticsWorkspace(state *State) map[string][]lsp.Diagnostic {
	workspaceDiagnostics := map[string][]lsp.Diagnostic{}

	for _, shFile := range state.WorkspaceShFiles() {
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			slog.Error("ERROR could not read file content", "file", shFile)
		}

		uri := utils.PathToURI(shFile)
		diagnostics := findDiagnostics(string(fileContent), uri, state.EnvVars)
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
		line = e.Pos.Line() - 1
		col = e.Pos.Col() - 1
		message = e.Text
	case syntax.LangError:
		line = e.Pos.Line() - 1
		col = e.Pos.Col() - 1
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

func fileNotExistentError(file ast.SourceStatement) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.NewRange(file.StartLine,
			file.StartChar,
			file.EndLine,
			file.EndChar,
		),
		Severity: lsp.DiagnosticError,
		Code:     nil,
		Source:   "bashd",
		Message:  fmt.Sprintf("File `%s` does not exist", file.SourcedFile),
	}
}
