package lsp

type WorkspaceSymbolRequest struct {
	Request
	Params WorkspaceSymbolParams `json:"params"`
}

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

type WorkspaceSymbolResponse struct {
	Response
	Result []WorkspaceSymbol `json:"result"`
}

type WorkspaceSymbol struct {
	Name string     `json:"name"`
	Kind SymbolKind `json:"kind"`
	// Tags
	// ContainerName
	Location Location `json:"location"`
}

func NewWorkspaceSymbolResponse(
	id int,
	workspaceSymbols []WorkspaceSymbol,
) WorkspaceSymbolResponse {
	return WorkspaceSymbolResponse{
		Response: Response{
			RPC: RPC_VERSION,
			ID:  &id,
		},
		Result: workspaceSymbols,
	}
}
