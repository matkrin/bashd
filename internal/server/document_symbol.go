package server

import (
	"log/slog"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleDocumentSymbol(request *lsp.DocumentSymbolsRequest, state *State) *lsp.DocumentSymbolResponse {
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri]
	documentSymbols := []lsp.DocumentSymbol{}
	fileAst, err := ast.ParseDocument(document.Text, uri, false)
	if err != nil {
		slog.Error("Could not parse document", "document", uri)
		return nil
	}

	documentSymbols = findDocumentSymbols(fileAst.DefNodes())

	response := lsp.NewDocumentSymbolResponse(request.ID, documentSymbols)
	return &response
}

func findDocumentSymbols(defNodes []ast.DefNode) []lsp.DocumentSymbol {
	locals := findLocals(defNodes)
	documentSymbols := []lsp.DocumentSymbol{}

	for _, defNode := range defNodes {
		var kind lsp.SymbolKind
		var endLine, endCol uint
		var selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol uint
		var children []lsp.DocumentSymbol

		switch n := defNode.Node.(type) {
		// Function declaration
		case *syntax.FuncDecl:
			kind = lsp.SymbolFunction

			endLine = n.Body.End().Line() - 1
			endCol = n.Body.End().Col() - 1

			selectionStartLine = n.Pos().Line() - 1
			selectionStartCol = n.Pos().Col() - 1
			selectionEndLine = n.End().Line() - 1
			selectionEndCol = n.End().Col() - 1
			children = locals[n.Name.Value]

		case *syntax.DeclClause:
			if defNode.IsScoped {
				continue
			}
			kind = lsp.SymbolVariable

			endLine = n.End().Line() - 1
			endCol = n.End().Col() - 1

			selectionStartLine = defNode.StartLine - 1
			selectionStartCol = defNode.StartChar - 1
			selectionEndLine = defNode.EndLine - 1
			selectionEndCol = defNode.EndChar - 1

		// Variable assignement
		case *syntax.Assign:
			kind = lsp.SymbolVariable

			endLine = n.End().Line() - 1
			endCol = n.End().Col() - 1

			selectionStartLine = n.Name.Pos().Line() - 1
			selectionStartCol = n.Name.Pos().Col() - 1
			selectionEndLine = n.Name.End().Line() - 1
			selectionEndCol = n.Name.End().Col() - 1

		// Loops/ `read` statements
		case *syntax.ForClause, *syntax.CallExpr:
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

		documentSymbols = append(documentSymbols, lsp.DocumentSymbol{
			Name:  defNode.Name,
			Kind:  kind,
			Range: lsp.NewRange(startLine, startCol, endLine, endCol),
			SelectionRange: lsp.NewRange(
				selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol,
			),
			Children: children,
		})
	}
	return documentSymbols
}

func findLocals(defNodes []ast.DefNode) map[string][]lsp.DocumentSymbol {
	locals := map[string][]lsp.DocumentSymbol{}

	for _, defNode := range defNodes {
		decl, ok := defNode.Node.(*syntax.DeclClause)
		if !ok {
			continue
		}

		if !defNode.IsScoped {
			continue
		}

		startLine := decl.Pos().Line() - 1
		startCol := decl.Pos().Col() - 1
		endLine := decl.End().Line() - 1
		endCol := decl.End().Col() - 1

		selectionStartLine := defNode.StartLine - 1
		selectionStartCol := defNode.StartChar - 1
		selectionEndLine := defNode.EndLine - 1
		selectionEndCol := defNode.EndChar - 1
		funcName := defNode.Scope.Name.Value

		locals[funcName] = append(locals[funcName], lsp.DocumentSymbol{
			Name:  defNode.Name,
			Kind:  lsp.SymbolVariable,
			Range: lsp.NewRange(startLine, startCol, endLine, endCol),
			SelectionRange: lsp.NewRange(
				selectionStartLine, selectionStartCol, selectionEndLine, selectionEndCol,
			),
			Children: []lsp.DocumentSymbol{},
		})
	}

	return locals
}
