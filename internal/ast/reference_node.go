package ast

import (
	"fmt"
	"log/slog"

	"github.com/matkrin/bashd/internal/lsp"
	"mvdan.cc/sh/v3/syntax"
)

// Wraps a node that can be part of a reference.
type RefNode struct {
	Node      syntax.Node
	Name      string
	Scope     *syntax.FuncDecl
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

// Enhanced refNodes that handles scoped declarations properly
func (a *Ast) RefNodes(includeDeclaration bool) []RefNode {
	refNodes := []RefNode{}

	// Track processed scoped assignments to avoid duplicates (same as in DefNodes)
	processedScopedAssignments := make(map[string]bool) // key: "line:col:name"

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

		// Function usage (commands) and read statements
		case *syntax.CallExpr:
			if len(n.Args) > 0 {
				cmdName := ExtractIdentifier(n.Args[0])

				// Variable assignments as part of read statements
				if cmdName == "read" && includeDeclaration {
					for _, arg := range n.Args[1:] { // Skip the "read" command itself
						for _, wp := range arg.Parts {
							switch p := wp.(type) {
							case *syntax.Lit:
								name = p.Value
								startLine, startChar = p.Pos().Line(), p.Pos().Col()
								endLine, endChar = p.End().Line(), p.End().Col()
							}
						}
					}
				} else if cmdName != "" && cmdName != "local" && cmdName != "declare" && cmdName != "typeset" && cmdName != "read" {
					// Function calls (except scoping commands or read statements)
					arg := n.Args[0]
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

		// Function declaration
		case *syntax.FuncDecl:
			if n.Name != nil && includeDeclaration {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()
			}

		// Variable assignment without `local`, `declare`, `typeset`
		case *syntax.Assign:
			if n.Name != nil && includeDeclaration {
				name = n.Name.Value
				startLine, startChar = n.Name.Pos().Line(), n.Name.Pos().Col()
				endLine, endChar = n.Name.End().Line(), n.Name.End().Col()

				// Check if this assignment is part of a scoped declaration
				assignmentKey := fmt.Sprintf("%d:%d:%s", startLine, startChar, name)
				if processedScopedAssignments[assignmentKey] {
					// Skip this - it's already been processed as part of a DeclClause
					return true
				}
			}

		// Scoped variable declarations with `local`, `declare`, `typeset`
		case *syntax.DeclClause:
			if includeDeclaration {
				cmd := n.Variant.Value
				if cmd == "local" || cmd == "declare" || cmd == "typeset" {
					for _, arg := range n.Args {
						if arg.Name != nil {
							name = arg.Name.Value
							startLine, startChar = arg.Name.ValuePos.Line(), arg.Name.ValuePos.Col()
							endLine, endChar = arg.Name.ValueEnd.Line(), arg.Name.ValueEnd.Col()

							// Mark this assignment as processed to avoid duplicate from *syntax.Assign
							assignmentKey := fmt.Sprintf("%d:%d:%s", startLine, startChar, name)
							processedScopedAssignments[assignmentKey] = true

							// Create a separate RefNode for each variable in the declaration
							cursor := Cursor{Line: startLine, Col: startChar}
							scope := a.findEnclosingFunction(cursor)

							refNodes = append(refNodes, RefNode{
								Node:      n,
								Name:      name,
								Scope:     scope,
								StartLine: startLine,
								StartChar: startChar,
								EndLine:   endLine,
								EndChar:   endChar,
							})
						}
					}
					return true // Continue walking, but don't add to the general list
				}
			}

		// Iteration variable in for/select loops and C-style loops
		case *syntax.ForClause:
			if includeDeclaration {
				switch loop := n.Loop.(type) {
				// for/select
				case *syntax.WordIter:
					if loop.Name != nil {
						name = loop.Name.Value
						startLine, startChar = loop.Name.Pos().Line(), loop.Name.Pos().Col()
						endLine, endChar = loop.Name.End().Line(), loop.Name.End().Col()
					}
				// C-style
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
			}
		}

		if name != "" {
			cursor := Cursor{Line: startLine, Col: startChar}
			scope := a.findEnclosingFunction(cursor)

			refNodes = append(refNodes, RefNode{
				Node:      node,
				Name:      name,
				Scope:     scope,
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

// Enhanced FindRefsInFile with proper scoping logic
// Fixed FindRefsInFile with proper scoping logic
func (a *Ast) FindRefsInFile(cursor Cursor, includeDeclaration bool) []RefNode {
	cursorNode := a.FindNodeUnderCursor(cursor)
	targetIdentifier := ExtractIdentifier(cursorNode)
	if targetIdentifier == "" {
		return nil
	}
	slog.Info("REFS", "includeDeclaration", includeDeclaration)

	references := []RefNode{}

	// Find the definition that the cursor is pointing to
	defNode := a.FindDefInFile(cursor)

	slog.Info("FINDREFS", "DEFNODE", defNode)

	if defNode == nil {
		// No definition found - return all references with same name (fallback behavior)
		for _, refNode := range a.RefNodes(includeDeclaration) {
			if refNode.Name == targetIdentifier {
				references = append(references, refNode)
			}
		}
		return references
	}

	// Definition found - find all references that would resolve to this same definition
	for _, refNode := range a.RefNodes(includeDeclaration) {
		if refNode.Name != targetIdentifier {
			continue
		}

		if a.wouldResolveToSameDefinition(refNode.Node, defNode) {
			references = append(references, refNode)
		}
	}

	return references
}

// Fixed wouldResolveToSameDefinition that properly handles declaration ordering
func (a *Ast) wouldResolveToSameDefinition(refCursorNode syntax.Node, targetDefNode *DefNode) bool {
	// Simulate FindDefInFile logic at the reference location
	pos := refCursorNode.Pos()
	cursor := Cursor{Line: pos.Line(), Col: pos.Col()}
	refScope := a.findEnclosingFunction(cursor)

	// Apply the same resolution logic as FindDefInFile
	targetIdentifier := targetDefNode.Name

	// First, look for scoped variables in the same function scope
	if refScope != nil {
		// Find the closest local declaration that comes BEFORE the reference
		var closestLocalDef *DefNode

		for _, defNode := range a.DefNodes() {
			if defNode.Name == targetIdentifier && defNode.IsScoped && defNode.Scope == refScope {
				// Check if the scoped variable is declared BEFORE the reference position
				if defNode.isBeforeCursor( cursor) {
					// Among all local variables declared before this reference,
					// find the one that's closest (latest declaration)
					if closestLocalDef == nil || defNode.isDefinitionAfter(closestLocalDef) {
						closestLocalDef = &defNode
					}
				}
			}
		}

		// If we found a local variable declared before the reference, use it
		if closestLocalDef != nil {
			return closestLocalDef.isSameDefinition(targetDefNode)
		}
	}

	// No local variable found that's declared before the reference
	// Look for global definitions (functions and non-scoped variables)
	for _, defNode := range a.DefNodes() {
		if defNode.Name == targetIdentifier {
			// Skip ALL scoped variables (they're only visible within their function)
			if defNode.IsScoped {
				continue
			}
			// This reference would resolve to this global definition
			return defNode.isSameDefinition(targetDefNode)
		}
	}

	return false
}

// Fixed IsVariableShadowed method to work with RefNode instead of cursorNode
func (a *Ast) IsVariableShadowed(varName string, cursor Cursor) (bool, *DefNode) {
	currentScope := a.findEnclosingFunction(cursor)

	// Look for local variables that shadow globals
	for _, defNode := range a.DefNodes() {
		if defNode.Name != varName {
			continue
		}

		// Only consider scoped (local) variables for shadowing
		if !defNode.IsScoped {
			continue
		}

		// Check if the local variable is in the same scope as the cursor
		if defNode.Scope == currentScope {
			// Check if the local variable is declared before the cursor position
			if defNode.isBeforeCursor(cursor) {
				// Found a local variable that shadows any global with the same name
				return true, &defNode
			}
		}
	}

	return false, nil
}

// Alternative version if you want to keep the original signature
func (a *Ast) IsVariableShadowedByNode(varName string, cursorNode syntax.Node) (bool, *DefNode) {
	pos := cursorNode.Pos()
	cursor := Cursor{Line: pos.Line(), Col: pos.Col()}
	return a.IsVariableShadowed(varName, cursor)
}

