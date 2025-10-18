package server

import (
	"bytes"
	"log/slog"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/shellcheck"
	"mvdan.cc/sh/v3/fileutil"
	"mvdan.cc/sh/v3/syntax"
)

var SHEBANG = "#!/usr/bin/env bash\n\n"

func handleCodeAction(request *lsp.CodeActionRequest, state *State) *lsp.CodeActionResponse {
	slog.Debug("CODE ACTION", "range", request.Params.Range)
	slog.Debug("CODE ACTION", "context", request.Params.Context)
	uri := request.Params.TextDocument.URI
	documentText := state.Documents[uri].Text
	hasShebang := fileutil.HasShebang([]byte(documentText))

	actions := []lsp.CodeAction{}
	if !hasShebang {
		action := shebangCodeAction(uri)
		actions = append(actions, *action)
	}

	shellcheck, err := shellcheck.Run(documentText, state.Config.ShellCheckOptions)
	if err == nil {
		// Fix all auto-fixable
		if shellcheck.ContainsFixable() {
			actions = append(actions, shellcheck.ToCodeActionFlat(uri))
		}

		// Fix for certain lint (position dependent)
		context := request.Params.Context
		if len(context.Diagnostics) != 0 {
			actions = append(actions, shellcheckCodeActions(shellcheck, uri, documentText, context)...)
		}
	}

	if fileAst, err := ast.ParseDocument(documentText, uri, false); err == nil {
		actions = append(actions, *minifyCodeAction(fileAst, uri))
	}

	response := &lsp.CodeActionResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
			ID:  &request.ID,
		},
		Result: actions,
	}
	return response
}

func shebangCodeAction(uri string) *lsp.CodeAction {
	action := &lsp.CodeAction{
		Title: "Add shebang",
		Edit: lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					lsp.TextEdit{
						Range:   lsp.NewRange(0, 0, 0, 0),
						NewText: SHEBANG,
					},
				},
			},
		},
	}
	return action
}

func minifyCodeAction(fileAst *ast.Ast, uri string) *lsp.CodeAction {
	singleLine := syntax.SingleLine(true)
	printer := syntax.NewPrinter(singleLine)
	buffer := bytes.NewBuffer([]byte{})
	printer.Print(buffer, fileAst.File)

	action := &lsp.CodeAction{
		Title: "Minify script",
		Edit: lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					lsp.TextEdit{
						Range:   lsp.NewRange(0, 0, 9999, 9999),
						NewText: buffer.String(),
					},
				},
			},
		},
	}
	return action
}

func shellcheckCodeActions(
	shellcheck *shellcheck.ShellCheckResult,
	uri string,
	documentText string,
	context lsp.CodeActionContext,
) []lsp.CodeAction {
	actions := []lsp.CodeAction{}
	for _, comment := range shellcheck.Comments {
		shellcheckDiagnostic := comment.ToDiagnostic()
		for _, contextDiagnostic := range context.Diagnostics {
			if shellcheckDiagnostic.Range != contextDiagnostic.Range {
				continue
			}
			// Lint fix
			if actionFixLint := comment.ToCodeActionFixLint(uri); actionFixLint != nil {
				actions = append(actions, *actionFixLint)
			}

			// Add ignore comment
			if actionIgnore := comment.ToCodeActionIgnore(
				uri,
				documentText,
				&contextDiagnostic.Range,
			); actionIgnore != nil {
				actions = append(actions, *actionIgnore)
			}
		}
	}

	return actions
}
