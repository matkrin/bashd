package server

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func handleHover(request *lsp.HoverRequest, state *State) *lsp.HoverResponse {

	uri := request.Params.TextDocument.URI
	cursor := newCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)
	documentText := state.Documents[uri].Text
	fileAst, err := parseDocument(documentText, uri)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	cursorNode := findNodeUnderCursor(fileAst, cursor)
	if cursorNode == nil {
		return nil
	}

	hoverResultValue := hoverFromDefinition(fileAst, cursorNode, state, uri)

	identifier := extractIdentifier(cursorNode)
	documentation := getDocumentation(identifier)
	if strings.Trim(documentation, "\n") != "" {
		hoverResultValue = fmt.Sprintf("```man\n%s\n```", documentation)
	}

	if hoverResultValue == "" {
		return nil
	}

	response := lsp.HoverResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: lsp.HoverResult{
			Contents: lsp.MarkupContent{
				Kind:  lsp.MarkupKindMarkdown,
				Value: hoverResultValue,
			},
		},
	}
	return &response
}

func (d *DefNode) toHoverString(documentText string, documentName string) string {
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

func hoverFromDefinition(
	fileAst *syntax.File,
	cursorNode syntax.Node,
	state *State,
	uri string,
) string {
	defNode := findDefInFile(cursorNode, fileAst)
	if defNode != nil {
		documentText := state.Documents[uri].Text
		return defNode.toHoverString(documentText, "")
	}

	filename, err := uriToPath(uri)
	if err != nil {
		slog.Error("ERROR: Could not convert uri to path", "uri", uri)
		return ""
	}
	baseDir := filepath.Dir(filename)
	sourcedFile, definition := findDefInSourcedFile(
		fileAst,
		cursorNode,
		state.EnvVars,
		baseDir,
	)
	fileContent, err := os.ReadFile(sourcedFile)
	if err != nil {
		slog.Error("ERROR: Could not read file", "file", sourcedFile)
		return ""
	}
	return definition.toHoverString(string(fileContent), sourcedFile)
}
