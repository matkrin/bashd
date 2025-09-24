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
	// First, extract the identifier we're looking for
	targetIdentifier := ast.ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return map[string][]ast.RefNode{}
	}

	// Get the original file path from URI for definition resolution
	originalFilePath, err := uriToPath(uri)
	if err != nil {
		slog.Error("Could not transform URI to path", "file", originalFilePath)
		return map[string][]ast.RefNode{}
	}

	// Load the original file to find the definition
	originalFileContent, err := os.ReadFile(originalFilePath)
	if err != nil {
		slog.Error("Could not read original file", "file", originalFilePath)
		return map[string][]ast.RefNode{}
	}

	originalAst, err := ast.ParseDocument(string(originalFileContent), originalFilePath)
	if err != nil {
		slog.Error("Could not parse original file", "file", originalFilePath)
		return map[string][]ast.RefNode{}
	}

	// Find the definition in the original file using cross-file logic
	baseDir := filepath.Dir(originalFilePath)
	targetFile, defNode := originalAst.FindDefinitionAcrossFiles(cursorNode, env, baseDir)

	slog.Info("WORKSPACE_REFS", "targetIdentifier", targetIdentifier, "targetFile", targetFile, "defNode", defNode)

	referenceNodes := map[string][]ast.RefNode{}

	// Search through all workspace files
	for _, shFile := range workspaceShFiles {
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			slog.Error("Could not read file", "file", shFile)
			continue
		}

		workspaceFileAst, err := ast.ParseDocument(string(fileContent), shFile)
		if err != nil {
			slog.Error("Could not parse file", "file", shFile)
			continue
		}

		// Check if this file sources the target file (where the definition is)
		shouldIncludeFile := false

		// Method 1: Check if this file sources the target file
		for _, sourceStatement := range workspaceFileAst.FindSourceStatments(env) {
			baseDir := filepath.Dir(shFile)
			path := sourceStatement.SourcedFile
			resolved := path
			if !filepath.IsAbs(path) {
				resolved = filepath.Join(baseDir, path)
			}
			resolved = filepath.Clean(resolved)

			if uri == pathToURI(resolved) {
				shouldIncludeFile = true
				slog.Info("File sources target", "file", shFile, "target", resolved)
				break
			}
		}

		// Method 2: Also include the file if it's the target file itself
		if shFile == targetFile {
			shouldIncludeFile = true
			slog.Info("File is target file", "file", shFile)
		}

		// Method 3: Include if both files are in the same sourcing chain
		if !shouldIncludeFile {
			// Check if the workspace file is in the same sourcing chain
			sourcedFiles := originalAst.FindAllSourcedFiles(env, baseDir, map[string]bool{})
			if slices.Contains(sourcedFiles, shFile) {
					shouldIncludeFile = true
					slog.Info("File in sourcing chain", "file", shFile)
				}
		}

		if !shouldIncludeFile {
			continue
		}

		// Find references in this file
		var refs []ast.RefNode

		if defNode == nil {
			// No definition found - fall back to simple name matching
			slog.Info("No definition found, using name matching", "file", shFile)
			for _, refNode := range workspaceFileAst.RefNodes(includeDeclaration) {
				if refNode.Name == targetIdentifier {
					refs = append(refs, refNode)
				}
			}
		} else {
			// Definition found - use proper scoping logic
			slog.Info("Using scoped reference finding", "file", shFile)
			for _, refNode := range workspaceFileAst.RefNodes(includeDeclaration) {
				if refNode.Name != targetIdentifier {
					continue
				}

				// Use cross-file resolution logic
				if workspaceFileAst.WouldResolveToSameDefinitionAcrossFiles(refNode.Node, defNode, targetFile, shFile) {
					refs = append(refs, refNode)
				}
			}
		}

		if len(refs) > 0 {
			referenceNodes[shFile] = refs
			slog.Info("Found references", "file", shFile, "count", len(refs))
		}
	}

	return referenceNodes
}
