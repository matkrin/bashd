package server

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/utils"
	"mvdan.cc/sh/v3/syntax"
)

func handleHover(request *lsp.HoverRequest, state *State) *lsp.HoverResponse {
	uri := request.Params.TextDocument.URI
	cursor := ast.NewCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)
	documentText := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(documentText, uri, true)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	cursorNode := fileAst.FindNodeUnderCursor(cursor)
	if cursorNode == nil {
		return nil
	}

	hoverResultValue := hoverFromDefinition(fileAst, cursor, state, uri)

	identifier := ast.ExtractIdentifier(cursorNode)
	documentation := getDocumentation(identifier)
	if documentation != "" {
		hoverResultValue = fmt.Sprintf("```man\n%s\n```", documentation)
	}

	if hoverResultValue == "" {
		return nil
	}

	response := lsp.HoverResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
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

func defNodeToHoverString(defNode *ast.DefNode, documentText string, documentName string) string {
	if n, ok := defNode.Node.(*syntax.FuncDecl); ok {
		lines := strings.Split(documentText, "\n")
		functionSnippet := strings.Join(lines[n.Pos().Line()-1:n.End().Line()], "\n")

		defLocation := fmt.Sprintf("defined at `%s` line **%d**", documentName, n.Pos().Line())
		if documentName == "" {
		defLocation = fmt.Sprintf("defined at line **%d**", n.Pos().Line())
		}

		return fmt.Sprintf("```sh\n%s\n```\n\n(%s)", functionSnippet, defLocation)
	}

	if documentName == "" {
		return fmt.Sprintf("defined at line **%d**", defNode.StartLine)
	}

	return fmt.Sprintf("defined at `%s` line **%d**", documentName, defNode.StartLine)
}

func hoverFromDefinition(
	ast *ast.Ast,
	cursor ast.Cursor,
	state *State,
	uri string,
) string {
	defNode := ast.FindDefInFile(cursor)
	if defNode != nil {
		documentText := state.Documents[uri].Text
		return defNodeToHoverString(defNode, documentText, "")
	}

	filename, err := utils.UriToPath(uri)
	if err != nil {
		slog.Error("ERROR: Could not convert uri to path", "uri", uri)
		return ""
	}
	baseDir := filepath.Dir(filename)
	sourcedFile, definition := ast.FindDefInSourcedFile(
		cursor,
		state.EnvVars,
		baseDir,
	)
	fileContent, err := os.ReadFile(sourcedFile)
	if err != nil {
		slog.Error("ERROR: Could not read file", "file", sourcedFile)
		return ""
	}
	return defNodeToHoverString(definition, string(fileContent), sourcedFile)
}
