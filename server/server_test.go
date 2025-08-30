package server_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matkrin/bashd/lsp"
	"github.com/matkrin/bashd/server"
)

func mockState() *server.State {
	state := server.NewState()
	state.OpenDocument("file://workspace/test.sh",
		`#!/usr/bin/env bash

echo "hello world"

`,
	)
	state.WorkspaceFolders = []lsp.WorkspaceFolder{
		{URI: "file://workspace", Name: "workspace"},
	}

	return &state
}

func TestHandleMessage(t *testing.T) {
	// Mocking the State and io.Writer

	var testCases = []struct {
		method   string
		contents []byte
	}{
		{
			method:   "initialize",
			contents: []byte(`{"params": {"clientInfo": {"name": "TestClient", "version": "1.0"}, "workspaceFolders": ["folder1"]}}`),
		},
		{
			method:   "shutdown",
			contents: []byte(`{"id": 1}`),
		},
		// Add more test cases for other methods
	}

	// Test loop for each case
	for _, tt := range testCases {
		t.Run(tt.method, func(t *testing.T) {
			// Mock the writer with a buffer
			var buf bytes.Buffer
			writer := &buf

			// Create the mock State (you can extend this mock as needed for your specific methods)
			state := mockState()

			// Call the HandleMessage function with the mock writer, state, method, and contents
			server.HandleMessage(writer, state, tt.method, tt.contents)

			// Example checks (depending on what you're testing)
			switch tt.method {
			case "initialize":
				expectedIn := []string{`"jsonrpc":"2.0"`}
				for _, exp := range expectedIn {
					if !strings.Contains(writer.String(), exp) {
						t.Errorf("'%s' failed. expected '%s' in '%s'", tt.method, exp, writer.String())
					}
				}

			case "shutdown":
				expectedIn := []string{"Content-Length: 38", `"jsonrpc"`, `"result":null`}
				for _, exp := range expectedIn {
					if !strings.Contains(writer.String(), exp) {
						t.Errorf("'%s' failed. expected '%s' in '%s'", tt.method, exp, writer.String())
					}
				}
			}
		})
	}
}
