package server

import (
	"log/slog"
	"path/filepath"

	"github.com/matkrin/bashd/ast"
	"github.com/matkrin/bashd/lsp"
)

// "declare", "local", "export", "readonly", "typeset", or "nameref".
func handleDeclaration(request *lsp.DeclarationRequest, state *State) *lsp.DeclarationResponse {
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
	declaration := fileAst.FindDeclInFile(cursorNode)

	if declaration == nil {
		// Check for the declaration in a sourced file
		filename, err := uriToPath(uri)
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		baseDir := filepath.Dir(filename)
		sourcedFile := ""
		sourcedFile, declaration = fileAst.FindDeclInSourcedFile(
			cursorNode,
			state.EnvVars,
			baseDir,
		)

		if declaration != nil {
			uri = pathToURI(sourcedFile)
		}
	}

	return lsp.NewDeclarationResponse(
		request.ID,
		uri,
		declaration.StartLine-1,
		declaration.StartChar-1,
		declaration.EndLine-1,
		declaration.EndChar-1,
	)
}
