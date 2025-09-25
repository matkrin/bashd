package lsp

type PrepareRenameRequest struct {
	Request
	Params PrepareRenameParams `json:"params"`
}

type PrepareRenameParams struct {
	TextDocumentPositionParams
}

type PrepareRenameResponse struct {
	Response
	Result Range `json:"result"`
}
