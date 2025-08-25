package lsp

type RangeFormattingRequest struct {
	Request
	Params RangeFormattingParams `json:"params"`
}

type RangeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Options      FormattingOptions      `json:"options"`
}

type RangeFormattingResponse struct {
	Response
	Result []TextEdit `json:"result"`
}
