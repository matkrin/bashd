package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/matkrin/bashd/lsp"
)

func handleHover(request *lsp.HoverRequest, state *State) *lsp.HoverResponse {

	documentName := request.Params.TextDocument.URI
	cursor := newCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)
	document := state.Documents[documentName].Text
	fileAst, err := parseDocument(document, documentName)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	node := findNodeUnderCursor(fileAst, cursor)
	if node == nil {
		return nil
	}

	identifier := extractIdentifier(node)
	documentation := getDocumentation(identifier)
	if strings.Trim(documentation, "\n") == "" {
		return nil
	}
	mdDocumentation := fmt.Sprintf("```man\n%s\n```", documentation)
	response := lsp.HoverResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: lsp.HoverResult{
			Contents: lsp.MarkupContent{
				Kind:  lsp.MarkupKindMarkdown,
				Value: mdDocumentation,
			},
		},
	}
	return &response
}
