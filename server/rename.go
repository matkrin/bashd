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

	// Check if rename target is a executable in PATH or a builtin
	identifier := extractIdentifier(cursorNode)
	if slices.Contains(state.PathItems, identifier) || slices.Contains(BASH_BUILTINS[:], identifier) {
		return nil
	}

	response := lsp.PrepareRenameResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: lsp.Range{
			Start: lsp.Position{
				Line:      cursorNode.Pos().Line() - 1,
				Character: cursorNode.Pos().Col() - 1,
			},
			End: lsp.Position{
				Line:      cursorNode.End().Line() - 1,
				Character: cursorNode.End().Col() - 1,
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
		slog.Error(err.Error())
	}
	cursorNode := findNodeUnderCursor(fileAst, cursor)
	referenceNodes := findRefsInFile(fileAst, cursorNode, true)

	slog.Info("Handle rename", "referenceNodes", referenceNodes)
	if len(referenceNodes) == 0 {
		return nil
	}

	changes := map[string][]lsp.TextEdit{}
	for _, node := range referenceNodes {
		changes[uri] = append(changes[uri], lsp.TextEdit{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      node.Start.Line() - 1,
					Character: node.Start.Col() - 1,
				},
				End: lsp.Position{
					Line:      node.End.Line() - 1,
					Character: node.End.Col() - 1,
				},
			},
			NewText: params.NewName,
		})
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
		true,
	)

	for file, refNodes := range referenceNodesInSourcedFiles {
		fileUri := pathToURI(file)
		for _, node := range refNodes {
			changes[fileUri] = append(changes[fileUri], lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      node.Start.Line() - 1,
						Character: node.Start.Col() - 1,
					},
					End: lsp.Position{
						Line:      node.End.Line() - 1,
						Character: node.End.Col() - 1,
					},
				},
				NewText: params.NewName,
			})
		}
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
