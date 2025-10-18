package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/matkrin/bashd/internal/lsp"
)

type queuedMessage struct {
	method   string
	contents []byte
}

type Server struct {
	name            string
	version         string
	state           State
	writer          io.Writer
	messageQueue    chan queuedMessage
	wg              sync.WaitGroup
	diagnosticTimer *time.Timer
	mu              sync.Mutex
}

func NewServer(name, version string, state State, writer io.Writer) *Server {
	s := &Server{
		name:         name,
		version:      version,
		state:        state,
		writer:       writer,
		messageQueue: make(chan queuedMessage),
	}

	s.wg.Add(1)
	go s.run()

	return s
}

func (s *Server) run() {
	defer s.wg.Done()
	for msg := range s.messageQueue {
		s.dispatchMessage(msg.method, msg.contents)
	}
}

func (s *Server) Stop() {
	close(s.messageQueue)
	s.wg.Wait()
}

func (s *Server) HandleMessage(method string, contents []byte) {
	s.messageQueue <- queuedMessage{method: method, contents: contents}
}

func (s *Server) dispatchMessage(method string, contents []byte) {
	slog.Info("Received message", "method", method)
	var err error
	switch method {
	case "initialize":
		err = s.onInitialize(contents)
	case "shutdown":
		err = s.onShutdown(contents)
	case "exit":
		s.onExit()
	case "textDocument/didOpen":
		err = s.onTextDocumentDidOpen(contents)
	case "textDocument/didChange":
		err = s.onTextDocumentDidChange(contents)
	case "textDocument/hover":
		err = s.onTextDocumentHover(contents)
	case "textDocument/definition":
		err = s.onTextDocumentDefinition(contents)
	case "textDocument/references":
		err = s.onTextDocumentReferences(contents)
	case "textDocument/completion":
		err = s.onTextDocumentCompletion(contents)
	case "completionItem/resolve":
		err = s.onCompletionItemResolve(contents)
	case "textDocument/documentSymbol":
		err = s.onTextDocumentDocumentSymbol(contents)
	case "textDocument/prepareRename":
		err = s.onTextDocumentPerpareRename(contents)
	case "textDocument/rename":
		err = s.onTextDocumentRename(contents)
	case "workspace/symbol":
		err = s.onWorkspaceSymbol(contents)
	case "textDocument/formatting":
		err = s.onTextDocumentFormatting(contents)
	case "textDocument/rangeFormatting":
		err = s.onTextDocumentRangeFormatting(contents)
	case "textDocument/codeAction":
		err = s.onTextDocumentCodeAction(contents)
	case "textDocument/documentColor":
		err = s.onTextDocumentDocumentColor(contents)
	case "textDocument/inlayHint":
		err = s.onTextDocumentInlayHint(contents)
	}

	if err != nil {
		slog.Error("ERROR", "method", method, "err", err)
	}
}

func (s *Server) pushDiagnostic(uri string, diagnostics []lsp.Diagnostic) {
	notification := lsp.NewDiagnosticNotification(uri, diagnostics)
	s.writeResponse(notification)
}

func (s *Server) writeResponse(msg any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reply := lsp.EncodeMessage(msg)
	// logger.Info(reply)
	s.writer.Write([]byte(reply))
}

func (s *Server) onInitialize(contents []byte) error {
	var request lsp.InitializeRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}

	slog.Info("ClientInfo", "name", request.Params.ClientInfo.Name,
		"version", request.Params.ClientInfo.Version)

	s.state.WorkspaceFolders = request.Params.WorkspaceFolders
	slog.Info("Workspace folders set", "workerspaceFolders", s.state.WorkspaceFolders)

	workspaceDiagnostics := findDiagnosticsWorkspace(&s.state)
	for uri, diagnostics := range workspaceDiagnostics {
		s.pushDiagnostic(uri, diagnostics)
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
	s.writeResponse(msg)

	return nil
}

func (s *Server) onShutdown(contents []byte) error {
	var request lsp.ShutdownRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}

	slog.Info("Received shutdown request")
	s.state.ShutdownRequested = true

	response := lsp.ShutdownResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
			ID:  &request.ID,
		},
		Result: nil,
	}
	s.writeResponse(response)

	return nil
}

func (s *Server) onExit() {
	slog.Info("Exiting")
	if s.state.ShutdownRequested {
		os.Exit(0)
	} else {
		slog.Warn("Exiting without shutdown preceding shutdown request")
		os.Exit(1)
	}
}

func (s *Server) onTextDocumentDidOpen(contents []byte) error {
	var request lsp.DidOpenTextDocumentNotification
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}

	uri := request.Params.TextDocument.URI
	slog.Info("Opened document", "URI", uri)
	documentText := request.Params.TextDocument.Text
	s.state.SetDocument(uri, documentText)

	diagnostics := findDiagnostics(
		documentText,
		uri,
		s.state.EnvVars,
		s.state.Config.ShellCheckOptions,
	)
	s.pushDiagnostic(request.Params.TextDocument.URI, diagnostics)

	return nil
}

func (s *Server) onTextDocumentDidChange(contents []byte) error {
	var request lsp.TextDocumentDidChangeNotification
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}

	uri := request.Params.TextDocument.URI
	slog.Info("Changed document", "URI", uri)

	for _, change := range request.Params.ContentChanges {
		s.state.SetDocument(uri, change.Text)
	}
	documentText := s.state.Documents[uri].Text

	s.mu.Lock()
	if s.diagnosticTimer != nil {
		s.diagnosticTimer.Stop()
	}

	debounceTime := s.state.Config.DiagnosticDebounceTime
	s.diagnosticTimer = time.AfterFunc(debounceTime, func() {
		diagnostics := findDiagnostics(
			documentText,
			uri,
			s.state.EnvVars,
			s.state.Config.ShellCheckOptions,
		)
		s.pushDiagnostic(request.Params.TextDocument.URI, diagnostics)
	})
	s.mu.Unlock()

	return nil
}

func (s *Server) onTextDocumentHover(contents []byte) error {
	var request lsp.HoverRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleHover(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentDefinition(contents []byte) error {
	var request lsp.DefinitionRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleDefinition(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentReferences(contents []byte) error {
	var request lsp.ReferencesRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleReferences(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentCompletion(contents []byte) error {
	var request lsp.CompletionRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleCompletion(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onCompletionItemResolve(contents []byte) error {
	var request lsp.CompletionItemResolveRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleCompletionItemResolve(&request)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentDocumentSymbol(contents []byte) error {
	var request lsp.DocumentSymbolsRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleDocumentSymbol(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentPerpareRename(contents []byte) error {
	var request lsp.PrepareRenameRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handlePrepareRename(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentRename(contents []byte) error {
	var request lsp.RenameRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleRename(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onWorkspaceSymbol(contents []byte) error {
	var request lsp.WorkspaceSymbolRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleWorkspaceSymbol(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentFormatting(contents []byte) error {
	var request lsp.FormattingRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleFormatting(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentRangeFormatting(contents []byte) error {
	var request lsp.RangeFormattingRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleRangeFormatting(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentCodeAction(contents []byte) error {
	var request lsp.CodeActionRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleCodeAction(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentDocumentColor(contents []byte) error {
	var request lsp.DocumentColorRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleDocumentColor(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}

func (s *Server) onTextDocumentInlayHint(contents []byte) error {
	var request lsp.InlayHintRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		return errors.New("ERROR: Could not parse request")
	}
	response := handleInlayHint(&request, &s.state)
	if response != nil {
		s.writeResponse(response)
	}
	return nil
}
