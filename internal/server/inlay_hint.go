package server

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/matkrin/bashd/internal/lsp"
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
	30: "black fg",
	31: "red fg",
	32: "green fg",
	33: "yellow fg",
	34: "blue fg",
	35: "magenta fg",
	36: "cyan fg",
	37: "white fg",
	40: "black bg",
	41: "red bg",
	42: "green bg",
	43: "yellow bg",
	44: "blue bg",
	45: "magenta bg",
	46: "cyan bg",
	47: "white bg",
	51: "framed",
	52: "encircled",
	53: "overlined",
	58: "underline color",
	90: "black bright fg",
	91: "red bright fg",
	92: "green bright fg",
	93: "yellow bright fg",
	94: "blue bright fg",
	95: "magenta bright fg",
	96: "cyan bright fg",
	97: "white bright fg",
	100: "black bright bg",
	101: "red bright bg",
	102: "green bright bg",
	103: "yellow bright bg",
	104: "blue bright bg",
	105: "magenta bright bg",
	106: "cyan bright bg",
	107: "white bright bg",
}
