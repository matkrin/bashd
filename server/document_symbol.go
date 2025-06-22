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
		var startLine, startCol, endLine, endCol int
		var selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol int
		switch n := node.Node.(type) {
		case *syntax.FuncDecl:
			kind = lsp.SymbolFunction

			startLine = int(n.Pos().Line()) - 1
			startCol = int(n.Pos().Col()) - 1
			endLine = int(n.Body.End().Line()) - 1
			endCol = int(n.Body.End().Col()) - 1

			selectionStartLine = int(n.Pos().Line()) - 1
			selectionStartCol = int(n.Pos().Col()) - 1
			selectionEndLine = int(n.End().Line()) - 1
			selectionEndCol = int(n.End().Col()) - 1

		case *syntax.Assign:
			kind = lsp.SymbolVariable

			startLine = int(n.Pos().Line()) - 1
			startCol = int(n.Pos().Col()) - 1
			endLine = int(n.End().Line()) - 1
			endCol = int(n.End().Col()) - 1

			selectionStartLine = int(n.Name.Pos().Line()) - 1
			selectionStartCol = int(n.Name.Pos().Col()) - 1
			selectionEndLine = int(n.Name.End().Line()) - 1
			selectionEndCol = int(n.Name.End().Col()) - 1
		}

		documentSymbols = append(documentSymbols, lsp.DocumentSymbol{
			Name: node.Name,
			Kind: kind,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      startLine,
					Character: startCol,
				},
				End: lsp.Position{
					Line:      endLine,
					Character: endCol,
				},
			},
			SelctionRange: lsp.Range{
				Start: lsp.Position{
					Line:      selectionStartLine,
					Character: selectionStartCol,
				},
				End: lsp.Position{
					Line:      selectionEndLine,
					Character: selectionEndCol,
				},
			},
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
