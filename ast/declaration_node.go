package ast

import (
	"log/slog"
	"os"

	"mvdan.cc/sh/v3/syntax"
)

type DeclNode struct {
	Node      syntax.Node
	Name      string
	StartLine uint
	StartChar uint
	EndLine   uint
	EndChar   uint
}

func (a *Ast) FindDeclInFile(cursorNode syntax.Node) *DeclNode {
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	for _, declNode := range DeclNodes(a.File) {
		if declNode.Name == targetIdentifier {
			return &declNode
		}

	}

	return nil
}

func DeclNodes(file *syntax.File) []DeclNode {

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

func (a *Ast) FindDeclInSourcedFile(
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
) (string, *DeclNode) {
	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})

	var declaration *DeclNode
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
		declaration = sourcedFileAst.FindDeclInFile(cursorNode)
		if declaration != nil {
			return sourcedFile, declaration
		}
	}

	return "", nil
}
