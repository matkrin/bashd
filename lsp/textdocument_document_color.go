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
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha"`
}
