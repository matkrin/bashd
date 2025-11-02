package lsp

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initialize
type InitializeRequest struct {
	Request
	Params InitializeRequestParams `json:"params"`
}

type InitializeRequestParams struct {
	ProcessID             *int              `json:"processId"`
	ClientInfo            *ClientInfo       `json:"clientInfo"`
	Locale                string            `json:"locale"`
	RootPath              *string           `json:"rootPath"`
	RootURI               *string           `json:"rootUri"`
	Trace                 *string           `json:"trace"`
	WorkspaceFolders      []WorkspaceFolder `json:"workspaceFolders"`
	InitializationOptions *any              `json:"initializationOptions"`
	// Capabilities ClientCapabilities `json:"capabilities"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

type InitializeResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initializeResult
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	TextDocumentSync                int               `json:"textDocumentSync"`
	DefinitionProvider              bool              `json:"definitionProvider"`
	DeclarationProvider             bool              `json:"declarationProvider"`
	ReferencesProvider              bool              `json:"referencesProvider"`
	HoverProvider                   bool              `json:"hoverProvider"`
	DocumentSymbolProvider          bool              `json:"documentSymbolProvider"`
	WorkspaceSymbolProvider         bool              `json:"workspaceSymbolProvider"`
	DocumentFormattingProvider      bool              `json:"documentFormattingProvider"`
	DocumentRangeFormattingProvider bool              `json:"documentRangeFormattingProvider"`
	CodeActionProvider              bool              `json:"codeActionProvider"`
	ColorProvider                   bool              `json:"colorProvider"`
	InlayHintProvider               bool              `json:"inlayHintProvider"`
	RenameProvider                  RenameOptions     `json:"renameProvider"`
	CompletionProvider              CompletionOptions `json:"completionProvider"`
	DiagnosticProvider              DiagnosticOptions `json:"diagnosticProvider"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
	ResolveProvider   bool     `json:"resolveProvider"`
}

type DiagnosticOptions struct {
	Identifier            *string `json:"identifier"`
	InterFileDependencies bool    `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool    `json:"workspaceDiagnostics"`
}

type RenameOptions struct {
	PrepareProvider bool `json:"prepareProvider"`
}

func NewInitializeResponse(id int, capabilities *ServerCapabilities, info *ServerInfo) InitializeResponse {
	return InitializeResponse{
		Response: Response{
			RPC: RPC_VERSION,
			ID:  &id,
		},
		Result: InitializeResult{
			Capabilities: *capabilities,
			ServerInfo:   *info,
		},
	}
}
