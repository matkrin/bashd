package shellcheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/utils"
)

type Options struct {
	Include       []string
	Exclude       []string
	OptionalLints []string // See `shellcheck --list-optional`
	Dialect       string   // sh, bash, dash, ksh, busybox
	Severity      string   // error, warning, info, style
}

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
	code := fmt.Sprintf("SC%d", c.Code)
	severity := c.levelToSeverity()
	codeActionAvailable := ""

	if c.Fix != nil {
		// codeActionAvailable = " \U0001F4A1"
		codeActionAvailable = " "
	}
	message := fmt.Sprintf("%s%s", c.Message, codeActionAvailable)

	return lsp.Diagnostic{
		Range: lsp.NewRange(
			c.Line-1,
			c.Column-1,
			c.EndLine-1,
			c.EndColumn-1,
		),
		Severity: severity,
		Code:     &code,
		Source:   "shellcheck",
		Message:  message,
	}
}

func (c *Comment) ToCodeActionFixLint(uri string) *lsp.CodeAction {
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

func (c *Comment) ToCodeActionIgnore(uri, documentText string, diagnosticRange *lsp.Range) *lsp.CodeAction {
	startLine := diagnosticRange.Start.Line
	diagnosticLine := strings.Split(documentText, "\n")[startLine]
	indentation := utils.GetIndentation(diagnosticLine)
	textEdits := []lsp.TextEdit{
		{
			Range:   lsp.NewRange(startLine, 0, startLine, 0),
			NewText: fmt.Sprintf("%s# shellcheck disable=SC%d\n", indentation, c.Code),
		},
	}
	action := &lsp.CodeAction{
		Title: fmt.Sprintf("Add ignore comment for lint SC%d", c.Code),
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
			Range: lsp.NewRange(
				rep.Line-1,
				rep.Column-1,
				rep.Line-1,
				rep.EndColumn-1,
			),
			NewText: rep.Replacement,
		}
		textEdits = append(textEdits, textEdit)

	}
	return textEdits
}

func Run(filecontent string, options Options) (*ShellCheckResult, error) {
	optionalLints := options.OptionalLints

	args := []string{
		"--format=json1",
		"--external-sources",
	}
	if len(optionalLints) != 0 {
		args = append(args, fmt.Sprintf("--enable=%s", strings.Join(optionalLints, ",")))
	}
	if len(options.Include) != 0 {
		args = append(args, fmt.Sprintf("--include=%s", strings.Join(options.Include, ",")))
	}
	if len(options.Exclude) != 0 {
		args = append(args, fmt.Sprintf("--exclude=%s", strings.Join(options.Exclude, ",")))
	}
	if options.Dialect != "" {
		args = append(args, fmt.Sprintf("--shell=%s", options.Dialect))
	}
	if options.Severity != "" {
		args = append(args, fmt.Sprintf("--severity=%s", options.Severity))
	}
	args = append(args, "-")
	cmd := exec.Command( "shellcheck", args...)
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
	slog.Info("SHELLCHECK", "shOutput", shOutput)

	var output ShellCheckResult
	if err = json.Unmarshal(shOutput, &output); err != nil {
		return nil, errors.New("Could not unmarshal shellcheck output")
	}
	return &output, nil
}
