package lsp

type DeclarationRequest struct {
	Request
	Params DefinitionParams `json:"params"`
}

type DeclarationParams struct {
	TextDocumentPositionParams
}

type DeclarationResponse struct {
	Response
	Result DeclarationResult `json:"result"`
}

type DeclarationResult struct {
	Location
}

func NewDeclarationResponse(
	id int,
	documentURI string,
	startLine, startChar, endLine, endChar uint,
) *DeclarationResponse {
	return &DeclarationResponse{
		Response: Response{
			RPC: RPC_VERSION,
			ID:  &id,
		},
		Result: DeclarationResult{
			Location: Location{
				URI: documentURI,
				Range: Range{
					Start: Position{
						Line:      startLine,
						Character: startChar,
					},
					End: Position{
						Line:      endLine,
						Character: endChar,
					},
				},
			},
		},
	}
}
