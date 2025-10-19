package ast

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/matkrin/bashd/internal/utils"
)

// Find reference nodes in all workspace files if they source the current file
func (a *Ast) FindRefsInWorkspaceFiles(
	uri string,
	workspaceShFiles []string,
	cursor Cursor,
	env map[string]string,
	includeDeclaration bool,
) map[string][]RefNode {
	// First, extract the identifier we're looking for
	cursorNode := a.FindNodeUnderCursor(cursor)
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return map[string][]RefNode{}
	}

	// Get the original file path from URI for definition resolution
	originalFilePath, err := utils.UriToPath(uri)
	if err != nil {
		slog.Error("Could not transform URI to path", "file", originalFilePath)
		return map[string][]RefNode{}
	}

	// Load the original file to find the definition
	originalFileContent, err := os.ReadFile(originalFilePath)
	if err != nil {
		slog.Error("Could not read original file", "file", originalFilePath)
		return map[string][]RefNode{}
	}

	originalAst, err := ParseDocument(string(originalFileContent), originalFilePath, false)
	if err != nil {
		slog.Error("Could not parse original file", "file", originalFilePath)
		return map[string][]RefNode{}
	}

	// Find the definition in the original file using cross-file logic
	baseDir := filepath.Dir(originalFilePath)
	targetFile, defNode := originalAst.FindDefinitionAcrossFiles(cursor, env, baseDir)

	slog.Info("WORKSPACE_REFS", "targetIdentifier", targetIdentifier, "targetFile", targetFile, "defNode", defNode)

	referenceNodes := map[string][]RefNode{}

	// Search through all workspace files
	for _, shFile := range workspaceShFiles {
		fileContent, err := os.ReadFile(shFile)
		if err != nil {
			slog.Error("Could not read file", "file", shFile)
			continue
		}

		workspaceFileAst, err := ParseDocument(string(fileContent), shFile, false)
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

			if uri == utils.PathToURI(resolved) {
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
		var refs []RefNode

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
				if workspaceFileAst.wouldResolveToSameDefinitionAcrossFiles(refNode.Node, defNode) {
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
