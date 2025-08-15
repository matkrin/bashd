package shellcheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/matkrin/bashd/lsp"
)

// https://github.com/koalaman/shellcheck/wiki/Integration
type ShellCheckResult struct {
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
	Fix       *Fix   `json:"fix"`
}

type Fix struct {
	Replacements []struct {
		Line           uint   `json:"line"`
		Column         uint   `json:"column"`
		EndColumn      uint   `json:"endColumn"`
		InsertionPoint string `json:"insertionPoint"`
		Replacement    string `json:"replacement"`
		Precedence     uint   `json:"precedence"`
	} `json:"replacements"`
}

func (s *ShellCheckResult) ToDiagnostics() []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	for _, comment := range s.Comments {
		diagnostics = append(diagnostics, comment.ToDiagnostic())
	}
	return diagnostics
}

func (s *ShellCheckResult) ToCodeActionFlat(uri string) lsp.CodeAction {
	textEdits := []lsp.TextEdit{}
	for _, comment := range s.Comments {
		if comment.Fix != nil {
			textEdits = append(textEdits, comment.Fix.toTextEdits()...)
		}
	}
	action := lsp.CodeAction{
		Title: "Fix all auto-fixable lints",
		Edit: lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: textEdits,
			},
		},
	}
	return action
}

func (s *ShellCheckResult) ContainsFixable() bool {
	for _, comment := range s.Comments {
		if comment.Fix != nil {
			return true
		}
	}
	return false
}

func (c *Comment) ToDiagnostic() lsp.Diagnostic {
	code := int(c.Code)
	severity := c.levelToSeverity()
	codeActionAvailable := ""

	if c.Fix != nil {
		// codeActionAvailable = " \U0001F4A1"
		codeActionAvailable = " "
	}
	message := fmt.Sprintf("%s%s", c.Message, codeActionAvailable)

	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      c.Line - 1,
				Character: c.Column - 1,
			},
			End: lsp.Position{
				Line:      c.EndLine - 1,
				Character: c.EndColumn - 1,
			},
		},
		Severity: severity,
		Code:     &code,
		Source:   "shellcheck",
		Message:  message,
	}
}

func (c *Comment) ToCodeAction(uri string) *lsp.CodeAction {
	if c.Fix == nil {
		return nil
	}

	textEdits := c.Fix.toTextEdits()
	action := &lsp.CodeAction{
		Title: fmt.Sprintf("Fix shellcheck lint %d", c.Code),
		Edit: lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: textEdits,
			},
		},
	}
	return action
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
	case "style":
		severity = lsp.DiagnosticHint
	default:
		slog.Warn("Unknown shellcheck level", "level", c.Level)
		severity = lsp.DiagnosticHint
	}
	return severity
}

func (f *Fix) toTextEdits() []lsp.TextEdit {
	textEdits := []lsp.TextEdit{}
	for _, rep := range f.Replacements {
		textEdit := lsp.TextEdit{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      rep.Line - 1,
					Character: rep.Column - 1,
				},
				End: lsp.Position{
					Line:      rep.Line - 1,
					Character: rep.EndColumn - 1,
				},
			},
			NewText: rep.Replacement,
		}
		textEdits = append(textEdits, textEdit)

	}
	return textEdits
}

func Run(filecontent string) (*ShellCheckResult, error) {
	optionalLints := []string{
		"add-default-case",
		"require-double-brackets",
	}
	cmd := exec.Command(
		"shellcheck",
		"--format=json1",
		"--external-sources",
		"--enable=add-default-case,require-double-brackets",
		fmt.Sprintf("--enable=%s", strings.Join(optionalLints, ",")),
		"-",
	)
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
		// https://github.com/koalaman/shellcheck/wiki/Integration#exit-codes
		// return nil, errors.New("Could not get stdout from shellcheck")
	}

	var output ShellCheckResult
	if err = json.Unmarshal(shOutput, &output); err != nil {
		return nil, errors.New("Could not unmarshal shellcheck output")
	}
	return &output, nil
}
