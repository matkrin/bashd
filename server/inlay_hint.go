package server

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/matkrin/bashd/lsp"
)

func handleInlayHint(request *lsp.InlayHintRequest, state *State) *lsp.InlayHintResponse {
	uri := request.Params.TextDocument.URI
	documentText := state.Documents[uri].Text

	inlayHints := extractInlayHints(documentText)

	response := &lsp.InlayHintResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},

		Result: inlayHints,
	}

	return response
}

func extractInlayHints(documentText string) []lsp.InlayHint {
	var results []lsp.InlayHint

	lines := strings.Split(documentText, "\n")

	for lineNum, line := range lines {
		matches := ansiRegex.FindAllStringIndex(line, -1)

		for _, match := range matches {
			startByte := match[0]
			endByte := match[1]
			rawSeq := line[startByte:endByte]

			normalized := normalizeEscape(rawSeq)

			inlayHintLabel := parseInfosFromSGR(normalized)
			if inlayHintLabel == "" {
				continue
			}

			// startChar := utf8.RuneCountInString(line[:startByte])
			endChar := utf8.RuneCountInString(line[:endByte])

			results = append(results, lsp.InlayHint{
				Position: lsp.Position{
					Line:      uint(lineNum),
					Character: uint(endChar),
				},
				Label:        inlayHintLabel,
				Kind:         lsp.InlayHintParameter,
				TextEdits:    nil,
				Tooltip:      nil,
				PaddingLeft:  true,
				PaddingRight: false,
			})
		}
	}

	return results
}

func parseInfosFromSGR(normalized string) string {
	if !strings.HasPrefix(normalized, "\x1b[") || !strings.HasSuffix(normalized, "m") {
		return ""
	}

	code := strings.TrimSuffix(strings.TrimPrefix(normalized, "\x1b["), "m")
	parts := strings.Split(code, ";")
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		n, err := strconv.Atoi(parts[0])
		if err != nil {
			return ""
		}
		sgr, ok := sgrTable[n]
		if !ok {
			return ""
		}
		return sgr
	}
	return ""
}

var sgrTable = map[int]string{
	0: "reset",
	1: "bold",
	2: "dim",
	3: "italic",
	4: "underline",
	5: "slow blink",
	6: "rapid blink",
	7: "invert",
	8: "hide",
	9: "strike",
	10: "primary font",
	20: "fraktur",
	21: "doubly underline",
	22: "normal intensity",
	23: "no italic",
	24: "not underline",
	25: "not blinking",
	26: "proportional spacing",
	27: "not reversed",
	28: "reveal",
	29: "not crossed out",
	51: "framed",
	52: "encircled",
	53: "overlined",
	58: "underline color",
}
