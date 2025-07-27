package lsp

type FormattingRequest struct {
	Request
	Params FormattingParams `json:"params"`
}

type FormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type FormattingOptions struct {
	TabSize                uint `json:"tabSize"`
	InsertSpaces           bool `json:"insertSpaces"`
	TrimTrailingWhiteSpace bool `json:"trimTrailingWhiteSpace"`
	InsertFinalNewLine     bool `json:"insertFinalNewLine"`
	TrimFinalNewLines      bool `json:"trimFinalNewLines"`
}

type FormattingResponse struct {
	Response
	Result []TextEdit `json:"result"`
}
