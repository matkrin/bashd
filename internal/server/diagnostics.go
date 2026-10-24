package server

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/shellcheck"
	"github.com/matkrin/bashd/internal/utils"
	"mvdan.cc/sh/v3/syntax"
)

const shellCheckWorkers = 8

func findDiagnostics(
	documentText string,
	uri string,
	envVars map[string]string,
	shellcheckOptions shellcheck.Options,
) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)

	shellcheck, err := shellcheck.Run(documentText, shellcheckOptions)
	if err != nil {
		slog.Error("ERROR running shellcheck", "err", err)
	} else {
		diagnostics = append(diagnostics, shellcheck.ToDiagnostics()...)
	}

	fileAst, err := ast.ParseDocument(documentText, uri, false)
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
	shFiles := state.WorkspaceShFiles()
	workspaceDiagnostics := make(map[string][]lsp.Diagnostic, len(shFiles))

	type result struct {
		uri         string
		diagnostics []lsp.Diagnostic
	}

	jobsChan := make(chan string, shellCheckWorkers*2)
	resultsChan := make(chan result, shellCheckWorkers*2)

	var wg sync.WaitGroup

	for range shellCheckWorkers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for shFile := range jobsChan {
				fileContent, err := os.ReadFile(shFile)
				if err != nil {
					slog.Error("could not read file content", "file", shFile, "err", err)
					continue
				}

				uri := utils.PathToURI(shFile)
				diagnostics := findDiagnostics(
					string(fileContent),
					uri,
					state.EnvVars,
					state.Config.ShellCheckOptions,
				)

				resultsChan <- result{
					uri:         uri,
					diagnostics: diagnostics,
				}
			}
		}()
	}

	go func() {
		for _, shFile := range shFiles {
			jobsChan <- shFile
		}
		close(jobsChan)

		wg.Wait()
		close(resultsChan)
	}()

	for res := range resultsChan {
		workspaceDiagnostics[res.uri] = res.diagnostics
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
