package lsp

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_definition
type DefinitionRequest struct {
	Request
	Params DefinitionParams `json:"params"`
}

type DefinitionParams struct {
	TextDocumentPositionParams
}

type DefinitionResponse struct {
	Response
	Result DefinitionResult `json:"result"`
}

type DefinitionResult struct {
	Location
}

func NewDefinitionResponse(
	id int,
	documentURI string,
	startLine, startChar, endLine, endChar uint,
) DefinitionResponse {
	return DefinitionResponse{
		Response: Response{
			RPC: RPC_VERSION,
			ID:  &id,
		},
		Result: DefinitionResult{
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
