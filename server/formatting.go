package server

import (
	"bytes"
	"log/slog"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleFormatting(request *lsp.FormattingRequest, state *State) *lsp.FormattingResponse {
	slog.Info("FORMATTING", "params", request.Params)
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		return nil
	}

	var indentWidth uint
	if request.Params.Options.InsertSpaces {
		indentWidth = request.Params.Options.TabSize
	} else {
		indentWidth = 0
	}
	indent := syntax.Indent(indentWidth)

	printer := syntax.NewPrinter(indent)

	buffer := bytes.NewBuffer([]byte{})
	printer.Print(buffer, fileAst)

	textedit := lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
			End: lsp.Position{
				Line:      99999,
				Character: 99999,
			},
		},
		NewText: buffer.String(),
	}

	response := lsp.FormattingResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: []lsp.TextEdit{textedit},
	}
	return &response
}
