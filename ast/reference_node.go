package ast

import (
	"log/slog"
	"os"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

// Wraps a node that can be part of a reference.
type RefNode struct {
	Node      *syntax.Node
	Name      string
	StartLine uint
	StartChar uint
	EndLine   uint
	EndChar   uint
}

func (r *RefNode) ToLspLocation(uri string) lsp.Location {
	return lsp.Location{
		URI: uri,
		Range: lsp.NewRange(
			r.StartLine-1,
			r.StartChar-1,
			r.EndLine-1,
			r.EndChar-1,
		),
	}
}

func (r *RefNode) ToLspTextEdit(newText string) lsp.TextEdit {
	return lsp.TextEdit{
		Range: lsp.NewRange(
			r.StartLine-1,
			r.StartChar-1,
			r.EndLine-1,
			r.EndChar-1,
		),
		NewText: newText,
	}
}

func (a *Ast) refNodes(includeDeclaration bool) []RefNode {
	refNodes := []RefNode{}

	syntax.Walk(a.File, func(node syntax.Node) bool {
		var name string
		var startLine, startChar, endLine, endChar uint

		switch n := node.(type) {
		// Variable usage
		case *syntax.ParamExp:
			if n.Param != nil {
				name = n.Param.Value
				startLine, startChar = n.Param.Pos().Line(), n.Param.Pos().Col()
				endLine, endChar = n.Param.End().Line(), n.Param.End().Col()
			}
		// Function usage
		case *syntax.Word:
			if len(n.Parts) == 1 {
				switch p := n.Parts[0].(type) {
				case *syntax.Lit:
					name = p.Value
					startLine, startChar = p.Pos().Line(), p.Pos().Col()
					endLine, endChar = p.End().Line(), p.End().Col()
				}
			}
		// Funtion declaration
		case *syntax.FuncDecl:
			if n.Name != nil && includeDeclaration {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()
			}
		// Variable assignement
		case *syntax.Assign:
			if n.Name != nil && includeDeclaration {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()
			}
		// Iteration variable in for/select loops
		case *syntax.ForClause:
			loop, ok := n.Loop.(*syntax.WordIter)
			if !ok {
				return true
			}
			if loop.Name != nil {
				name = loop.Name.Value
				startLine, startChar = loop.Name.Pos().Line(), loop.Name.Pos().Col()
				endLine, endChar = loop.Name.End().Line(), loop.Name.End().Col()
			}
		}

		if name != "" {
			refNodes = append(refNodes, RefNode{
				Node:      &node,
				Name:      name,
				StartLine: startLine,
				StartChar: startChar,
				EndLine:   endLine,
				EndChar:   endChar,
			})
		}

		return true
	})

	return refNodes
}

func (a *Ast) FindRefsInFile(cursorNode syntax.Node, includeDeclaration bool) []RefNode {
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	references := []RefNode{}
	for _, node := range a.refNodes(includeDeclaration) {
		if node.Name == targetIdentifier {
			references = append(references, node)
		}

	}

	return references
}

func (a *Ast) FindRefsinSourcedFile(
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
	includeDeclaration bool,
) map[string][]RefNode {
	sourcedFiles := a.FindAllSourcedFiles(env, baseDir, map[string]bool{})

	filesRefNodes := map[string][]RefNode{}
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
		references := sourcedFileAst.FindRefsInFile(cursorNode, includeDeclaration)
		filesRefNodes[sourcedFile] = references
	}

	return filesRefNodes
}
