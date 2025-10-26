package server

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/matkrin/bashd/internal/lsp"
)

var ansiRegex = regexp.MustCompile(`(\\e|\\033|\\x1b|\x1b)\[[0-9;]*m`)

func handleDocumentColor(request *lsp.DocumentColorRequest, state *State) *lsp.DocumentColorResponse {
	uri := request.Params.TextDocument.URI
	documentText := state.Documents[uri].Text

	colorInformation := extractColorsFromDocument(documentText)

	response := &lsp.DocumentColorResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
			ID:  &request.ID,
		},
		Result: colorInformation,
	}
	return response
}

func extractColorsFromDocument(documentText string) []lsp.ColorInformation {
	var results []lsp.ColorInformation
	lines := strings.Split(documentText, "\n")

	for lineNum, line := range lines {
		matches := ansiRegex.FindAllStringIndex(line, -1)

		for _, match := range matches {
			startByte := match[0]
			endByte := match[1]
			rawSeq := line[startByte:endByte]

			normalized := normalizeEscape(rawSeq)
			color, ok := parseColorFromSGR(normalized)
			if !ok {
				continue
			}

			startChar := utf8.RuneCountInString(line[:startByte])
			endChar := utf8.RuneCountInString(line[:endByte])

			results = append(results, lsp.ColorInformation{
				Range: lsp.NewRange(
					uint(lineNum),
					uint(startChar),
					uint(lineNum),
					uint(endChar),
				),
				Color: color,
			})
		}
	}

	return results
}

// Transform different escape notations to \x1b format
func normalizeEscape(raw string) string {
	raw = strings.ReplaceAll(raw, `\e`, "\x1b")
	raw = strings.ReplaceAll(raw, `\033`, "\x1b")
	raw = strings.ReplaceAll(raw, `\x1b`, "\x1b")
	return raw
}

func parseColorFromSGR(normalized string) (lsp.Color, bool) {
	if !strings.HasPrefix(normalized, "\x1b[") || !strings.HasSuffix(normalized, "m") {
		return lsp.Color{}, false
	}

	code := strings.TrimSuffix(strings.TrimPrefix(normalized, "\x1b["), "m")
	parts := strings.Split(code, ";")
	if len(parts) == 0 {
		return lsp.Color{}, false
	}

	// 256-color foreground: 38;5;<n>
	if len(parts) == 3 && parts[0] == "38" && parts[1] == "5" {
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			return lsp.Color{}, false
		}
		return xterm256ToRGB(idx), true
	}

	// 256-color background: 48;5;<n>
	if len(parts) == 3 && parts[0] == "48" && parts[1] == "5" {
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			return lsp.Color{}, false
		}
		return xterm256ToRGB(idx), true
	}

	// True color foreground: 38;2;<r>;<g>;<b>
	if len(parts) == 5 && parts[0] == "38" && parts[1] == "2" {
		r, _ := strconv.Atoi(parts[2])
		g, _ := strconv.Atoi(parts[3])
		b, _ := strconv.Atoi(parts[4])
		return lsp.Color{
			Red:   float32(r) / 255,
			Green: float32(g) / 255,
			Blue:  float32(b) / 255,
			Alpha: 1.0,
		}, true
	}

	// True color background: 48;2;<r>;<g>;<b>
	if len(parts) == 5 && parts[0] == "48" && parts[1] == "2" {
		r, _ := strconv.Atoi(parts[2])
		g, _ := strconv.Atoi(parts[3])
		b, _ := strconv.Atoi(parts[4])
		return lsp.Color{
			Red:   float32(r) / 255,
			Green: float32(g) / 255,
			Blue:  float32(b) / 255,
			Alpha: 1.0,
		}, true
	}

	// TODO: Maybe change that
	// Fall back to basic color approximations
	return escapeCodeToColor(normalized)
}

// Converts a color index (0-255) to RGB according to xterm palette
func xterm256ToRGB(index int) lsp.Color {
	if index < 16 {
		// TODO: Maybe change that
		// Basic colors (only approximinations, depend on user's terminal emulator)
		// 0-  7:  standard colors (as in ESC [ 30–37 m)
		// 8- 15:  high intensity colors (as in ESC [ 90–97 m)
		return approxColor(index)
	} else if index >= 16 && index < 232 {
		// 16-231:  6 × 6 × 6 cube (216 colors): 16 + 36 × r + 6 × g + b (0 ≤ r, g, b ≤ 5)
		index -= 16
		r := (index / 36) % 6
		g := (index / 6) % 6
		b := index % 6
		return lsp.Color{
			Red:   float32(r) / 5.0,
			Green: float32(g) / 5.0,
			Blue:  float32(b) / 5.0,
			Alpha: 1.0,
		}
	} else if index >= 232 && index <= 255 {
		// 232-255:  grayscale from dark to light in 24 steps
		gray := float32(index-232) / 23.0
		return lsp.Color{Red: gray, Green: gray, Blue: gray, Alpha: 1.0}
	}
	return lsp.Color{Red: 0, Green: 0, Blue: 0, Alpha: 0}
}

func escapeCodeToColor(seq string) (lsp.Color, bool) {
	colorMap := map[string]int{
		"\x1b[30m": 0,
		"\x1b[31m": 1,
		"\x1b[32m": 2,
		"\x1b[33m": 3,
		"\x1b[34m": 4,
		"\x1b[35m": 5,
		"\x1b[36m": 6,
		"\x1b[37m": 7,

		"\x1b[40m": 0,
		"\x1b[41m": 1,
		"\x1b[42m": 2,
		"\x1b[43m": 3,
		"\x1b[44m": 4,
		"\x1b[45m": 5,
		"\x1b[46m": 6,
		"\x1b[47m": 7,

		"\x1b[90m": 8,
		"\x1b[91m": 9,
		"\x1b[92m": 10,
		"\x1b[93m": 11,
		"\x1b[94m": 12,
		"\x1b[95m": 13,
		"\x1b[96m": 14,
		"\x1b[97m": 15,

		"\x1b[100m": 8,
		"\x1b[101m": 9,
		"\x1b[102m": 10,
		"\x1b[103m": 11,
		"\x1b[104m": 12,
		"\x1b[105m": 13,
		"\x1b[106m": 14,
		"\x1b[107m": 15,
	}

	approxColorIndex, ok := colorMap[seq]
	return approxColor(approxColorIndex), ok
}

func approxColor(index int) lsp.Color {
	approxColors := []lsp.Color{
		{Red: 0, Green: 0, Blue: 0, Alpha: 1},       // 0: black
		{Red: 0.8, Green: 0, Blue: 0, Alpha: 1},     // 1: red
		{Red: 0, Green: 0.8, Blue: 0, Alpha: 1},     // 2: green
		{Red: 0.8, Green: 0.8, Blue: 0, Alpha: 1},   // 3: yellow
		{Red: 0, Green: 0, Blue: 0.8, Alpha: 1},     // 4: blue
		{Red: 0.8, Green: 0, Blue: 0.8, Alpha: 1},   // 5: magenta
		{Red: 0, Green: 0.8, Blue: 0.8, Alpha: 1},   // 6: cyan
		{Red: 0.8, Green: 0.8, Blue: 0.8, Alpha: 1}, // 7: white

		{Red: 0.2, Green: 0.2, Blue: 0.2, Alpha: 1}, // 8: bright black
		{Red: 1, Green: 0, Blue: 0, Alpha: 1},       // 9: bright red
		{Red: 0, Green: 1, Blue: 0, Alpha: 1},       // 10: bright green
		{Red: 1, Green: 1, Blue: 0, Alpha: 1},       // 11: bright yellow
		{Red: 0.4, Green: 0.4, Blue: 1, Alpha: 1},   // 12: bright blue
		{Red: 1, Green: 0, Blue: 1, Alpha: 1},       // 13: bright magenta
		{Red: 0, Green: 1, Blue: 1, Alpha: 1},       // 14: bright cyan
		{Red: 1, Green: 1, Blue: 1, Alpha: 1},       // 15: bright white
	}
	return approxColors[index]
}
