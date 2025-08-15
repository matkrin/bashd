package server

import (
	"log/slog"
	"os"
	"path/filepath"

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
		slog.Error(err.Error())
	}
	cursorNode := findNodeUnderCursor(fileAst, cursor)
	referenceNodes := findRefsInFile(fileAst, cursorNode, params.Context.IncludeDeclaration)

	// In current file
	locations := []lsp.Location{}
	for _, node := range referenceNodes {
		location := lsp.Location{
			URI: uri,
			Range: lsp.NewRange(
				node.Start.Line()-1,
				node.Start.Col()-1,
				node.End.Line()-1,
				node.End.Col()-1,
			),
		}
		locations = append(locations, location)
	}

	// In sourced files
	filename, err := uriToPath(uri)
	if err != nil {
		slog.Error(err.Error())
	}
	baseDir := filepath.Dir(filename)
	referenceNodesInSourcedFiles := findRefsinSourcedFile(
		fileAst,
		cursorNode,
		state.EnvVars,
		baseDir,
		params.Context.IncludeDeclaration,
	)

	for file, refNodes := range referenceNodesInSourcedFiles {
		for _, node := range refNodes {
			location := lsp.Location{
				URI: pathToURI(file),
				Range: lsp.NewRange(
					node.Start.Line()-1,
					node.Start.Col()-1,
					node.End.Line()-1,
					node.End.Col()-1,
				),
			}
			locations = append(locations, location)
		}
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
	if targetIdentifier == "" {
		return nil
	}

	references := []RefNode{}
	for _, node := range refNodes(file, includeDeclaration) {
		if node.Name == targetIdentifier {
			references = append(references, node)
		}

	}

	return references
}

func findRefsinSourcedFile(
	fileAst *syntax.File,
	cursorNode syntax.Node,
	env map[string]string,
	baseDir string,
	includeDeclaration bool,
) map[string][]RefNode {
	sourcedFiles := findAllSourcedFiles(fileAst, env, baseDir, map[string]bool{})

	filesRefNodes := map[string][]RefNode{}
	for _, sourcedFile := range sourcedFiles {
		fileContent, err := os.ReadFile(sourcedFile)
		if err != nil {
			slog.Error("Could not read file", "file", sourcedFile)
			continue
		}
		sourcedFileAst, err := parseDocument(string(fileContent), sourcedFile)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		references := findRefsInFile(sourcedFileAst, cursorNode, includeDeclaration)
		filesRefNodes[sourcedFile] = references
	}

	return filesRefNodes
}
