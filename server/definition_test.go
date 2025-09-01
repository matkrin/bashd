package server

import (
	"testing"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

func mockState(documentText string) *State {
	state := NewState(Config{ExcludeDirs: nil})
	state.OpenDocument("file://workspace/test.sh", documentText)
	state.WorkspaceFolders = []lsp.WorkspaceFolder{
		{URI: "file://workspace", Name: "workspace"},
	}

	return &state
}

func mockRequest(position lsp.Position) *lsp.DefinitionRequest {
	return &lsp.DefinitionRequest{
		Request: lsp.Request{
			RPC:    "2.0",
			ID:     0,
			Method: "textdocument/definition",
		},
		Params: lsp.DefinitionParams{
			TextDocumentPositionParams: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: "file://workspace/test.sh",
				},
				Position: position},
		},
	}
}

func mockResponse(_range lsp.Range) *lsp.DefinitionResponse {
	id := 0
	return &lsp.DefinitionResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: lsp.DefinitionResult{
			Location: lsp.Location{
				URI:   "file://workspace/test.sh",
				Range: _range,
			},
		},
	}
}

func Test_handleDefinition(t *testing.T) {
	state := mockState(`#!/usr/bin/env bash

a="test"
echo "$a"

foo() {
	echo "bar"
}

foo
`,
	)

	tests := []struct {
		name    string
		request *lsp.DefinitionRequest
		want    *lsp.DefinitionResponse
	}{
		{
			"Variable",
			mockRequest(lsp.Position{Line: 3, Character: 7}),
			mockResponse(lsp.NewRange(2, 0, 2, 1)),
		},
		{
			"Function",
			mockRequest(lsp.Position{Line: 9, Character: 0}),
			mockResponse(lsp.NewRange(5, 0, 5, 3)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleDefinition(tt.request, state)
			if got.Result.Location != tt.want.Result.Location {
				t.Errorf("handleDefinition() = %v, want %v", got.Result.Location, tt.want.Result.Location)
			}
		})
	}
}

func Test_defNodes(t *testing.T) {
	input := `#!/usr/bin/env bash

a="test"
echo "$a"

foo() {
	echo "bar"
}

foo
`
	fileAst, _ := parseDocument(input, "test.sh")
	defNodes := defNodes(fileAst)
	if len(defNodes) != 2 {
		t.Errorf("length of defNodes not 2; got %v", len(defNodes))
	}
}

func Test_findDefInFile(t *testing.T) {
	input := `#!/usr/bin/env bash

a="test"
echo "$a"

foo() {
	echo "bar"
}

foo
`
	tests := []struct {
		name   string
		cursor Cursor
		want   DefNode
	}{
		{
			"Variable", newCursor(3, 7), DefNode{
				Name:  "a",
				Start: syntax.NewPos(21, 3, 1),
				End:   syntax.NewPos(22, 3, 2),
			},
		},
		{
			"Function", newCursor(9, 0), DefNode{
				Name:  "foo",
				Start: syntax.NewPos(41, 6, 1),
				End:   syntax.NewPos(44, 6, 4),
			},
		},
	}

	fileAst, _ := parseDocument(input, "test.sh")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursorNode := findNodeUnderCursor(fileAst, tt.cursor)
			got := findDefInFile(cursorNode, fileAst)
			if (*got).Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", (*got).Name, tt.want.Name)
			}
			if (*got).Start != tt.want.Start {
				t.Errorf("Start = %#v, want %#v", (*got).Start, tt.want.Start)
			}
			if (*got).End != tt.want.End {
				t.Errorf("End = %#v, want %#v", (*got).End, tt.want.End)
			}
		})
	}

}
