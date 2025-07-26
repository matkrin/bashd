package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func HandleMessage(writer io.Writer, state *State, method string, contents []byte) {
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

	case "shutdown":
		var request lsp.ShutdownRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		logger.Info("Shutdown server")
		response := lsp.ShutdownResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  &request.ID,
			},
			Result: nil,
		}
		writeResponse(writer, response)

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		logger.Infof("Opened: %s", request.Params.TextDocument.URI)

		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)

		diagnostics := checkFile(request.Params.TextDocument.URI, state)
		pushDiagnostic(writer, request.Params.TextDocument.URI, diagnostics)

	case "textDocument/didChange":
		var request lsp.TextDocumentDidChangeNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		logger.Infof("Changed: %s", request.Params.TextDocument.URI)
		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}

		diagnostics := checkFile(request.Params.TextDocument.URI, state)
		pushDiagnostic(writer, request.Params.TextDocument.URI, diagnostics)

	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}

		documentName := request.Params.TextDocument.URI
		cursor := newCursor(
			request.Params.Position.Line,
			request.Params.Position.Character,
		)
		document := state.Documents[documentName].Text
		fileAst, err := parseDocument(document, documentName)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		node := findNodeUnderCursor(fileAst, cursor)

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
		response := handleDefinition(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/references":
		var request lsp.ReferencesRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse '%s' request", method)
		}
		response := handleReferences(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/completion":
		var request lsp.CompletionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		response := handleCompletion(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "completionItem/resolve":

		logger.Infof("RESOLVE REQUEST ##################")
		var request lsp.CompletionItemResolveRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		logger.Infof("RESOLVE REQUEST: %v", request)
		response := handleCompletionItemResolve(&request)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/documentSymbol":
		var request lsp.DocumentSymbolsRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		response := handleDocumentSymbol(&request, state)
		writeResponse(writer, response)

	case "textDocument/prepareRename":
		var request lsp.PrepareRenameRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		response := handlePrepareRename(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/rename":
		var request lsp.RenameRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		response := handleRename(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "workspace/symbol":
		var request lsp.WorkspaceSymbolRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Errorf("Could not parse `%s' request", method)
		}
		response := handleWorkspaceSymbol(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	}
}

func pushDiagnostic(writer io.Writer, uri string, diagnostics []lsp.Diagnostic) {
	notification := lsp.NewDiagnosticNotification(uri, diagnostics)
	writeResponse(writer, notification)
}

func writeResponse(writer io.Writer, msg any) {
	reply := lsp.EncodeMessage(msg)
	// logger.Info(reply)
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
