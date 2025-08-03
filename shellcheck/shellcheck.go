package shellcheck

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os/exec"

	"github.com/matkrin/bashd/lsp"
)

type ShCheckOutput struct {
	Comments []Comment `json:"comments"`
}

type Comment struct {
	File      string `json:"file"`
	Line      uint   `json:"line"`
	EndLine   uint   `json:"endLine"`
	Column    uint   `json:"column"`
	EndColumn uint   `json:"endColumn"`
	Level     string `json:"level"`
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	// Fix ? `json:"fix"`
}

func (s *ShCheckOutput) ToDiagnostic() []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	for _, comment := range s.Comments {
		code := int(comment.Code)
		severity := comment.levelToSeverity()
		diagnostic := lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      int(comment.Line) - 1,
					Character: int(comment.Column) - 1,
				},
				End: lsp.Position{
					Line:      int(comment.EndLine) - 1,
					Character: int(comment.EndColumn) - 1,
				},
			},
			Severity: severity,
			Code:     &code,
			Source:   "shellcheck",
			Message:  comment.Message,
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}

func (c *Comment) levelToSeverity() lsp.DiagnosticSeverity {
	var severity lsp.DiagnosticSeverity
	switch c.Level {
	case "error":
		severity = lsp.DiagnosticError
	case "warning":
		severity = lsp.DiagnosticWarning
	case "info":
		severity = lsp.DiagnosticInformation
	default:
		slog.Warn("Unknown shellcheck level", "level", c.Level)
		severity = lsp.DiagnosticHint
	}
	return severity
}

func Run(filecontent string) (*ShCheckOutput, error) {
	cmd := exec.Command("shellcheck", "--format=json1", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.New("Could not acquire stdin")
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, filecontent)
	}()

	shOutput, err := cmd.CombinedOutput()
	if err != nil {
		// shellcheck exists with non-zero exit code if lints were founD
		// return nil, errors.New("Could not get stdout from shellcheck")
	}

	var output ShCheckOutput
	if err = json.Unmarshal(shOutput, &output); err != nil {
		return nil, errors.New("Could not unmarshal shellcheck output")
	}
	return &output, nil
}
