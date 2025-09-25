package lsp

type InlayHintRequest struct {
	Request
	Params InlayHintParams `json:"params"`
}

type InlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type InlayHintResponse struct {
	Response
	Result []InlayHint `json:"result"`
}

type InlayHint struct {
	Position     Position      `json:"position"`
	Label        string        `json:"label"`
	Kind         InlayHintKind `json:"kind"`
	TextEdits    *[]TextEdit   `json:"textEdits"`
	Tooltip      *string       `json:"tooltip"`
	PaddingLeft  bool          `json:"paddingLeft"`
	PaddingRight bool          `json:"paddingRight"`
}

type InlayHintKind int

const (
	InlayHintType InlayHintKind = iota + 1
	InlayHintParameter
)
