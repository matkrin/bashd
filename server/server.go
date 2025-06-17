package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/matkrin/bashd/ast"
	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

type Document struct {
	Text         string
	SourcedFiles []Document
}

type State struct {
	Documents        map[string]Document
	EnvVars          map[string]string
	WorkspaceFolders []lsp.WorkspaceFolder
	PathItems        []string
}

func NewState() State {
	envVars := getEnvVars()
	pathItems := getPathItems(envVars)

	return State{
		Documents: map[string]Document{},
		EnvVars:   envVars,
		PathItems: pathItems,
	}
}

func (s *State) OpenDocument(uri, text string) {
	s.Documents[uri] = Document{
		Text:         text,
		SourcedFiles: []Document{},
	}
}
func (s *State) UpdateDocument(uri, text string) {
	s.Documents[uri] = Document{
		Text:         text,
		SourcedFiles: []Document{},
	}
}

func getEnvVars() map[string]string {
	env := os.Environ()
	envVars := make(map[string]string)

	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		key := pair[0]
		value := ""
		if len(pair) > 1 {
			value = pair[1]
		}
		envVars[key] = value
	}

	return envVars
}

func getPathItems(envVars map[string]string) []string {
	pathStr := envVars["PATH"]
	pathItems := []string{}
	for pathPart := range strings.SplitSeq(pathStr, ":") {
		entries, err := os.ReadDir(pathPart)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				logger.Errorf("Error getting the info of %s", entry.Name())
				continue
			}

			perm := info.Mode().Perm()
			// Check if file is executable
			if perm&0111 != 0 {
				pathItems = append(pathItems, entry.Name())
			}
		}
	}
	return pathItems
}

func HandleMessage(writer io.Writer, state State, method string, contents []byte) {
	logger.Infof("Received msg with method: `%s`", method)

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		logger.Infof(
			"Connected to: %s %s",
			request.Params.ClientInfo.Name,
			request.Params.ClientInfo.Version,
		)

		state.WorkspaceFolders = request.Params.WorkspaceFolders
		logger.Infof("Workspace folders set to: %#v", state.WorkspaceFolders)

		msg := lsp.NewInitializeResponse(request.ID)
		writeResponse(writer, msg)

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		logger.Infof("Opened: %s", request.Params.TextDocument.URI)

		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)

	case "textDocument/didChange":
		var request lsp.TextDocumentDidChangeNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		logger.Infof("Changed: %s", request.Params.TextDocument.URI)
		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}

	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		documentName := request.Params.TextDocument.URI
		cursor := ast.NewCursor(
			request.Params.Position.Line,
			request.Params.Position.Character,
		)
		document := state.Documents[documentName].Text
		fileAst, err := ast.ParseDocument(document, documentName)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		node := ast.FindNodeUnderCursor(fileAst, cursor)

		if node != nil {
			var buf bytes.Buffer
			syntax.DebugPrint(&buf, node)

			response := lsp.HoverResponse{
				Response: lsp.Response{
					RPC: "2.0",
					ID:  &request.ID,
				},
				Result: lsp.HoverResult{
					Contents: buf.String(),
				},
			}

			writeResponse(writer, response)
		}

	case "textDocument/definition":
		var request lsp.DefinitionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		response := handleDefinition(&request, &state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/completion":
		var request lsp.CompletionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		completionList := []lsp.CompletionItem{}
		for _, keyword := range KEYWORDS {
			completionItem := lsp.CompletionItem{
				Label:         keyword,
				Kind:          lsp.CompletionKeyword,
				Detail:        "detail",
				Documentation: "doc",
			}
			completionList = append(completionList, completionItem)
		}

		for _, pathItem := range state.PathItems {
			completionItem := lsp.CompletionItem{
				Label:         pathItem,
				Kind:          lsp.CompletionFunction,
				Detail:        "detail",
				Documentation: "doc",
			}
			completionList = append(completionList, completionItem)
		}

		response := lsp.CompletionResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  &request.ID,
			},
			Result: completionList,
		}
		writeResponse(writer, response)

	}

}

var KEYWORDS = [...]string{
	"if",
	"then",
	"elif",
	"else",
	"fi",
	"for",
	"in",
	"do",
	"done",
	"case",
	"esac",
	"select",
	"function",
	"{",
	"}",
	"[[",
	"]]",
	"!",
	"time",
	"until",
	"while",
	"coproc",
}

func handleDefinition(request *lsp.DefinitionRequest, state *State) *lsp.DefinitionResponse {
	uri := request.Params.TextDocument.URI
	cursor := ast.NewCursor(
		request.Params.Position.Line,
		request.Params.Position.Character,
	)

	document := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(document, uri)
	if err != nil {
		logger.Error(err.Error())
	}
	cursorNode := ast.FindNodeUnderCursor(fileAst, cursor)
	definition := ast.FindDefinition(cursorNode, fileAst)

	if definition == nil {
		// Check for the definition in a sourced file
		filename, err := uriToPath(uri)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		sourcedFile := ""
		sourcedFile, definition = ast.FindDefInSourcedFile(filename,
			fileAst, cursorNode, state.EnvVars)

		if definition != nil {
			uri = pathToURI(sourcedFile)
		}
	}

	if definition == nil {
		// Check if we are over a filename in a source statement
		filename, err := uriToPath(uri)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		baseDir := filepath.Dir(filename)
		sourcePath := ast.FindSourcedFile(fileAst, cursor, state.EnvVars, baseDir)
		// Check file exists
		if _, err := os.Stat(sourcePath); err != nil {
			// Send diagnostic here?
			return nil
		}
		definition = &ast.DefNode{
			Node:  cursorNode,
			Start: syntax.NewPos(0, 1, 1),
			End:   syntax.NewPos(0, 1, 1),
		}
		uri = pathToURI(sourcePath)
	}

	if definition == nil {
		return nil
	}

	response := lsp.NewDefinitionResponse(
		request.ID,
		uri,
		int(definition.Start.Line())-1,
		int(definition.Start.Col())-1,
		int(definition.End.Line())-1,
		int(definition.End.Col())-1,
	)
	return &response
}

func writeResponse(writer io.Writer, msg any) {
	reply := lsp.EncodeMessage(msg)
	logger.Info(reply)
	writer.Write([]byte(reply))
}

func uriToPath(uri string) (string, error) {
	if !strings.HasPrefix(uri, "file://") {
		return "", fmt.Errorf("unsupported URI scheme")
	}
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}

func pathToURI(path string) string {
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	return uri.String()
}
