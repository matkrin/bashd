package server

import (
	"log/slog"
	"os"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/utils"
	"mvdan.cc/sh/v3/syntax"
)

func handleWorkspaceSymbol(request *lsp.WorkspaceSymbolRequest, state *State) *lsp.WorkspaceSymbolResponse {
	shFiles := state.WorkspaceShFiles()

	var workspaceSymbols []lsp.WorkspaceSymbol
	for _, shFile := range shFiles {
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			slog.Error("Could not read file", "file", shFile)
			continue
		}

		fileAst, err := ast.ParseDocument(string(fileContent), shFile, false)
		if err != nil {
			slog.Error("Could not parse file", "file", shFile)
			continue
		}

		for _, defNode := range fileAst.DefNodes() {
			workspaceSymbols = append(
				workspaceSymbols,
				findWorkSpaceSymbol(&defNode, shFile),
			)
		}
	}
	response := lsp.NewWorkspaceSymbolResponse(request.ID, workspaceSymbols)
	return &response
}

func findWorkSpaceSymbol(defNode *ast.DefNode, filePath string) lsp.WorkspaceSymbol {
	var kind lsp.SymbolKind
	var  endLine, endCol uint
	switch n := defNode.Node.(type) {
	case *syntax.FuncDecl:
		kind = lsp.SymbolFunction

		endLine = n.Body.End().Line() - 1
		endCol = n.Body.End().Col() - 1

	case *syntax.Assign:
		kind = lsp.SymbolVariable

		endLine = n.End().Line() - 1
		endCol = n.End().Col() - 1

	case *syntax.ForClause, *syntax.CallExpr:
		kind = lsp.SymbolVariable

		endLine = defNode.EndLine - 1
		endCol = defNode.EndChar - 1

	}

	startLine := defNode.Node.Pos().Line() -1
	startCol := defNode.Node.Pos().Col() - 1

	return lsp.WorkspaceSymbol{
		Name: defNode.Name,
		Kind: kind,
		Location: lsp.Location{
			URI: utils.PathToURI(filePath),
			Range: lsp.NewRange(
				startLine,
				startCol,
				endLine,
				endCol,
			),
		},
	}
}
