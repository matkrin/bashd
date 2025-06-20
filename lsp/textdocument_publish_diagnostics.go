package lsp

type DiagnosticNotification struct {
	Notification
	Params PublishDiagnosticsParams `json:"params"`
}

func NewDiagnosticNotification(uri string, diagnostics []Diagnostic) DiagnosticNotification {
	return DiagnosticNotification{
		Notification: Notification{
			RPC:    "2.0",
			Method: "textDocument/publishDiagnostics",
		},
		Params: PublishDiagnosticsParams{
			URI:         uri,
			Version:     nil,
			Diagnostics: diagnostics,
		},
	}
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Code     *int               `json:"code"`
	// CodeDescription
	Source  string `json:"source"`
	Message string `json:"message"`
	// Tags
	// RelatedInformation
	// Data
}

type DiagnosticSeverity int

const (
	DiagnosticError DiagnosticSeverity = iota + 1
	DiagnosticWarning
	DiagnosticInformation
	DiagnosticHint
)
