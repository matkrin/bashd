package server

import (
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/matkrin/bashd/lsp"
)

func handlePrepareRename(
	request *lsp.PrepareRenameRequest,
	state *State,
) *lsp.PrepareRenameResponse {
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
	referenceNodes := findRefsInFile(fileAst, cursorNode, true)

	slog.Info("Prepare rename", "referenceNodes", referenceNodes)
	if len(referenceNodes) == 0 {
		return nil
	}

	identifier := extractIdentifier(cursorNode)
	if slices.Contains(state.PathItems, identifier) || slices.Contains(BASH_BUILTINS[:], identifier) {
		return nil
	}

	response := lsp.PrepareRenameResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: lsp.NewRange(
			cursorNode.Pos().Line()-1,
			cursorNode.Pos().Col()-1,
			cursorNode.End().Line()-1,
			cursorNode.End().Col()-1,
		),
	}
	return &response
}

func handleRename(request *lsp.RenameRequest, state *State) *lsp.RenameResponse {
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
	referenceNodes := findRefsInFile(fileAst, cursorNode, true)

	changes := map[string][]lsp.TextEdit{}
	changes[uri] = findTextEditsInFile(referenceNodes, params.NewName)

	// In sourced files
	filename, err := uriToPath(uri)
	if err != nil {
		slog.Error("Could not transform URI to path", "err", err.Error())
	}
	baseDir := filepath.Dir(filename)
	referenceNodesInSourcedFiles := findRefsinSourcedFile(
		fileAst,
		cursorNode,
		state.EnvVars,
		baseDir,
		true,
	)

	for file, refNodes := range referenceNodesInSourcedFiles {
		fileUri := pathToURI(file)
		changes[fileUri] = append(
			changes[fileUri],
			findTextEditsInFile(refNodes, params.NewName)...,
		)
	}

	response := lsp.RenameResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: &lsp.WorkspaceEdit{
			Changes: changes,
		},
	}
	return &response
}

func findTextEditsInFile(referenceNodes []RefNode, newText string) []lsp.TextEdit {
	textEdits := []lsp.TextEdit{}
	for _, refNode := range referenceNodes {
		textEdits = append(textEdits, refNode.toLspTextEdit(newText))
	}
	return textEdits
}
