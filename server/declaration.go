package server

import (
	"log/slog"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

// "declare", "local", "export", "readonly", "typeset", or "nameref".
func handleDeclaration(request *lsp.DeclarationRequest, state *State) *lsp.DeclarationResponse {
	uri := request.Params.TextDocument.URI
	cursor := newCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		slog.Error(err.Error())
	}
	cursorNode := findNodeUnderCursor(fileAst, cursor)
	declaration := findDeclInFile(cursorNode, fileAst)
	if declaration == nil {
		return nil
	}

	return lsp.NewDeclarationResponse(
		request.ID,
		uri,
		declaration.StartLine-1,
		declaration.StartChar-1,
		declaration.EndLine-1,
		declaration.EndChar-1,
	)
}

type DeclNode struct {
	Node      syntax.Node
	Name      string
	StartLine uint
	StartChar uint
	EndLine   uint
	EndChar   uint
}

func findDeclInFile(cursorNode syntax.Node, fileAst *syntax.File) *DeclNode {
	targetIdentifier := extractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	for _, declNode := range declNodes(fileAst) {
		if declNode.Name == targetIdentifier {
			return &declNode
		}

	}

	return nil
}

func declNodes(file *syntax.File) []DeclNode {

	declNodes := []DeclNode{}

	syntax.Walk(file, func(node syntax.Node) bool {
		var name string
		var startLine, startChar, endLine, endChar uint

		switch n := node.(type) {
		case *syntax.DeclClause:
			if n.Args != nil {
				for _, arg := range n.Args {
					if arg.Name != nil {
						name = arg.Name.Value
						startLine, startChar = arg.Name.ValuePos.Line(), arg.Name.ValuePos.Col()
						endLine, endChar = arg.Name.End().Line(), arg.Name.End().Col()
					}
				}
			}

			if name != "" {
				declNodes = append(declNodes, DeclNode{
					Node:      node,
					Name:      name,
					StartLine: startLine,
					StartChar: startChar,
					EndLine:   endLine,
					EndChar:   endChar,
				})
			}
		}
		return true
	})

	return declNodes
}
