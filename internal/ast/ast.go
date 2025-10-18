package ast

import (
	"os"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

type Cursor struct {
	Line uint
	Col  uint
}

// Otherwise I will mess that up for sure. In the LSP 0-based, in the parser 1-based
func NewCursor(lspLine, lspCol uint) Cursor {
	return Cursor{Line: lspLine + 1, Col: lspCol + 1}
}

func (c *Cursor) isCursorInNode(node syntax.Node) bool {
	startLine := node.Pos().Line()
	startCol := node.Pos().Col()
	endLine := node.End().Line()
	endCol := node.End().Col()

	// Compare lines first
	if c.Line < startLine || c.Line > endLine {
		return false
	}

	if startLine == endLine {
		// Node is on a single line
		return c.Line == startLine &&
			c.Col >= startCol && c.Col <= endCol
	}

	// Multi-line node
	switch c.Line {
	case startLine:
		// On first line of node: col must be >= start.Col()
		return c.Col >= startCol
	case endLine:
		// On last line of node: col must be <= end.Col()
		return c.Col <= endCol
	default:
		// Any line in between start and end line is inside
		return true
	}
}

type Ast struct {
	File *syntax.File
}

func ParseDocument(documentText, documentName string, fallible bool) (*Ast, error) {
	reader := strings.NewReader(documentText)
	var parser *syntax.Parser
	if fallible {
		parser = syntax.NewParser(syntax.KeepComments(true), syntax.RecoverErrors(9999))
	} else {
		parser = syntax.NewParser(syntax.KeepComments(true))
	}
	file, err := parser.Parse(reader, documentName)
	if err != nil {
		return nil, err
	}
	return &Ast{File: file}, nil
}

func (a *Ast) FindNodeUnderCursor(cursor Cursor) syntax.Node {
	var found syntax.Node

	syntax.Walk(a.File, func(node syntax.Node) bool {
		if node == nil {
			return true
		}
		if cursor.isCursorInNode(node) {
			found = node
			// Continue walking to find deepest node containing cursor
			return true
		}
		return true
	})

	return found
}

func ExtractIdentifier(node syntax.Node) string {
	switch n := node.(type) {
	case *syntax.Lit:
		return n.Value
	case *syntax.ParamExp:
		if n.Param != nil {
			return n.Param.Value
		}
	case *syntax.Word:
		if len(n.Parts) == 1 {
			switch p := n.Parts[0].(type) {
			case *syntax.Lit:
				return p.Value
			}
		}
	case *syntax.Assign:
		if n.Name != nil {
			return n.Name.Value
		}
	case *syntax.FuncDecl:
		if n.Name != nil {
			return n.Name.Value
		}
	}
	return ""
}

func extractAndExpandWord(word *syntax.Word, env map[string]string) string {
	var b strings.Builder
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(p.Value)

		case *syntax.ParamExp:
			val := env[p.Param.Value]
			b.WriteString(val)

		case *syntax.SglQuoted:
			b.WriteString(p.Value)

		case *syntax.DblQuoted:
			for _, qpart := range p.Parts {
				switch qp := qpart.(type) {
				case *syntax.Lit:
					b.WriteString(qp.Value)
				case *syntax.ParamExp:
					val := env[qp.Param.Value]
					b.WriteString(val)
				}
			}
		}
	}
	// Expand things like $HOME and ${VAR}
	return os.Expand(b.String(), func(key string) string {
		return env[key]
	})
}

// Functions are the only contructs with scope
func (a *Ast) findEnclosingFunction(cursor Cursor) *syntax.FuncDecl {
	var enclosingFunc *syntax.FuncDecl

	syntax.Walk(a.File, func(node syntax.Node) bool {
		fn, ok := node.(*syntax.FuncDecl)
		if !ok {
			return true
		}
		if cursor.isCursorInNode(fn) {
			enclosingFunc = fn
		}
		return true
	})

	return enclosingFunc
}
