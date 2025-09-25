package server

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/matkrin/bashd/internal/lsp"
)

type Server struct {
	name    string
	version string
	state   State
	writer  io.Writer
}

func NewServer(name, version string, state State, writer io.Writer) *Server {
	return &Server{
		name:    name,
		version: version,
		state:   state,
		writer:  writer,
	}
}

func (s *Server) HandleMessage(method string, contents []byte) {
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

		s.state.WorkspaceFolders = request.Params.WorkspaceFolders
		slog.Info("Workspace folders set", "workerspaceFolders", s.state.WorkspaceFolders)

		workspaceDiagnostics := findDiagnosticsWorkspace(&s.state)
		for uri, diagnostics := range workspaceDiagnostics {
			pushDiagnostic(s.writer, uri, diagnostics)
		}

		capabilities := lsp.ServerCapabilities{
			TextDocumentSync:                1,
			HoverProvider:                   true,
			DefinitionProvider:              true,
			DeclarationProvider:             false,
			ReferencesProvider:              true,
			DocumentSymbolProvider:          true,
			WorkspaceSymbolProvider:         true,
			DocumentFormattingProvider:      true,
			DocumentRangeFormattingProvider: true,
			CodeActionProvider:              true,
			ColorProvider:                   true,
			InlayHintProvider:               true,
			RenameProvider: lsp.RenameOptions{
				PrepareProvider: true,
			},
			CompletionProvider: lsp.CompletionOptions{
				TriggerCharacters: []string{"$", "{"},
				ResolveProvider:   true,
			},
			DiagnosticProvider: lsp.DiagnosticOptions{
				Identifier:            nil,
				InterFileDependencies: false,
				WorkspaceDiagnostics:  false,
			},
		}
		info := lsp.ServerInfo{
			Name:    "bashd",
			Version: "0.1.0a1",
		}

		msg := lsp.NewInitializeResponse(request.ID, &capabilities, &info)
		writeResponse(s.writer, msg)

	case "shutdown":
		var request lsp.ShutdownRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}

		slog.Info("Shutdown server")
		response := lsp.ShutdownResponse{
			Response: lsp.Response{
				RPC: lsp.RPC_VERSION,
				ID:  &request.ID,
			},
			Result: nil,
		}
		writeResponse(s.writer, response)

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}

		uri := request.Params.TextDocument.URI
		slog.Info("Opened document", "URI", uri)
		documentText := request.Params.TextDocument.Text
		s.state.OpenDocument(uri, documentText)

		diagnostics := findDiagnostics(documentText, uri, s.state.EnvVars)
		pushDiagnostic(s.writer, request.Params.TextDocument.URI, diagnostics)

	case "textDocument/didChange":
		var request lsp.TextDocumentDidChangeNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}

		uri := request.Params.TextDocument.URI
		slog.Info("Changed document", "URI", uri)

		for _, change := range request.Params.ContentChanges {
			s.state.UpdateDocument(uri, change.Text)
		}

		documentText := s.state.Documents[uri].Text
		diagnostics := findDiagnostics(documentText, uri, s.state.EnvVars)
		pushDiagnostic(s.writer, request.Params.TextDocument.URI, diagnostics)

	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleHover(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/definition":
		var request lsp.DefinitionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)

		}
		response := handleDefinition(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/references":
		var request lsp.ReferencesRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleReferences(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/completion":
		var request lsp.CompletionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleCompletion(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "completionItem/resolve":
		var request lsp.CompletionItemResolveRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleCompletionItemResolve(&request)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/documentSymbol":
		var request lsp.DocumentSymbolsRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleDocumentSymbol(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/prepareRename":
		var request lsp.PrepareRenameRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handlePrepareRename(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/rename":
		var request lsp.RenameRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleRename(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "workspace/symbol":
		var request lsp.WorkspaceSymbolRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleWorkspaceSymbol(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/formatting":
		var request lsp.FormattingRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleFormatting(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/rangeFormatting":
		var request lsp.RangeFormattingRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleRangeFormatting(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/codeAction":
		var request lsp.CodeActionRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleCodeAction(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/documentColor":
		var request lsp.DocumentColorRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleDocumentColor(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
		}

	case "textDocument/inlayHint":
		var request lsp.InlayHintRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			slog.Error("Could not parse request", "method", method)
		}
		response := handleInlayHint(&request, &s.state)
		if response != nil {
			writeResponse(s.writer, response)
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
