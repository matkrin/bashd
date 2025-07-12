package lsp

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_references
type ReferencesRequest struct {
	Request
	Params ReferencesParams `json:"params"`
}

type ReferencesParams struct {
	TextDocumentPositionParams
	Context ReferencesContext `json:"context"`
}

type ReferencesContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

type ReferencesResponse struct {
	Response
	Result []Location `json:"result"`
}
