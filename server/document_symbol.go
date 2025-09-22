package server

import (
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleDocumentSymbol(request *lsp.DocumentSymbolsRequest, state *State) *lsp.DocumentSymbolResponse {
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri]
	documentSymbols := []lsp.DocumentSymbol{}
	fileAst, err := parseDocument(document.Text, uri)
	if err != nil {
		return nil
	}

	for _, defNode := range defNodes(fileAst) {
		documentSymbols = append(documentSymbols, findDocumentSymbol(&defNode))
	}

	response := lsp.NewDocumentSymbolResponse(request.ID, documentSymbols)
	return &response
}

func findDocumentSymbol(defNode *DefNode) lsp.DocumentSymbol {
	var kind lsp.SymbolKind
	var endLine, endCol uint
	var selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol uint
	switch n := defNode.Node.(type) {
	case *syntax.FuncDecl:
		kind = lsp.SymbolFunction

		endLine = n.Body.End().Line() - 1
		endCol = n.Body.End().Col() - 1

		selectionStartLine = n.Pos().Line() - 1
		selectionStartCol = n.Pos().Col() - 1
		selectionEndLine = n.End().Line() - 1
		selectionEndCol = n.End().Col() - 1

	case *syntax.Assign:
		kind = lsp.SymbolVariable

		endLine = n.End().Line() - 1
		endCol = n.End().Col() - 1

		selectionStartLine = n.Name.Pos().Line() - 1
		selectionStartCol = n.Name.Pos().Col() - 1
		selectionEndLine = n.Name.End().Line() - 1
		selectionEndCol = n.Name.End().Col() - 1

	case *syntax.ForClause:
		kind = lsp.SymbolVariable

		endLine = n.End().Line() - 1
		endCol = n.End().Col() - 1

		selectionStartLine = defNode.StartLine - 1
		selectionStartCol = defNode.StartChar - 1
		selectionEndLine = defNode.EndLine - 1
		selectionEndCol = defNode.EndChar - 1
	}

	startLine := defNode.Node.Pos().Line() - 1
	startCol := defNode.Node.Pos().Col() - 1

	return lsp.DocumentSymbol{
		Name:  defNode.Name,
		Kind:  kind,
		Range: lsp.NewRange(startLine, startCol, endLine, endCol),
		SelectionRange: lsp.NewRange(
			selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol,
		),
		Children: []lsp.DocumentSymbol{},
	}
}
