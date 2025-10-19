package ast

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/matkrin/bashd/internal/utils"
)

func (a *Ast) FindRefsInWorkspaceFiles(
	uri string,
	workspaceShFiles []string,
	baseDir string,
	cursor Cursor,
	env map[string]string,
	includeDeclaration bool,
) map[string][]RefNode {
	cursorNode := a.FindNodeUnderCursor(cursor)
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return map[string][]RefNode{}
	}

	_, defNode := a.FindDefinitionAcrossFiles(cursor, env, baseDir)
	if defNode == nil {
		return map[string][]RefNode{}
	}

	referenceNodes := map[string][]RefNode{}

	for _, workspaceShFile := range workspaceShFiles {
		fileContent, err := os.ReadFile(workspaceShFile)
		if err != nil {
			slog.Error("Could not read file", "file", workspaceShFile)
			continue
		}

		workspaceFileAst, err := ParseDocument(string(fileContent), workspaceShFile, false)
		if err != nil {
			slog.Error("Could not parse file", "file", workspaceShFile)
			continue
		}

		// Check if this file sources the target file
		shouldIncludeFile := false
		for _, sourceStatement := range workspaceFileAst.FindSourceStatments(env) {
			baseDir := filepath.Dir(workspaceShFile)
			path := sourceStatement.SourcedFile
			resolved := path
			if !filepath.IsAbs(path) {
				resolved = filepath.Join(baseDir, path)
			}
			resolved = filepath.Clean(resolved)

			if uri == utils.PathToURI(resolved) {
				shouldIncludeFile = true
				break
			}
		}

		if !shouldIncludeFile {
			// Check if the workspace file is in the same sourcing chain
			sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})
			if slices.Contains(sourcedFiles, workspaceShFile) {
				shouldIncludeFile = true
			}
		}

		if !shouldIncludeFile {
			continue
		}

		var refs []RefNode
		for _, refNode := range workspaceFileAst.RefNodes(includeDeclaration) {
			if refNode.Name != targetIdentifier {
				continue
			}

			if workspaceFileAst.wouldResolveToSameDefinitionAcrossFiles(refNode.Node, defNode) {
				refs = append(refs, refNode)
			}
		}

		if len(refs) > 0 {
			referenceNodes[workspaceShFile] = refs
		}
	}

	return referenceNodes
}
