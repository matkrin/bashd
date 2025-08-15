package server

import (
	"log/slog"
	"os"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

// TODO: Only for .sh files (could also do files without extension and check #!)
func handleWorkspaceSymbol(request *lsp.WorkspaceSymbolRequest, state *State) *lsp.WorkspaceSymbolResponse {
	slog.Info("Workspace Symbols Query", "query", request.Params.Query)
	shFiles := state.WorkspaceShFiles()

	workspaceSymbols := []lsp.WorkspaceSymbol{}
	for _, shFile := range shFiles {
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			slog.Error("Could not read file", "file", shFile)
			continue
		}
		fileAst, err := parseDocument(string(fileContent), shFile)

		for _, node := range defNodes(fileAst) {
			var kind lsp.SymbolKind
			var startLine, startCol, endLine, endCol uint
			switch n := node.Node.(type) {
			case *syntax.FuncDecl:
				kind = lsp.SymbolFunction

				startLine = n.Pos().Line() - 1
				startCol = n.Pos().Col() - 1
				endLine = n.Body.End().Line() - 1
				endCol = n.Body.End().Col() - 1

			case *syntax.Assign:
				kind = lsp.SymbolVariable

				startLine = n.Pos().Line() - 1
				startCol = n.Pos().Col() - 1
				endLine = n.End().Line() - 1
				endCol = n.End().Col() - 1

			}

			workspaceSymbols = append(workspaceSymbols, lsp.WorkspaceSymbol{
				Name: node.Name,
				Kind: kind,
				Location: lsp.Location{
					URI: pathToURI(shFile),
					Range: lsp.NewRange(
						startLine,
						startCol,
						endLine,
						endCol,
					),
				},
			})
		}
	}
	response := lsp.WorkspaceSymbolResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: workspaceSymbols,
	}
	return &response

}
