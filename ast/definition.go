package ast

import (
	"os"
	"path/filepath"

	"github.com/matkrin/bashd/logger"
	"mvdan.cc/sh/v3/syntax"
)

// Wraps a node that can be part of a definition or reference.
type DefNode struct {
	Node  syntax.Node
	Start syntax.Pos // Starting position of the node
	End   syntax.Pos // End position of the node
}

func FindDefinition(cursorNode syntax.Node, file *syntax.File) *DefNode {
	targetIdentifier := extractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	var result *DefNode

	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil || result != nil {
			return false
		}

		var name string
		var pos, end syntax.Pos

		switch n := node.(type) {
		case *syntax.ParamExp:
			if n.Param != nil {
				name = n.Param.Value
				pos, end = n.Pos(), n.End()
			}
		case *syntax.Assign:
			if n.Name != nil {
				name = n.Name.Value
				pos, end = n.Name.Pos(), n.Name.End()
			}
		case *syntax.FuncDecl:
			if n.Name != nil {
				name = n.Name.Value
				pos, end = n.Name.Pos(), n.Name.End()
			}
		case *syntax.CallExpr:

		}

		if name == targetIdentifier {
			result = &DefNode{
				Node:  node,
				Start: pos,
				End:   end,
			}
			return false
		}

		return true
	})

	return result
}

func FindDefInSourcedFile(
	filename string,
	fileAst *syntax.File,
	cursorNode syntax.Node,
	env map[string]string,
) (string, *DefNode) {
	baseDir := filepath.Dir(filename)
	sourcedFiles := FindAllSourcedFiles(fileAst, env, baseDir, map[string]bool{})

	var definition *DefNode
	for _, sourcedFile := range sourcedFiles {
		fileContent, err := os.ReadFile(sourcedFile)
		if err != nil {
			logger.Error("Could not read file: %s", sourcedFile)
			continue
		}
		sourcedFileAst, err := ParseDocument(string(fileContent), sourcedFile)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		definition = FindDefinition(cursorNode, sourcedFileAst)
		if definition != nil {
			return sourcedFile, definition
		}
	}

	return "", nil
}

func FindSourcedFile(file *syntax.File, cursor Cursor, env map[string]string, baseDir string) string {
	found := ""

	syntax.Walk(file, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}

		cmdName := extractWord(call.Args[0], env)
		if cmdName != "source" && cmdName != "." {
			return true
		}

		argNode := call.Args[1]
		start, end := argNode.Pos(), argNode.End()
		if isCursorInNode(cursor, start, end) {
			path := extractWord(argNode, env)

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
