package lsp

type DidChangeConfigurationRequest struct {
	Request
	Params DidChangeConfigurationParams `json:"params"`
}

type DidChangeConfigurationParams struct {
	Settings any `json:"settings"`
}
