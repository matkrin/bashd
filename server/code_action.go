package server

import (
	"bytes"
	"log/slog"

	"github.com/matkrin/bashd/lsp"
	"github.com/matkrin/bashd/shellcheck"
	"mvdan.cc/sh/v3/fileutil"
	"mvdan.cc/sh/v3/syntax"
)

var SHEBANG = "#!/usr/bin/env bash\n\n"

func handleCodeAction(request *lsp.CodeActionRequest, state *State) *lsp.CodeActionResponse {
	slog.Info("CODE ACTION", "range", request.Params.Range)
	slog.Info("CODE ACTION", "context", request.Params.Context)
	uri := request.Params.TextDocument.URI
	documentText := state.Documents[uri].Text
	hasShebang := fileutil.HasShebang([]byte(documentText))

	actions := []lsp.CodeAction{}
	if !hasShebang {
		action := shebangCodeAction(uri)
		actions = append(actions, *action)
	}

	actions = append(actions, *singleLineCodeAction(documentText, uri))

	context := request.Params.Context
	if len(context.Diagnostics) != 0 {
		actions = append(actions, shellCheckCodeActions(documentText, uri, context)...)
	}

	response := &lsp.CodeActionResponse{
		Response: lsp.Response{
			RPC: "2.0",
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
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      0,
								Character: 0,
							},
							End: lsp.Position{
								Line:      0,
								Character: 0,
							},
						},
						NewText: SHEBANG,
					},
				},
			},
		},
	}
	return action
}

func singleLineCodeAction(document string, uri string) *lsp.CodeAction {
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		return nil
	}

	singleLine := syntax.SingleLine(true)

	printer := syntax.NewPrinter(singleLine)

	buffer := bytes.NewBuffer([]byte{})
	printer.Print(buffer, fileAst)

	action := &lsp.CodeAction{
		Title: "Single Line",
		Edit: lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					lsp.TextEdit{
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      0,
								Character: 0,
							},
							End: lsp.Position{
								Line:      9999,
								Character: 9999,
							},
						},
						NewText: buffer.String(),
					},
				},
			},
		},
	}
	return action
}

func shellCheckCodeActions(documentText string, uri string, context lsp.CodeActionContext) []lsp.CodeAction {
	actions := []lsp.CodeAction{}
	shellcheck, err := shellcheck.Run(documentText)
	if err != nil {
		slog.Error("ERROR running shellcheck", "err", err)
	}
	if shellcheck != nil {
		for _, comment := range shellcheck.Comments {
			shellcheckDiagnostic := comment.ToDiagnostic()
			for _, contextDiagnostic := range context.Diagnostics {
				if shellcheckDiagnostic.Range == contextDiagnostic.Range {
					action := comment.ToCodeAction(uri)
					if action != nil {
						actions = append(actions, *action)

					}
				}
			}
		}
	}

	return actions
}
