package ast

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Wraps a node that can be part of a definition or reference.
type DefNode struct {
	Node      syntax.Node
	Name      string
	Scope     *syntax.FuncDecl
	StartLine uint
	StartChar uint
	EndLine   uint
	EndChar   uint
}

func (a *Ast) DefNodes() []DefNode {
	defNodes := []DefNode{}

	syntax.Walk(a.File, func(node syntax.Node) bool {
		var name string
		var startLine, startChar, endLine, endChar uint

		switch n := node.(type) {
		// Variable Assignment
		case *syntax.Assign:
			if n.Name != nil {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()
			}
		// Function Definition
		case *syntax.FuncDecl:
			if n.Name != nil {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()
			}
		// Iteration variable in for/select loops
		case *syntax.ForClause:
			switch loop := n.Loop.(type) {
			case *syntax.WordIter:
				if loop.Name != nil {
					name = loop.Name.Value
					startLine, startChar = loop.Name.Pos().Line(), loop.Name.Pos().Col()
					endLine, endChar = loop.Name.End().Line(), loop.Name.End().Col()
				}
			case *syntax.CStyleLoop:
				if loop.Init != nil {
					a, ok := loop.Init.(*syntax.BinaryArithm)
					if !ok {
						return true
					}
					if a.Op == syntax.Assgn {
						word, ok := a.X.(*syntax.Word)
						if !ok {
							return true
						}
						for _, wp := range word.Parts {
							switch p := wp.(type) {
							case *syntax.Lit:
								name = p.Value
								startLine, startChar = p.Pos().Line(), p.Pos().Col()
								endLine, endChar = p.End().Line(), p.End().Col()
							}
						}
					}
				}
			}
		// Variable "assinment" in read statements
		case *syntax.CallExpr:
			if len(n.Args) == 0 {
				return true
			}
			cmdName := ExtractIdentifier(n.Args[0])
			if cmdName != "read" {
				return true
			}
			for _, arg := range n.Args {
				for _, wp := range arg.Parts {
					switch p := wp.(type) {
					case *syntax.Lit:
						name = p.Value
						startLine, startChar = p.Pos().Line(), p.Pos().Col()
						endLine, endChar = p.End().Line(), p.End().Col()
					}
				}
			}
		}

		if name != "" {
			defNodes = append(defNodes, DefNode{
				Node:      node,
				Name:      name,
				StartLine: startLine,
				StartChar: startChar,
				EndLine:   endLine,
				EndChar:   endChar,
			})
		}

		return true
	})

	return defNodes
}

func (a *Ast) FindDefInFile(cursorNode syntax.Node) *DefNode {
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	pos := cursorNode.Pos()
	cursor := Cursor{Line: pos.Line(), Col: pos.Col()}
	cursorScope := a.FindEnclosingFunction(cursor)
	if cursorScope != nil {
		for _, scopedDefNode := range a.ScopedDefNodes() {
			if scopedDefNode.Name == targetIdentifier && scopedDefNode.Scope == cursorScope {
				return &scopedDefNode
			}
		}
	}


	for _, defNode := range a.DefNodes() {
		if defNode.Name == targetIdentifier {
			return &defNode
		}

	}

	return nil
}

// Find a definition in a sourced file.
func (a *Ast) FindDefInSourcedFile(
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
) (string, *DefNode) {
	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})

	var definition *DefNode
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
		definition = sourcedFileAst.FindDefInFile(cursorNode)
		if definition != nil {
			return sourcedFile, definition
		}
	}

	return "", nil
}

// Find a sourced file itself (cursor over filepath).
func (a *Ast) FindSourcedFile(
	cursor Cursor,
	env map[string]string,
	baseDir string,
) string {
	found := ""

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
		start, end := argNode.Pos(), argNode.End()
		if isCursorInNode(cursor, start, end) {
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

func (d *DefNode) ToHoverString(documentText string, documentName string) string {
	switch n := d.Node.(type) {
	case *syntax.FuncDecl:
		lines := strings.Split(documentText, "\n")
		functionSnippet := strings.Join(lines[n.Pos().Line()-1:n.End().Line()], "\n")

		defLocation := fmt.Sprintf("defined at `%s` line **%d**", documentName, n.Pos().Line())
		if documentName == "" {
			defLocation = fmt.Sprintf("defined at line **%d**", n.Pos().Line())
		}

		return fmt.Sprintf("```sh\n%s\n```\n\n(%s)", functionSnippet, defLocation)

	case *syntax.Assign:
		if documentName == "" {
			return fmt.Sprintf("defined at line **%d**", n.Pos().Line())
		}

		return fmt.Sprintf("defined at `%s` line **%d**", documentName, n.Pos().Line())
	}

	return ""
}
