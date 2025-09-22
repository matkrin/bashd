package server

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/matkrin/bashd/ast"
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
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
	fileAst, err := ast.ParseDocument(documentText, uri)
	if err != nil {
		slog.Error(err.Error())
	}
	cursorNode := fileAst.FindNodeUnderCursor(cursor)
	referenceNodes := fileAst.FindRefsInFile(cursorNode, params.Context.IncludeDeclaration)

	locations := []lsp.Location{}
	for _, refNode := range referenceNodes {
		locations = append(locations, refNode.ToLspLocation(uri))
	}

	// In sourced files
	filename, err := uriToPath(uri)
	if err != nil {
		slog.Error(err.Error())
	}
	baseDir := filepath.Dir(filename)
	referenceNodesInSourcedFiles := fileAst.FindRefsinSourcedFile(
		cursorNode,
		state.EnvVars,
		baseDir,
		params.Context.IncludeDeclaration,
	)

	for file, refNodes := range referenceNodesInSourcedFiles {
		for _, refNode := range refNodes {
			locations = append(locations, refNode.ToLspLocation(pathToURI(file)))
		}
	}

	// In workspace files that source current file
	refNodesInWorkspaceFile := findRefsInWorkspaceFiles(
		uri,
		state.WorkspaceShFiles(),
		cursorNode,
		state.EnvVars,
		params.Context.IncludeDeclaration,
	)

	for file, refNodes := range refNodesInWorkspaceFile {
		for _, refNode := range refNodes {
			location := refNode.ToLspLocation(pathToURI(file))
			if !slices.Contains(locations, location) {
				locations = append(locations, location)
			}
		}
	}

	response := lsp.ReferencesResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: locations,
	}

	return &response
}

// Find reference nodes in all workspace files if they source the current file
func findRefsInWorkspaceFiles(
	uri string,
	workspaceShFiles []string,
	cursorNode syntax.Node,
	env map[string]string,
	includeDeclaration bool,
) map[string][]ast.RefNode {
	referenceNodes := map[string][]ast.RefNode{}
	for _, shFile := range workspaceShFiles {
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			slog.Error("ERROR: Could not read file", "file", shFile)
			continue
		}
		workspaceFileAst, err := ast.ParseDocument(string(fileContent), shFile)
		if err != nil {
			slog.Error("ERROR: Could not parse file", "file", shFile)
			continue
		}
		for _, sourceStatement := range workspaceFileAst.FindSourceStatments(env) {
			baseDir := filepath.Dir(shFile)
			path := sourceStatement.SourcedFile
			resolved := path
			if !filepath.IsAbs(path) {
				resolved = filepath.Join(baseDir, path)
			}
			resolved = filepath.Clean(resolved)
			slog.Info("REFERENCES", "resolved", resolved)
			if uri == pathToURI(resolved) {
				refs := workspaceFileAst.FindRefsInFile(cursorNode, includeDeclaration)
				slog.Info("REFERENCES", "refs", refs)
				referenceNodes[shFile] = append(referenceNodes[shFile], refs...)
			}
		}
	}
	return referenceNodes
}
