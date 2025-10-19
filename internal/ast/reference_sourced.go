package ast

import (
	"log/slog"
	"os"

	"mvdan.cc/sh/v3/syntax"
)

// Cross-file reference finding
func (a *Ast) FindRefsinSourcedFile(
	cursor Cursor,
	env map[string]string,
	baseDir string,
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

	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})
	filesRefNodes := map[string][]RefNode{}

	for _, sourcedFile := range sourcedFiles {
		fileContent, err := os.ReadFile(sourcedFile)
		if err != nil {
			slog.Error("Could not read file", "file", sourcedFile)
			continue
		}
		sourcedFileAst, err := ParseDocument(string(fileContent), sourcedFile, false)
		if err != nil {
			slog.Error("Could not parse file", "file", sourcedFile)
			continue
		}

		refs := []RefNode{}
		for _, refNode := range sourcedFileAst.RefNodes(includeDeclaration) {
			if refNode.Name != targetIdentifier {
				continue
			}

			if sourcedFileAst.wouldResolveToSameDefinitionAcrossFiles(refNode.Node, defNode) {
				refs = append(refs, refNode)
			}
		}
		if len(refs) > 0 {
			filesRefNodes[sourcedFile] = refs
		}
	}

	return filesRefNodes
}

func (a *Ast) wouldResolveToSameDefinitionAcrossFiles(refCursorNode syntax.Node, targetDefNode *DefNode) bool {
	pos := refCursorNode.Pos()
	cursor := Cursor{Line: pos.Line(), Col: pos.Col()}
	refScope := a.findEnclosingFunction(cursor)

	targetIdentifier := targetDefNode.Name

	// First, look for scoped variables in the same function scope within the reference file
	if refScope != nil {
		// Find the closest local declaration that comes BEFORE the reference
		var closestLocalDef *DefNode

		for _, defNode := range a.DefNodes() {
			if defNode.Name == targetIdentifier && defNode.IsScoped && defNode.Scope == refScope {
				if defNode.isBeforeCursor(cursor) {
					if closestLocalDef == nil || defNode.isDefinitionAfter(closestLocalDef) {
						closestLocalDef = &defNode
					}
				}
			}
		}

		// If we found a local variable declared before the reference, use it
		if closestLocalDef != nil {
			return closestLocalDef.isSameDefinition(targetDefNode)
		}
	}

	if targetDefNode.IsScoped {
		return false
	}

	// If no global definition in reference file, but target is global, it's still visible
	return true
}
