package lsp

type CodeActionRequest struct {
	Request
	Params CodeActionParams `json:"params"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic          `json:"diagnostics"`
	TiggerKind  CodeActionTriggerKind `json:"triggerKind"`
}

type CodeActionTriggerKind int

const (
	CodeActionTriggerInvoked CodeActionTriggerKind = iota + 1
	CodeActionTiggerAutomatic
)

type CodeActionResponse struct {
	Response
	Result []CodeAction `json:"result"`
}

type CodeAction struct {
	Title string `json:"title"`
	// Kind  CodeActionKind `json:"kind"`
	Edit WorkspaceEdit `json:"edit"`
}
