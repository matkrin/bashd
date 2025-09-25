package server

import (
	"bytes"
	"log/slog"
	"strings"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"mvdan.cc/sh/v3/syntax"
)

// TODO: Cleaner end position
func handleFormatting(request *lsp.FormattingRequest, state *State) *lsp.FormattingResponse {
	slog.Info("FORMATTING", "params", request.Params)
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(document, uri)
	if err != nil {
		return nil
	}

	var indentWidth uint
	if request.Params.Options.InsertSpaces {
		indentWidth = request.Params.Options.TabSize
	} else {
		indentWidth = 0
	}
	indent := syntax.Indent(indentWidth)

	printer := syntax.NewPrinter(indent)

	buffer := bytes.NewBuffer([]byte{})
	printer.Print(buffer, fileAst.File)

	textedit := lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
			End: lsp.Position{
				Line:      99999,
				Character: 99999,
			},
		},
		NewText: buffer.String(),
	}

	response := lsp.FormattingResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: []lsp.TextEdit{textedit},
	}
	return &response
}

// TODO: Think about determining nodes from AST of entire document that are
// in range instead of slicing document and parse then. What's better?
func handleRangeFormatting(request *lsp.RangeFormattingRequest, state *State) *lsp.RangeFormattingResponse {
	slog.Info("RANGE-FORMATTING", "params", request.Params)
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text

	startLine := request.Params.Range.Start.Line
	endLine := request.Params.Range.End.Line

	lines := strings.Split(document, "\n")
	rangeLines := lines[startLine : endLine+1]
	rangeString := strings.Join(rangeLines, "\n")

	rangeAst, err := ast.ParseDocument(rangeString, uri)
	if err != nil {
		return nil
	}

	var indentWidth uint
	if request.Params.Options.InsertSpaces {
		indentWidth = request.Params.Options.TabSize
	} else {
		indentWidth = 0
	}
	indent := syntax.Indent(indentWidth)

	printer := syntax.NewPrinter(indent)

	buffer := bytes.NewBuffer([]byte{})
	err = printer.Print(buffer, rangeAst.File)
	if err != nil {
		return nil
	}

	newText := strings.TrimRight(buffer.String(), "\n")

	textEdit := lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      startLine,
				Character: request.Params.Range.Start.Character,
			},
			End: lsp.Position{
				Line:      endLine,
				Character: request.Params.Range.End.Character,
			},
		},
		NewText: newText,
	}

	response := lsp.RangeFormattingResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: []lsp.TextEdit{textEdit},
	}
	return &response
}
