package ast

import (
	"log/slog"
	"os"
)

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

	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})

	// Search for globals in reverse source order (last sourced file first)
	for i := len(sourcedFiles) - 1; i >= 0; i-- {
		file, def := findGlobalInSourcedFile(targetIdentifier, sourcedFiles[i])
		if def != nil {
			return file, def
		}
	}

	return "", nil
}

func findGlobalInSourcedFile(targetIdentifier, sourcedFile string) (string, *DefNode) {
	fileContent, err := os.ReadFile(sourcedFile)
	if err != nil {
		slog.Error("Could not read file", "file", sourcedFile)
		return "", nil
	}

	sourcedAst, err := ParseDocument(string(fileContent), sourcedFile)
	if err != nil {
		slog.Error("Could not parse file", "file", sourcedFile)
		return "", nil
	}

	for _, defNode := range sourcedAst.DefNodes() {
		if defNode.Name == targetIdentifier && !defNode.IsScoped {
			return sourcedFile, &defNode
		}
	}

	return "", nil
}
