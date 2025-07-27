package lsp

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_hover
type HoverRequest struct {
	Request
	Params HoverParams `json:"params"`
}

type HoverParams struct {
	TextDocumentPositionParams
}

type HoverResponse struct {
	Response
	Result HoverResult `json:"result"`
}

type HoverResult struct {
	Contents MarkupContent `json:"contents"`
}
