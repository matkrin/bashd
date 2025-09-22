package ast

import "mvdan.cc/sh/v3/syntax"

// Functions are the only contructs with scope
func (a *Ast) FindEnclosingFunction(cursor Cursor) *syntax.FuncDecl {
	var enclosingFunc *syntax.FuncDecl

	syntax.Walk(a.File, func(node syntax.Node) bool {
		fn, ok := node.(*syntax.FuncDecl)
		if !ok {
			return true
		}
		start, end := fn.Pos(), fn.End()
		if isCursorInNode(cursor, start, end) {
			enclosingFunc = fn
		}
		return true
	})

	return enclosingFunc
}

// Only `local`, `declare` and `typeset` statements are scoped bound and only in functions
func (a *Ast) ScopedDefNodes() []DefNode {
	scopedDefNodes := []DefNode{}

	syntax.Walk(a.File, func(node syntax.Node) bool {
		var name string
		var startLine, startChar, endLine, endChar uint

		funcDecl, ok := node.(*syntax.FuncDecl)
		if !ok {
			return true
		}
		if funcDecl.Body.Cmd == nil {
			return true
		}
		block, ok := funcDecl.Body.Cmd.(*syntax.Block)
		if !ok {
			return true
		}
		for _, stmt := range block.Stmts {
			declClause, ok := stmt.Cmd.(*syntax.DeclClause)
			if !ok {
				continue
			}
			cmd := declClause.Variant.Value
			if cmd != "local" && cmd != "declare" && cmd != "typeset" {
				continue
			}
			for _, arg := range declClause.Args {
				name = arg.Name.Value
				startLine, startChar = arg.Name.ValuePos.Line(), arg.Name.ValuePos.Col()
				endLine, endChar = arg.Name.ValueEnd.Line(), arg.Name.ValueEnd.Col()
				scopedDefNodes = append(scopedDefNodes, DefNode{
					Node:      declClause,
					Name:      name,
					Scope:     funcDecl,
					StartLine: startLine,
					StartChar: startChar,
					EndLine:   endLine,
					EndChar:   endChar,
				})
			}

		}

		return true
	})

	return scopedDefNodes
}
