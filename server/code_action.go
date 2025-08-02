package server

import (
	"bytes"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/fileutil"
	"mvdan.cc/sh/v3/syntax"
)

var SHEBANG = "#!/usr/bin/env bash\n\n"

func handleCodeAction(request *lsp.CodeActionRequest, state *State) *lsp.CodeActionResponse {
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text
	hasShebang := fileutil.HasShebang([]byte(document))

	actions := []lsp.CodeAction{}
	if !hasShebang {
		action := shebangCodeAction(uri)
		actions = append(actions, *action)
	}

	actions = append(actions, *singleLineCodeAction(document, uri))

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
