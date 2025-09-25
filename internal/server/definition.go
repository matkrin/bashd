package server

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/utils"
)

func handleDefinition(request *lsp.DefinitionRequest, state *State) *lsp.DefinitionResponse {
	uri := request.Params.TextDocument.URI
	cursor := ast.NewCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(document, uri)
	if err != nil {
		slog.Error(err.Error())
	}
	cursorNode := fileAst.FindNodeUnderCursor(cursor)
	definition := fileAst.FindDefInFile(cursor)

	if definition == nil {
		// Check for the definition in a sourced file
		filename, err := utils.UriToPath(uri)
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		baseDir := filepath.Dir(filename)
		sourcedFile := ""
		sourcedFile, definition = fileAst.FindDefInSourcedFile(
			cursor,
			state.EnvVars,
			baseDir,
		)

		if definition != nil {
			uri = utils.PathToURI(sourcedFile)
		}
	}

	if definition == nil {
		// Check if the cursor is over a filename in a source statement
		filename, err := utils.UriToPath(uri)
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		baseDir := filepath.Dir(filename)
		sourcePath := fileAst.FindSourcedFile(cursor, state.EnvVars, baseDir)
		// Check if file exists
		if _, err := os.Stat(sourcePath); err != nil {
			return nil
		}
		definition = &ast.DefNode{
			Node:      cursorNode,
			StartLine: 1,
			StartChar: 1,
			EndLine:   1,
			EndChar:   1,
		}
		uri = utils.PathToURI(sourcePath)
	}

	if definition == nil {
		return nil
	}

	response := lsp.NewDefinitionResponse(
		request.ID,
		uri,
		definition.StartLine-1,
		definition.StartChar-1,
		definition.EndLine-1,
		definition.EndChar-1,
	)
	return &response
}

