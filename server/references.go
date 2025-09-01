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
	for _, refNode := range referenceNodes {
		locations = append(locations, refNode.toLspLocation(uri))
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
		for _, refNode := range refNodes {
			locations = append(locations, refNode.toLspLocation(pathToURI(file)))
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
	Node      *syntax.Node
	Name      string
	StartLine uint
	StartChar uint
	EndLine   uint
	EndChar   uint
}

func (r *RefNode) toLspLocation(uri string) lsp.Location {
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

func (r *RefNode) toLspTextEdit(newText string) lsp.TextEdit {
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

func refNodes(ast *syntax.File, includeDeclaration bool) []RefNode {
	refNodes := []RefNode{}

	syntax.Walk(ast, func(node syntax.Node) bool {
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
