package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/matkrin/bashd/lsp"
)

func HandleMessage(writer io.Writer, state *State, method string, contents []byte) {
	slog.Info("Received message", "method", method)

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}

		slog.Info("Connected to client",
			"name", request.Params.ClientInfo.Name,
			"version", request.Params.ClientInfo.Version,
		)

		state.WorkspaceFolders = request.Params.WorkspaceFolders
		slog.Info("Workspace folders set", "workerspaceFolders", state.WorkspaceFolders)

		workspaceDiagnostics := checkDiagnosticsWorkspace(state)
		for uri, diagnostics := range workspaceDiagnostics {
			pushDiagnostic(writer, uri, diagnostics)
		}

		msg := lsp.NewInitializeResponse(request.ID)
		writeResponse(writer, msg)

	case "shutdown":
		var request lsp.ShutdownRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}

		slog.Info("Shutdown server")
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
			slog.Error("Could not parse request", "method", method)
		}

		slog.Info("Opened", "URI", request.Params.TextDocument.URI)

		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)

		diagnostics := checkDiagnostics(request.Params.TextDocument.URI, state)
		pushDiagnostic(writer, request.Params.TextDocument.URI, diagnostics)

	case "textDocument/didChange":
		var request lsp.TextDocumentDidChangeNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}

		slog.Info("Changed", "URI", request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}

		diagnostics := checkDiagnostics(request.Params.TextDocument.URI, state)
		pushDiagnostic(writer, request.Params.TextDocument.URI, diagnostics)

	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleHover(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/definition":
		var request lsp.DefinitionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)

		}
		response := handleDefinition(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/references":
		var request lsp.ReferencesRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleReferences(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/completion":
		var request lsp.CompletionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleCompletion(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "completionItem/resolve":
		var request lsp.CompletionItemResolveRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleCompletionItemResolve(&request)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/documentSymbol":
		var request lsp.DocumentSymbolsRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleDocumentSymbol(&request, state)
		writeResponse(writer, response)

	case "textDocument/prepareRename":
		var request lsp.PrepareRenameRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handlePrepareRename(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/rename":
		var request lsp.RenameRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleRename(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "workspace/symbol":
		var request lsp.WorkspaceSymbolRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleWorkspaceSymbol(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/formatting":
		var request lsp.FormattingRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleFormatting(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/rangeFormatting":
		var request lsp.RangeFormattingRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleRangeFormatting(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/codeAction":
		var request lsp.CodeActionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleCodeAction(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/documentColor":
		var request lsp.DocumentColorRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleDocumentColor(&request, state)
		if response != nil {
			writeResponse(writer, response)
		}

	case "textDocument/inlayHint":
		var request lsp.InlayHintRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleInlayHint(&request, state)
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
