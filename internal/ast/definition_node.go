package ast

import (
	"fmt"
	"log/slog"

	"mvdan.cc/sh/v3/syntax"
)

// Wrapper for nodes identified as definition
type DefNode struct {
	Node      syntax.Node
	Name      string
	Scope     *syntax.FuncDecl  // nil for global scope
	IsScoped  bool             // true for local/declare/typeset variables
	StartLine uint
	StartChar uint
	EndLine   uint
	EndChar   uint
}

// Helper function to check if a definition appears before the cursor position
// Definition is before cursor if:
//   1. Definition line <= cursor line, OR
//   2. Same line but definition column < cursor column
func (d *DefNode) isBeforeCursor(cursor Cursor) bool {
	if d.StartLine <= cursor.Line {
		return true
	}
	if d.StartLine == cursor.Line && d.StartChar < cursor.Col {
		return true
	}
	return false
}

// Helper method to check if DefNode comes after otherDef source code
func (d *DefNode) isDefinitionAfter(otherDef *DefNode) bool {
	if d.StartLine > otherDef.StartLine {
		return true
	}
	if d.StartLine == otherDef.StartLine && d.StartChar > otherDef.StartChar {
		return true
	}
	return false
}

// Check if two DefNode instances represent the same definition
func (d *DefNode) isSameDefinition(def2 *DefNode) bool {
	return d.StartLine == def2.StartLine &&
		d.StartChar == def2.StartChar &&
		d.Name == def2.Name
}

// Identify DefNodes
// with tracking processed scoped assignments to avoid duplicates
func (a *Ast) DefNodes() []DefNode {
	defNodes := []DefNode{}

	processedScopedAssignments := make(map[string]bool) // key: "line:col:name"

	syntax.Walk(a.File, func(node syntax.Node) bool {
		var name string
		var startLine, startChar, endLine, endChar uint
		var enclosingFunc *syntax.FuncDecl
		var isScoped bool

		switch n := node.(type) {
		// Variable Assignment that are not part done within a DeclClause (`local`, etc.)
		// These are always global (even inside functions)
		case *syntax.Assign:
			if n.Name != nil {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()

				// Check if this assignment is part of a scoped declaration and skip if it is
				assignmentKey := fmt.Sprintf("%d:%d:%s", startLine, startChar, name)
				if processedScopedAssignments[assignmentKey] {
					return true
				}

				// Check if we're inside a function
				enclosingFunc = a.findEnclosingFunctionForNode(n)
				// Regular assignments are not scoped (they're global even in functions)
				isScoped = false
			}

		// Scoped Variable Declaration with `local`, `declare`, `typeset`
		// These are always scoped variables
		case *syntax.DeclClause:
			cmd := n.Variant.Value
			if cmd == "local" || cmd == "declare" || cmd == "typeset" {
				enclosingFunc = a.findEnclosingFunctionForNode(n)
				isScoped = true

				for _, arg := range n.Args {
					if arg.Name != nil {
						name = arg.Name.Value
						startLine, startChar = arg.Name.ValuePos.Line(), arg.Name.ValuePos.Col()
						endLine, endChar = arg.Name.ValueEnd.Line(), arg.Name.ValueEnd.Col()

						// Mark this assignment as processed to avoid duplicate from *syntax.Assign
						assignmentKey := fmt.Sprintf("%d:%d:%s", startLine, startChar, name)
						processedScopedAssignments[assignmentKey] = true

						slog.Info("Adding DefNode", "name", name, "isScoped", isScoped, "nodeType",
							fmt.Sprintf("%T", node), "line", startLine)
						defNodes = append(defNodes, DefNode{
							Node:      n,
							Name:      name,
							Scope:     enclosingFunc,
							IsScoped:  isScoped,
							StartLine: startLine,
							StartChar: startChar,
							EndLine:   endLine,
							EndChar:   endChar,
						})
					}
				}
				return true
			}
			// If it's not a scoped declaration, don't process it here
			return true

		// Function Definition
		case *syntax.FuncDecl:
			if n.Name != nil {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()
				// Functions are always global
				enclosingFunc = nil
				isScoped = false
			}

		// Iteration variable in for/select loops
		case *syntax.ForClause:
			// For loop variables are scoped to the loop (treat as local to enclosing function)
			enclosingFunc = a.findEnclosingFunctionForNode(n)
			isScoped = (enclosingFunc != nil) // Only scoped if inside a function

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

		// Variable "assignment" in `read` statements
		case *syntax.CallExpr:
			if len(n.Args) == 0 {
				return true
			}
			cmdName := ExtractIdentifier(n.Args[0])
			if cmdName != "read" {
				return true
			}

			// Read variables are scoped to function if inside one
			enclosingFunc = a.findEnclosingFunctionForNode(n)
			isScoped = (enclosingFunc != nil) // Only scoped if inside a function

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

		// Add the definition if we found a name
		if name != "" {
			defNodes = append(defNodes, DefNode{
				Node:      node,
				Name:      name,
				Scope:     enclosingFunc,
				IsScoped:  isScoped,
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

// Helper function to find the enclosing function for a given node
// Check if the target node is within this function's bounds
// Target is inside function if: fnStart <= targetStart < targetEnd <= fnEnd
func (a *Ast) findEnclosingFunctionForNode(targetNode syntax.Node) *syntax.FuncDecl {
	var enclosingFunc *syntax.FuncDecl
	targetStart := targetNode.Pos()
	targetEnd := targetNode.End()

	syntax.Walk(a.File, func(node syntax.Node) bool {
		fn, ok := node.(*syntax.FuncDecl)
		if !ok {
			return true
		}

		fnStart, fnEnd := fn.Pos(), fn.End()

		if (fnStart.Line() < targetStart.Line() ||
			(fnStart.Line() == targetStart.Line() && fnStart.Col() <= targetStart.Col())) &&
		   (targetEnd.Line() < fnEnd.Line() ||
			(targetEnd.Line() == fnEnd.Line() && targetEnd.Col() <= fnEnd.Col())) {
			enclosingFunc = fn
		}
		return true
	})

	return enclosingFunc
}

// // Convenience methods for filtering DefNodes by type
//
// // Get only scoped variable definitions (local, declare, typeset)
// func (a *Ast) ScopedVarDefNodes() []DefNode {
// 	var scopedNodes []DefNode
// 	for _, node := range a.DefNodes() {
// 		if node.IsScoped {
// 			scopedNodes = append(scopedNodes, node)
// 		}
// 	}
// 	return scopedNodes
// }
//
// // Get only global variable definitions (assignments outside functions or not declared as local)
// func (a *Ast) GlobalVarDefNodes() []DefNode {
// 	var globalNodes []DefNode
// 	for _, node := range a.DefNodes() {
// 		// Functions are always global
// 		if _, ok := node.Node.(*syntax.FuncDecl); ok {
// 			globalNodes = append(globalNodes, node)
// 			continue
// 		}
//
// 		// Variables that are not scoped
// 		if !node.IsScoped {
// 			globalNodes = append(globalNodes, node)
// 		}
// 	}
// 	return globalNodes
// }
//
// // Get only function definitions
// func (a *Ast) FunctionDefNodes() []DefNode {
// 	var funcNodes []DefNode
// 	for _, node := range a.DefNodes() {
// 		if _, ok := node.Node.(*syntax.FuncDecl); ok {
// 			funcNodes = append(funcNodes, node)
// 		}
// 	}
// 	return funcNodes
// }

// Updated FindDefInFile that uses the unified DefNodes
func (a *Ast) FindDefInFile(cursor Cursor) *DefNode {
	cursorNode := a.FindNodeUnderCursor(cursor)
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}

	cursorScope := a.findEnclosingFunction(cursor)

	// Find scoped variables in the same function scope as cursor
	if cursorScope != nil {
		for _, defNode := range a.DefNodes() {
			if defNode.Name == targetIdentifier && defNode.IsScoped && defNode.Scope == cursorScope {
				// Check if the scoped variable is declared before the cursor position (shadowing)
				if defNode.isBeforeCursor(cursor) {
					return &defNode
				}
			}
		}
	}

	// Find global definitions i.e., functions and non-scoped variables
	for _, defNode := range a.DefNodes() {
		if defNode.Name == targetIdentifier {
			if defNode.IsScoped {
				continue
			}
			return &defNode
		}
	}

	return nil
}

