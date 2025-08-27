package lsp

type DocumentColorRequest struct {
	Request
	Params DocumentColorParams `json:"params"`
}

type DocumentColorParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentColorResponse struct {
	Response
	Result []ColorInformation `json:"result"`
}

type ColorInformation struct {
	Range Range `json:"range"`
	Color Color `json:"color"`
}

type Color struct {
	Red   float32 `json:"red"`
	Green float32 `json:"green"`
	Blue  float32 `json:"blue"`
	Alpha float32 `json:"alpha"`
}
