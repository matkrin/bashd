package server

import (
	"testing"

	"github.com/matkrin/bashd/ast"
	"github.com/matkrin/bashd/lsp"
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
	fileAst, _ := ast.ParseDocument(input, "test.sh")
	defNodes := fileAst.DefNodes()
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
		cursor ast.Cursor
		want   ast.DefNode
	}{
		{
			"Variable", ast.NewCursor(3, 7), ast.DefNode{
				Name:  "a",
				StartLine: 3,
				StartChar: 1,
				EndLine: 3,
				EndChar: 2,
			},
		},
		{
			"Function", ast.NewCursor(9, 0), ast.DefNode{
				Name:  "foo",
				StartLine: 6,
				StartChar: 1,
				EndLine: 6,
				EndChar: 4,
			},
		},
	}

	fileAst, _ := ast.ParseDocument(input, "test.sh")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursorNode := fileAst.FindNodeUnderCursor(tt.cursor)
			got := fileAst.FindDefInFile(cursorNode)
			if (*got).Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", (*got).Name, tt.want.Name)
			}
			if (*got).StartLine != tt.want.StartLine {
				t.Errorf("StartLine = %#v, want %#v", (*got).StartLine, tt.want.StartLine)
			}
			if (*got).StartChar != tt.want.StartChar {
				t.Errorf("StartChar = %#v, want %#v", (*got).StartChar, tt.want.StartChar)
			}
			if (*got).EndLine != tt.want.EndLine {
				t.Errorf("EndLine = %#v, want %#v", (*got).EndLine, tt.want.EndLine)
			}
			if (*got).EndChar != tt.want.EndChar {
				t.Errorf("EndChar = %#v, want %#v", (*got).EndChar, tt.want.EndChar)
			}
		})
	}

}
