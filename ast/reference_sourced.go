package ast

import (
	"fmt"
	"log/slog"
	"os"

	"mvdan.cc/sh/v3/syntax"
)

// Enhanced cross-file references with proper scoping
func (a *Ast) FindRefsinSourcedFile(
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
	includeDeclaration bool,
) map[string][]RefNode {
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return map[string][]RefNode{}
	}

	// First, find the definition using cross-file definition resolution
	targetFile, defNode := a.FindDefinitionAcrossFiles(cursorNode, env, baseDir)

	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})
	filesRefNodes := map[string][]RefNode{}

	slog.Info("SOURCED_REFS", "targetFile", targetFile, "defNode", defNode, "identifier", targetIdentifier)

	// If no definition found, fall back to simple name matching
	if defNode == nil {
		slog.Info("No definition found, falling back to name matching")

		for _, sourcedFile := range sourcedFiles {
			fileContent, err := os.ReadFile(sourcedFile)
			if err != nil {
				slog.Error("Could not read file", "file", sourcedFile)
				continue
			}
			sourcedFileAst, err := ParseDocument(string(fileContent), sourcedFile)
			if err != nil {
				slog.Error(err.Error())
				continue
			}

			refs := []RefNode{}
			for _, refNode := range sourcedFileAst.RefNodes(includeDeclaration) {
				if refNode.Name == targetIdentifier {
					refs = append(refs, refNode)
				}
			}
			filesRefNodes[sourcedFile] = refs
		}
		return filesRefNodes
	}

	slog.Info("Definition found, using proper scoping logic")

	// Definition found - find references across all files with proper scoping
	for _, sourcedFile := range sourcedFiles {
		fileContent, err := os.ReadFile(sourcedFile)
		if err != nil {
			slog.Error("Could not read file", "file", sourcedFile)
			continue
		}
		sourcedFileAst, err := ParseDocument(string(fileContent), sourcedFile)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		refs := []RefNode{}
		for _, refNode := range sourcedFileAst.RefNodes(includeDeclaration) {
			if refNode.Name != targetIdentifier {
				continue
			}

			// Use the same logic as single-file references but adapted for cross-file
			if sourcedFileAst.WouldResolveToSameDefinitionAcrossFiles(refNode.Node, defNode, targetFile, sourcedFile) {
				refs = append(refs, refNode)
			}
		}
		filesRefNodes[sourcedFile] = refs
	}

	return filesRefNodes
}

// Cross-file version of wouldResolveToSameDefinition
func (a *Ast) WouldResolveToSameDefinitionAcrossFiles(refCursorNode syntax.Node, targetDefNode *DefNode, defFile, refFile string) bool {
	// Always load the reference file's AST since we don't know which file the current AST represents
	fileContent, err := os.ReadFile(refFile)
	if err != nil {
		slog.Error("Could not read reference file", "file", refFile)
		return false
	}
	refFileAst, err := ParseDocument(string(fileContent), refFile)
	if err != nil {
		slog.Error("Could not parse reference file", "file", refFile)
		return false
	}

	// Simulate FindDefInFile logic at the reference location
	pos := refCursorNode.Pos()
	cursor := Cursor{Line: pos.Line(), Col: pos.Col()}
	refScope := refFileAst.findEnclosingFunction(cursor)

	targetIdentifier := targetDefNode.Name

	// First, look for scoped variables in the same function scope within the reference file
	if refScope != nil {
		// Find the closest local declaration that comes BEFORE the reference
		var closestLocalDef *DefNode

		for _, defNode := range refFileAst.DefNodes() {
			if defNode.Name == targetIdentifier && defNode.IsScoped && defNode.Scope == refScope {
				// Check if the scoped variable is declared BEFORE the reference position
				if defNode.isBeforeCursor(cursor) {
					// Among all local variables declared before this reference,
					// find the one that's closest (latest declaration)
					if closestLocalDef == nil || defNode.isDefinitionAfter(closestLocalDef) {
						closestLocalDef = &defNode
					}
				}
			}
		}

		// If we found a local variable declared before the reference, use it
		if closestLocalDef != nil {
			return isSameDefinitionAcrossFiles(closestLocalDef, targetDefNode, refFile, defFile)
		}
	}

	// No local variable found that's declared before the reference
	// Check if the target definition is global and would be visible

	// Case 1: Target definition is scoped (local) - only visible within same function name across files
	if targetDefNode.IsScoped {
		if refScope != nil && targetDefNode.Scope != nil {
			// Both reference and definition are in functions
			refFuncName := refScope.Name.Value
			defFuncName := targetDefNode.Scope.Name.Value

			if refFuncName == defFuncName {
				// Same function name across files - the reference can see the definition
				return true
			}
		}
		// Different function names or one is not in a function - scoped variables are not visible
		return false
	}

	// Case 2: Target definition is global
	// Check for local shadowing in the reference file
	if refScope != nil {
		// Check if there's a local variable shadowing the global one
		for _, candidateDef := range refFileAst.DefNodes() {
			if candidateDef.Name == targetIdentifier && candidateDef.IsScoped &&
				candidateDef.Scope == refScope {
				if candidateDef.isBeforeCursor(cursor) {
					return false // Local variable shadows the global definition
				}
			}
		}
	}

	// No local shadowing found - check if this is the same global definition
	// For cross-file, we need to find the corresponding global definition in the reference file
	for _, defNode := range refFileAst.DefNodes() {
		if defNode.Name == targetIdentifier && !defNode.IsScoped {
			// Found a global definition in the reference file
			// In bash, global variables are shared across sourced files
			// So any global definition of the same name refers to the same variable
			return true
		}
	}

	// If no global definition in reference file, but target is global, it's still visible
	return true
}

// Helper method to check if two definitions are the same across files
func isSameDefinitionAcrossFiles(def1 *DefNode, def2 *DefNode, file1, file2 string) bool {
	// If they're in the same file, use the regular comparison
	if file1 == file2 {
		return def1.isSameDefinition(def2)
	}

	// Cross-file comparison
	// For scoped variables, they're the same if they're in functions with the same name
	if def1.IsScoped && def2.IsScoped {
		if def1.Scope != nil && def2.Scope != nil {
			return def1.Scope.Name.Value == def2.Scope.Name.Value && def1.Name == def2.Name
		}
		return false
	}

	// For global variables, they're the same if they have the same name
	// (bash global variables are shared across sourced files)
	if !def1.IsScoped && !def2.IsScoped {
		return def1.Name == def2.Name
	}

	// One scoped, one global - they're different
	return false
}

// Debug method for cross-file reference resolution
func (a *Ast) DebugCrossFileReferenceResolution(
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
) {
	targetIdentifier := ExtractIdentifier(cursorNode)
	fmt.Printf("=== Cross-File Reference Resolution for '%s' ===\n", targetIdentifier)

	targetFile, defNode := a.FindDefinitionAcrossFiles(cursorNode, env, baseDir)

	if defNode == nil {
		fmt.Printf("No definition found\n")
		return
	}

	fmt.Printf("Target definition: %s at line %d:%d in %s\n",
		defNode.Name, defNode.StartLine, defNode.StartChar, targetFile)
	if defNode.IsScoped {
		fmt.Printf("  (scoped in function: %s)\n", defNode.Scope.Name.Value)
	} else {
		fmt.Printf("  (global)\n")
	}

	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})

	for _, sourcedFile := range sourcedFiles {
		fmt.Printf("\n--- File: %s ---\n", sourcedFile)

		fileContent, err := os.ReadFile(sourcedFile)
		if err != nil {
			fmt.Printf("Error reading file: %s\n", err)
			continue
		}
		sourcedFileAst, err := ParseDocument(string(fileContent), sourcedFile)
		if err != nil {
			fmt.Printf("Error parsing file: %s\n", err)
			continue
		}

		for _, refNode := range sourcedFileAst.RefNodes(true) {
			if refNode.Name != targetIdentifier {
				continue
			}

			fmt.Printf("Line %d:%d - '%s'", refNode.StartLine, refNode.StartChar, refNode.Name)

			if sourcedFileAst.WouldResolveToSameDefinitionAcrossFiles(refNode.Node, defNode, targetFile, sourcedFile) {
				fmt.Printf(" -> MATCHES target definition\n")
			} else {
				fmt.Printf(" -> different definition\n")
			}
		}
	}
}
