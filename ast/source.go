package ast

import (
	"os"
	"path/filepath"

	"mvdan.cc/sh/v3/syntax"
)

type SourceStatement struct {
	SourcedFile string
	StartLine   uint
	StartChar   uint
	EndLine     uint
	EndChar     uint
}

// Find `source` statements in AST
func (a *Ast) FindSourceStatments(env map[string]string) []SourceStatement {
	sourcedStatements := []SourceStatement{}
	syntax.Walk(a.File, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}

		cmdName := extractAndExpandWord(call.Args[0], env)
		if cmdName != "source" && cmdName != "." {
			return true
		}

		path := extractAndExpandWord(call.Args[1], env)
		if path == "" {
			return true
		}

		sourcedStatements = append(sourcedStatements, SourceStatement{
			SourcedFile: path,
			StartLine:   node.Pos().Line() - 1,
			StartChar:   node.Pos().Col() - 1,
			EndLine:     node.End().Line() - 1,
			EndChar:     node.End().Col() - 1,
		})
		return true
	})
	return sourcedStatements
}

// Find sourced files recursively and return their filenames
func (a *Ast) FindAllSourcedFiles(
	env map[string]string,
	baseDir string,
	visited map[string]bool,
) []string {
	var sourcedFiles []string
	for _, sourcedFile := range a.FindSourceStatments(env) {
		path := sourcedFile.SourcedFile
		resolved := path
		if !filepath.IsAbs(path) {
			resolved = filepath.Join(baseDir, path)
		}
		resolved = filepath.Clean(resolved)

		if visited[resolved] {
			continue
		}
		visited[resolved] = true

		sourcedFiles = append(sourcedFiles, resolved)

		// Recurse
		if content, err := os.ReadFile(resolved); err == nil {
			if parsed, err := ParseDocument(string(content), ""); err == nil {
				subFiles := parsed.FindAllSourcedFiles(env, filepath.Dir(resolved), visited)
				sourcedFiles = append(sourcedFiles, subFiles...)
			}
		}
	}
	return sourcedFiles
}

// Find a sourced file itself (cursor over filepath).
func (a *Ast) FindSourcedFile(
	cursor Cursor,
	env map[string]string,
	baseDir string,
) string {
	var found string

	syntax.Walk(a.File, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}

		cmdName := extractAndExpandWord(call.Args[0], env)
		if cmdName != "source" && cmdName != "." {
			return true
		}

		argNode := call.Args[1]
		if cursor.isCursorInNode(argNode) {
			path := extractAndExpandWord(argNode, env)

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}
			path = filepath.Clean(path)
			found = path

			return false // stop walking
		}
		return true
	})

	return found
}
