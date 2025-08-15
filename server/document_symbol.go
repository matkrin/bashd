package server

import (
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleDocumentSymbol(request *lsp.DocumentSymbolsRequest, state *State) lsp.DocumentSymbolResponse {
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri]
	documentSymbols := []lsp.DocumentSymbol{}
	fileAst, err := parseDocument(document.Text, uri)
	if err != nil {
		return lsp.DocumentSymbolResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  &request.ID,
			},
			Result: documentSymbols,
		}
	}

	for _, node := range defNodes(fileAst) {
		var kind lsp.SymbolKind
		var startLine, startCol, endLine, endCol uint
		var selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol uint
		switch n := node.Node.(type) {
		case *syntax.FuncDecl:
			kind = lsp.SymbolFunction

			startLine = n.Pos().Line() - 1
			startCol = n.Pos().Col() - 1
			endLine = n.Body.End().Line() - 1
			endCol = n.Body.End().Col() - 1

			selectionStartLine = n.Pos().Line() - 1
			selectionStartCol = n.Pos().Col() - 1
			selectionEndLine = n.End().Line() - 1
			selectionEndCol = n.End().Col() - 1

		case *syntax.Assign:
			kind = lsp.SymbolVariable

			startLine = n.Pos().Line() - 1
			startCol = n.Pos().Col() - 1
			endLine = n.End().Line() - 1
			endCol = n.End().Col() - 1

			selectionStartLine = n.Name.Pos().Line() - 1
			selectionStartCol = n.Name.Pos().Col() - 1
			selectionEndLine = n.Name.End().Line() - 1
			selectionEndCol = n.Name.End().Col() - 1
		}

		documentSymbols = append(documentSymbols, lsp.DocumentSymbol{
			Name:  node.Name,
			Kind:  kind,
			Range: lsp.NewRange(startLine, startCol, endLine, endCol),
			SelectionRange: lsp.NewRange(
				selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol,
			),
			Children: []lsp.DocumentSymbol{},
		})
	}

	return lsp.DocumentSymbolResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: documentSymbols,
	}
}
