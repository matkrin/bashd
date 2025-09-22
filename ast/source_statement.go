package ast

import "mvdan.cc/sh/v3/syntax"

type SourceStatement struct {
	SourcedFile string
	StartLine   uint
	StartChar   uint
	EndLine     uint
	EndChar     uint
}

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

