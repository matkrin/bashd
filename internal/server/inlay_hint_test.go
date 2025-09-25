package server

import (
	"testing"

	"github.com/matkrin/bashd/internal/lsp"
)

func Test_extractInlayHints(t *testing.T) {
	input := `\e[0m
\e[1m
echo -e "\x1b[7m test \x1b[0m"
echo -e "\033[2m test \033[0m"
`

	tests := []lsp.InlayHint{
		{
			Position: lsp.Position{Line: 0, Character: 5},
			Label:    "reset",
			Kind:     lsp.InlayHintParameter,
		},
		{
			Position: lsp.Position{Line: 1, Character: 5},
			Label:    "bold",
			Kind:     lsp.InlayHintParameter,
		},
		{
			Position: lsp.Position{Line: 2, Character: 16},
			Label:    "invert",
			Kind:     lsp.InlayHintParameter,
		},
		{
			Position: lsp.Position{Line: 2, Character: 29},
			Label:    "reset",
			Kind:     lsp.InlayHintParameter,
		},
		{
			Position: lsp.Position{Line: 3, Character: 16},
			Label:    "dim",
			Kind:     lsp.InlayHintParameter,
		},
		{
			Position: lsp.Position{Line: 3, Character: 29},
			Label:    "reset",
			Kind:     lsp.InlayHintParameter,
		},
	}

	inlayHints := extractInlayHints(input)

	if len(tests) != len(inlayHints) {
		t.Errorf("Lenth of tests %d is not equal length of inlay hint results %d", len(tests), len(inlayHints))
	}

	for i, tt := range tests {
		inlayHint := inlayHints[i]
		if tt.Label != inlayHint.Label {
			t.Errorf("Expected label '%s' got '%s'", tt.Label, inlayHint.Label)
		}
		if tt.Position != inlayHint.Position {
			t.Errorf("Expected position %+v got %+v", tt.Position, inlayHint.Position)
		}

	}
}

func Test_parseInfosFromSGR(t *testing.T) {
	tests := []struct {
		normalized string
		want       string
	}{
		{"\x1b[0m", "reset"},
		{"\x1b[1m", "bold"},
		{"\x1b[7m", "invert"},
		{"\x1b[20m", "fraktur"},
		{"\x1b[28m", "reveal"},
		{"\x1b[51m", "framed"},
	}
	for _, tt := range tests {
		got := parseInfosFromSGR(tt.normalized)
		if got != tt.want {
			t.Errorf("parseInfosFromSGR() = %v, want %v", got, tt.want)
		}
	}
}
