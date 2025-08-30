package server

import (
	"testing"

	"github.com/matkrin/bashd/lsp"
)

func mockState(documentText string) *State {
	state := NewState()
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
			// TODO: update the condition below to compare got with tt.want.
			if got.Result.Location != tt.want.Result.Location {
				t.Errorf("handleDefinition() = %v, want %v", got.Result.Location, tt.want.Result.Location)
			}
		})
	}
}
