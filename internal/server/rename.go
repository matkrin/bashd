package server

import (
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/utils"
)

func handlePrepareRename(
	request *lsp.PrepareRenameRequest,
	state *State,
) *lsp.PrepareRenameResponse {
	params := request.Params
	uri := params.TextDocument.URI
	cursor := ast.NewCursor(
		params.Position.Line,
		params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(document, uri, false)
	if err != nil {
		slog.Error(err.Error())
	}
	cursorNode := fileAst.FindNodeUnderCursor(cursor)
	referenceNodes := fileAst.FindRefsInFile(cursor, true)

	slog.Info("Prepare rename", "referenceNodes", referenceNodes)
	if len(referenceNodes) == 0 {
		return nil
	}

	identifier := ast.ExtractIdentifier(cursorNode)
	if slices.Contains(state.PathItems, identifier) || slices.Contains(BASH_BUILTINS[:], identifier) {
		return nil
	}

	response := lsp.PrepareRenameResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
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
	cursor := ast.NewCursor(
		params.Position.Line,
		params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(document, uri, false)
	if err != nil {
		slog.Error(err.Error())
	}
	referenceNodes := fileAst.FindRefsInFile(cursor, true)

	changes := map[string][]lsp.TextEdit{}
	changes[uri] = findTextEditsInFile(referenceNodes, params.NewName)

	// In sourced files
	filename, err := utils.UriToPath(uri)
	if err != nil {
		slog.Error("Could not transform URI to path", "err", err.Error())
	}
	baseDir := filepath.Dir(filename)
	referenceNodesInSourcedFiles := fileAst.FindRefsinSourcedFile(
		cursor,
		state.EnvVars,
		baseDir,
		true,
	)

	for file, refNodes := range referenceNodesInSourcedFiles {
		fileUri := utils.PathToURI(file)
		changes[fileUri] = append(
			changes[fileUri],
			findTextEditsInFile(refNodes, params.NewName)...,
		)
	}

	// In workspace files that source current file
	refNodesInWorkspaceFile := fileAst.FindRefsInWorkspaceFiles(
		uri,
		state.WorkspaceShFiles(),
		baseDir,
		cursor,
		state.EnvVars,
		true,
	)

	for file, refNodes := range refNodesInWorkspaceFile {
		for _, refNode := range refNodes {
			fileUri := utils.PathToURI(file)
			textEdit := refNode.ToLspTextEdit(params.NewName)
			if !slices.Contains(changes[fileUri], textEdit) {
				changes[fileUri] = append(changes[fileUri], textEdit)
			}
		}
	}

	response := lsp.RenameResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
			ID:  &request.ID,
		},
		Result: &lsp.WorkspaceEdit{
			Changes: changes,
		},
	}
	return &response
}

func findTextEditsInFile(referenceNodes []ast.RefNode, newText string) []lsp.TextEdit {
	textEdits := []lsp.TextEdit{}
	for _, refNode := range referenceNodes {
		textEdits = append(textEdits, refNode.ToLspTextEdit(newText))
	}
	return textEdits
}
