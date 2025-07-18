package lsp

type RenameRequest struct {
	Request
	Params RenameParams `json:"params"`
}

type RenameParams struct {
	TextDocumentPositionParams
	NewName string `json:"newName"`
}

type RenameResponse struct {
	Response
	Result *WorkspaceEdit `json:"result"`
}

type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
	// DocumentChanges []TextDocumentEdit `json:"documentChangs"`
	// ChangeAnnotations `json:"changeAnnotations"`
}

type TextEdit struct {
	Range
	NewText string `json:"newText"`
}
