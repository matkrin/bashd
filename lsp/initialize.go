package lsp

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initialize
type InitializeRequest struct {
	Request
	Params InitializeRequestParams `json:"params"`
}

type InitializeRequestParams struct {
	ProcessID        *int              `json:"processId"`
	ClientInfo       *ClientInfo       `json:"clientInfo"`
	Locale           string            `json:"locale"`
	RootPath         *string           `json:"rootPath"`
	RootURI          *string           `json:"rootUri"`
	Trace            *string           `json:"trace"`
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders"`
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
	TextDocumentSync       int               `json:"textDocumentSync"`
	DefinitionProvider     bool              `json:"definitionProvider"`
	HoverProvider          bool              `json:"hoverProvider"`
	DocumentSymbolProvider bool              `json:"documentSymbolProvider"`
	CompletionProvider     CompletionOptions `json:"completionProvider"`
	DiagnosticProvider     DiagnosticOptions `json:"diagnosticProvider"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
}

type DiagnosticOptions struct {
	Identifier            *string `json:"identifier"`
	InterFileDependencies bool    `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool    `json:"workspaceDiagnostics"`
}

func NewInitializeResponse(id int) InitializeResponse {
	return InitializeResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:       1,
				HoverProvider:          true,
				DefinitionProvider:     true,
				DocumentSymbolProvider: true,
				CompletionProvider: CompletionOptions{
					TriggerCharacters: []string{"$"},
				},
				DiagnosticProvider: DiagnosticOptions{
					Identifier:            nil,
					InterFileDependencies: false,
					WorkspaceDiagnostics:  false,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "bashd",
				Version: "0.1.0a1",
			},
		},
	}
}
