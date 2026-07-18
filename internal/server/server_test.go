package server

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matkrin/bashd/internal/lsp"
)

func mockState1(documentText string) *State {
	state := NewState(Config{ExcludeDirs: nil})
	state.SetDocument("file://workspace/test.sh", documentText)
	state.WorkspaceFolders = []lsp.WorkspaceFolder{
		{URI: "file://workspace", Name: "workspace"},
	}

	return &state
}

func TestHandleMessage(t *testing.T) {
	var testCases = []struct {
		name     string
		method   string
		contents []byte
	}{
		{
			name:     "initialize",
			method:   "initialize",
			contents: []byte(`{"id": 1, "params": {"clientInfo": {"name": "TestClient", "version": "1.0"}, "workspaceFolders": [{"uri": "file://workspace", "name": "workspace"}]}}`),
		},
		{
			name:     "initialize_missing_clientInfo",
			method:   "initialize",
			contents: []byte(`{"id": 1, "params": {"workspaceFolders": [{"uri": "file://workspace", "name": "workspace"}]}}`),
		},
		{
			name:     "initialize_empty_workspaceFolders",
			method:   "initialize",
			contents: []byte(`{"id": 1, "params": {"workspaceFolders": []}}`),
		},
		{
			name:     "initialize_missing_workspaceFolders",
			method:   "initialize",
			contents: []byte(`{"id": 1, "params": {"clientInfo": {"name": "TestClient", "version": "1.0"}}}`),
		},
		{
			name:     "shutdown",
			method:   "shutdown",
			contents: []byte(`{"id": 1}`),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.method, func(t *testing.T) {
			var buf bytes.Buffer
			writer := &buf

			state := mockState1(
				`#!/usr/bin/env bash

echo "hello world"

`,
			)

			server := NewServer("", "", *state, writer)
			server.HandleMessage(tt.method, tt.contents)
			server.Stop()

			switch tt.method {
			case "initialize":
				expectedIn := []string{`"jsonrpc":"2.0"`}
				response := writer.String()
				for _, exp := range expectedIn {
					if !strings.Contains(response, exp) {
						t.Errorf("'%s' failed. expected '%s' in '%s'", tt.name, exp, response)
					}
				}

			case "shutdown":
				expectedIn := []string{"Content-Length: 24", `"jsonrpc"`}
				response := writer.String()
				for _, exp := range expectedIn {
					if !strings.Contains(response, exp) {
						t.Errorf("'%s' failed. expected '%s' in '%s'", tt.name, exp, response)
					}
				}
			}
		})
	}
}
