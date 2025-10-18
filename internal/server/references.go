package server

import (
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/utils"
)

func handleReferences(request *lsp.ReferencesRequest, state *State) *lsp.ReferencesResponse {
	params := request.Params
	uri := params.TextDocument.URI
	cursor := ast.NewCursor(
		params.Position.Line,
		params.Position.Character,
	)

	// In current file
	documentText := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(documentText, uri, false)
	if err != nil {
		slog.Error("Could not parse document", "err", err.Error())
		return nil
	}
	referenceNodes := fileAst.FindRefsInFile(cursor, params.Context.IncludeDeclaration)

	locations := []lsp.Location{}
	for _, refNode := range referenceNodes {
		locations = append(locations, refNode.ToLspLocation(uri))
	}

	// In sourced files
	filename, err := utils.UriToPath(uri)
	if err != nil {
		slog.Error(err.Error())
	}
	baseDir := filepath.Dir(filename)
	referenceNodesInSourcedFiles := fileAst.FindRefsinSourcedFile(
		cursor,
		state.EnvVars,
		baseDir,
		params.Context.IncludeDeclaration,
	)

	for file, refNodes := range referenceNodesInSourcedFiles {
		for _, refNode := range refNodes {
			locations = append(locations, refNode.ToLspLocation(utils.PathToURI(file)))
		}
	}

	// In workspace files that source current file
	refNodesInWorkspaceFile := fileAst.FindRefsInWorkspaceFiles(
		uri,
		state.WorkspaceShFiles(),
		cursor,
		state.EnvVars,
		params.Context.IncludeDeclaration,
	)

	for file, refNodes := range refNodesInWorkspaceFile {
		for _, refNode := range refNodes {
			location := refNode.ToLspLocation(utils.PathToURI(file))
			if !slices.Contains(locations, location) {
				locations = append(locations, location)
			}
		}
	}

	response := lsp.ReferencesResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
			ID:  &request.ID,
		},
		Result: locations,
	}

	return &response
}

