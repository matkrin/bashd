package server

import (
	"slices"

	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
)

func handlePrepareRename(request *lsp.PrepareRenameRequest, state *State) *lsp.PrepareRenameResponse {
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
	referenceNodes := findRefsInFile(fileAst, cursorNode, true)

	logger.Infof("referenceNodes : %#v", referenceNodes)
	if len(referenceNodes) == 0 {
		return nil
	}

	// Check if rename target is a executable in PATH
	identifier := extractIdentifier(cursorNode)
	if slices.Contains(state.PathItems, identifier) {
		return nil
	}

	// TODO: Check if rename target is bash builtin

	response := lsp.PrepareRenameResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: lsp.Range{
			Start: lsp.Position{
				Line:      int(cursorNode.Pos().Line()) - 1,
				Character: int(cursorNode.Pos().Col()) - 1,
			},
			End: lsp.Position{
				Line:      int(cursorNode.End().Line()) - 1,
				Character: int(cursorNode.End().Col()) - 1,
			},
		},
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
		logger.Error(err.Error())
	}
	cursorNode := findNodeUnderCursor(fileAst, cursor)
	referenceNodes := findRefsInFile(fileAst, cursorNode, true)

	if len(referenceNodes) == 0 {
		return nil
	}

	response := lsp.RenameResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{},
		},
	}
	return &response
}
