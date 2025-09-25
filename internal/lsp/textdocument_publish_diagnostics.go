package lsp

// type DiagnosticsRequest struct {
// 	Request
// 	Params DiagnosticsParams `json:"params"`
// }
//
// type DiagnosticsParams struct {
// 	TextDocument     TextDocumentIdentifier `json:"textDocument"`
// 	Identifier       *string                `json:"identifier"`
// 	PreviousResultID *string                `json:"previousResultId"`
// }
//
// type DiagnosticResponse struct {
// 	Response
// 	Result DocumentDiagnosticReport `json:"result"`
// }
//
// type DocumentDiagnosticReport struct  {
//
// }

type DiagnosticNotification struct {
	Notification
	Params PublishDiagnosticsParams `json:"params"`
}

func NewDiagnosticNotification(uri string, diagnostics []Diagnostic) DiagnosticNotification {
	return DiagnosticNotification{
		Notification: Notification{
			RPC:    RPC_VERSION,
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
	Code     *string            `json:"code"`
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
