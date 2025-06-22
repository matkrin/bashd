package server

import (
	"os"
	"path/filepath"

	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleDefinition(request *lsp.DefinitionRequest, state *State) *lsp.DefinitionResponse {
	uri := request.Params.TextDocument.URI
	cursor := newCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		logger.Error(err.Error())
	}
	cursorNode := findNodeUnderCursor(fileAst, cursor)
	definition := findDefInFile(cursorNode, fileAst)

	if definition == nil {
		// Check for the definition in a sourced file
		filename, err := uriToPath(uri)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		baseDir := filepath.Dir(filename)
		sourcedFile := ""
		sourcedFile, definition = findDefInSourcedFile(
			fileAst,
			cursorNode,
			state.EnvVars,
			baseDir,
		)

		if definition != nil {
			uri = pathToURI(sourcedFile)
		}
	}

	if definition == nil {
		// Check if the cursor is over a filename in a source statement
		filename, err := uriToPath(uri)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		baseDir := filepath.Dir(filename)
		sourcePath := findSourcedFile(fileAst, cursor, state.EnvVars, baseDir)
		// Check if file exists
		if _, err := os.Stat(sourcePath); err != nil {
			return nil
		}
		definition = &DefNode{
			Node:  cursorNode,
			Start: syntax.NewPos(0, 1, 1),
			End:   syntax.NewPos(0, 1, 1),
		}
		uri = pathToURI(sourcePath)
	}

	if definition == nil {
		return nil
	}

	response := lsp.NewDefinitionResponse(
		request.ID,
		uri,
		int(definition.Start.Line())-1,
		int(definition.Start.Col())-1,
		int(definition.End.Line())-1,
		int(definition.End.Col())-1,
	)
	return &response
}

// Wraps a node that can be part of a definition or reference.
type DefNode struct {
	Node  syntax.Node
	Name  string
	Start syntax.Pos // Starting position of the node
	End   syntax.Pos // End position of the node
}

func defNodes(file *syntax.File) []DefNode {
	defNodes := []DefNode{}

	syntax.Walk(file, func(node syntax.Node) bool {
		var name string
		var pos, end syntax.Pos

		switch n := node.(type) {
		// case *syntax.ParamExp:
		// 	if n.Param != nil {
		// 		name = n.Param.Value
		// 		pos, end = n.Pos(), n.End()
		// 	}
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
		}

		if name != "" {
			defNodes = append(defNodes, DefNode{
				Node:  node,
				Name:  name,
				Start: pos,
				End:   end,
			})
		}

		return true
	})

	return defNodes
}

func findDefInFile(cursorNode syntax.Node, file *syntax.File) *DefNode {
	targetIdentifier := extractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	for _, node := range defNodes(file) {
		if node.Name == targetIdentifier {
			return &node
		}

	}

	return nil
}

// Find a definition in a sourced file.
func findDefInSourcedFile(
	fileAst *syntax.File,
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
) (string, *DefNode) {
	sourcedFiles := findAllSourcedFiles(fileAst, env, baseDir, map[string]bool{})

	var definition *DefNode
	for _, sourcedFile := range sourcedFiles {
		fileContent, err := os.ReadFile(sourcedFile)
		if err != nil {
			logger.Error("Could not read file: %s", sourcedFile)
			continue
		}
		sourcedFileAst, err := parseDocument(string(fileContent), sourcedFile)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		definition = findDefInFile(cursorNode, sourcedFileAst)
		if definition != nil {
			return sourcedFile, definition
		}
	}

	return "", nil
}

// Find a sourced file itself (cursor over filepath).
func findSourcedFile(
	file *syntax.File,
	cursor Cursor,
	env map[string]string,
	baseDir string,
) string {
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
