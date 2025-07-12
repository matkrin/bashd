package server

import (
	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleReferences(request *lsp.ReferencesRequest, state *State) *lsp.ReferencesResponse {
	params := request.Params
	uri := params.TextDocument.URI
	cursor := newCursor(
		params.Position.Line,
		params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		logger.Error(err.Error())
	}
	cursorNode := findNodeUnderCursor(fileAst, cursor)
	referenceNodes := findRefsInFile(fileAst, cursorNode, params.Context.IncludeDeclaration)

	locations := []lsp.Location{}
	for _, node := range referenceNodes {
		location := lsp.Location{
			URI: uri,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      int(node.Start.Line()) - 1,
					Character: int(node.Start.Col()) - 1,
				},
				End: lsp.Position{
					Line:      int(node.End.Line()) - 1,
					Character: int(node.End.Col()) - 1,
				},
			},
		}
		locations = append(locations, location)
	}

	response := lsp.ReferencesResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: locations,
	}

	return &response
}

// Wraps a node that can be part of a reference.
type RefNode struct {
	Node  *syntax.Node
	Name  string
	Start syntax.Pos
	End   syntax.Pos
}

func refNodes(ast *syntax.File, includeDeclaration bool) []RefNode {
	refNodes := []RefNode{}

	syntax.Walk(ast, func(node syntax.Node) bool {
		var name string
		var pos, end syntax.Pos

		switch n := node.(type) {
		// Variable usage
		case *syntax.ParamExp:
			if n.Param != nil {
				name = n.Param.Value
				pos, end = n.Param.Pos(), n.Param.End()
			}
		// Function usage
		case *syntax.Word:
			if len(n.Parts) == 1 {
				switch p := n.Parts[0].(type) {
				case *syntax.Lit:
					name = p.Value
					pos, end = p.Pos(), p.End()
				}
			}
		// Funtion declaration
		case *syntax.FuncDecl:
			if n.Name != nil && includeDeclaration {
				name = n.Name.Value
				pos, end = n.Name.Pos(), n.Name.End()
			}
		// Variable assignement
		case *syntax.Assign:
			if n.Name != nil && includeDeclaration {
				name = n.Name.Value
				pos, end = n.Name.Pos(), n.Name.End()
			}
		}

		if name != "" {
			refNodes = append(refNodes, RefNode{
				Node:  &node,
				Name:  name,
				Start: pos,
				End:   end,
			})
		}

		return true
	})

	return refNodes
}

func findRefsInFile(file *syntax.File, cursorNode syntax.Node, includeDeclaration bool) []RefNode {
	targetIdentifier := extractIdentifier(cursorNode)
	logger.Infof("targetIdentifier : %v", targetIdentifier)
	if targetIdentifier == "" {
		return nil
	}

	references := []RefNode{}
	for _, node := range refNodes(file, includeDeclaration) {
		logger.Infof("node.Name : %v", node.Name)
		if node.Name == targetIdentifier {
			references = append(references, node)
		}

	}

	return references
}
