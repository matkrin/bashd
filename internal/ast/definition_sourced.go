package ast

import (
	"log/slog"
	"os"

	"mvdan.cc/sh/v3/syntax"
)

// Find a definition in a sourced file with proper scoping
func (a *Ast) FindDefInSourcedFile(
	cursor Cursor,
	env map[string]string,
	baseDir string,
) (string, *DefNode) {
	cursorNode := a.FindNodeUnderCursor(cursor)
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return "", nil
	}

	cursorScope := a.findEnclosingFunction(cursor)

	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})

	// Step 1: If cursor is in a function, look for local variables in that same function across sourced files
	if cursorScope != nil {
		cursorFuncName := cursorScope.Name.Value

		for _, sourcedFile := range sourcedFiles {
			if file, def := findLocalVarInSourcedFile(targetIdentifier, cursorFuncName, sourcedFile); def != nil {
				return file, def
			}
		}
	}

	// Step 2: Look for function definitions in sourced files (functions are always global)
	for _, sourcedFile := range sourcedFiles {
		if file, def := findFunctionInSourcedFile(targetIdentifier, sourcedFile); def != nil {
			return file, def
		}
	}

	// Step 3: Look for global variables in reverse source order (last sourced wins)
	for i := len(sourcedFiles) - 1; i >= 0; i-- {
		if file, def := findGlobalVarInSourcedFile(targetIdentifier, sourcedFiles[i]); def != nil {
			return file, def
		}
	}

	return "", nil
}

// Updated unified definition search that combines current file + sourced files
func (a *Ast) FindDefinitionAcrossFiles(
	cursor Cursor,
	env map[string]string,
	baseDir string,
) (string, *DefNode) {
	if def := a.FindDefInFile(cursor); def != nil {
		return "", def
	}

	return a.FindDefInSourcedFile(cursor, env, baseDir)
}

// Find local variable in a sourced file within the same function name
func findLocalVarInSourcedFile(targetIdentifier, cursorFuncName, sourcedFile string) (string, *DefNode) {
	fileContent, err := os.ReadFile(sourcedFile)
	if err != nil {
		slog.Error("Could not read file", "file", sourcedFile)
		return "", nil
	}

	sourcedAst, err := ParseDocument(string(fileContent), sourcedFile)
	if err != nil {
		slog.Error(err.Error())
		return "", nil
	}

	// Find function with same name in sourced file
	targetFunc := sourcedAst.findFunctionByName(cursorFuncName)
	if targetFunc == nil {
		return "", nil
	}

	// Look for scoped variables in that function
	// NOTE: For cross-file shadowing, we typically consider ALL local variables in the sourced function
	// as being declared "before" the current cursor, since the sourced file is processed first
	for _, defNode := range sourcedAst.DefNodes() {
		if defNode.Name == targetIdentifier && defNode.IsScoped && defNode.Scope == targetFunc {
			return sourcedFile, &defNode
		}
	}

	return "", nil
}

// Find global variable in sourced file
func findGlobalVarInSourcedFile(targetIdentifier, sourcedFile string) (string, *DefNode) {
	fileContent, err := os.ReadFile(sourcedFile)
	if err != nil {
		slog.Error("Could not read file", "file", sourcedFile)
		return "", nil
	}

	sourcedAst, err := ParseDocument(string(fileContent), sourcedFile)
	if err != nil {
		slog.Error(err.Error())
		return "", nil
	}

	// Look for global variables
	for _, defNode := range sourcedAst.DefNodes() {
		if defNode.Name == targetIdentifier && !defNode.IsScoped {
			// Make sure it's not a function (we handle those separately)
			if _, ok := defNode.Node.(*syntax.FuncDecl); !ok {
				return sourcedFile, &defNode
			}
		}
	}

	return "", nil
}

// Find function definition in sourced file
func findFunctionInSourcedFile(targetIdentifier, sourcedFile string) (string, *DefNode) {
	fileContent, err := os.ReadFile(sourcedFile)
	if err != nil {
		slog.Error("Could not read file", "file", sourcedFile)
		return "", nil
	}

	sourcedAst, err := ParseDocument(string(fileContent), sourcedFile)
	if err != nil {
		slog.Error(err.Error())
		return "", nil
	}

	for _, defNode := range sourcedAst.DefNodes() {
		if defNode.Name == targetIdentifier {
			if _, ok := defNode.Node.(*syntax.FuncDecl); ok {
				return sourcedFile, &defNode
			}
		}
	}

	return "", nil
}

// Helper function to find function by name in AST
func (a *Ast) findFunctionByName(funcName string) *syntax.FuncDecl {
	var targetFunc *syntax.FuncDecl
	syntax.Walk(a.File, func(node syntax.Node) bool {
		if fn, ok := node.(*syntax.FuncDecl); ok && fn.Name != nil && fn.Name.Value == funcName {
			targetFunc = fn
			return false // stop walking
		}
		return true
	})
	return targetFunc
}

