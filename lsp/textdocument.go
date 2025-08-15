package lsp

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"lanuguageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type VersionTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type Position struct {
	Line      uint `json:"line"`
	Character uint `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func NewRange(startLine, startChar, endLine, endChar uint) Range {
	return Range{
		Start: Position{
			Line:      startLine,
			Character: startChar,
		},
		End: Position{
			Line:      endLine,
			Character: endChar,
		},
	}
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}
