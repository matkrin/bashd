package server

import (
	"os"
	"strings"

	"github.com/matkrin/bashd/logger"
	"github.com/matkrin/bashd/lsp"
)

type Document struct {
	Text         string
	SourcedFiles []Document
}

type State struct {
	Documents        map[string]Document
	EnvVars          map[string]string
	WorkspaceFolders []lsp.WorkspaceFolder
	PathItems        []string
}

func NewState() State {
	envVars := getEnvVars()
	pathItems := getPathItems(envVars)

	return State{
		Documents: map[string]Document{},
		EnvVars:   envVars,
		PathItems: pathItems,
	}
}

func (s *State) OpenDocument(uri, text string) {
	s.Documents[uri] = Document{
		Text:         text,
		SourcedFiles: []Document{},
	}
}
func (s *State) UpdateDocument(uri, text string) {
	s.Documents[uri] = Document{
		Text:         text,
		SourcedFiles: []Document{},
	}
}

func getEnvVars() map[string]string {
	env := os.Environ()
	envVars := make(map[string]string)

	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		key := pair[0]
		value := ""
		if len(pair) > 1 {
			value = pair[1]
		}
		envVars[key] = value
	}

	return envVars
}

func getPathItems(envVars map[string]string) []string {
	pathStr := envVars["PATH"]
	pathItems := []string{}
	for pathPart := range strings.SplitSeq(pathStr, ":") {
		entries, err := os.ReadDir(pathPart)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				logger.Errorf("Error getting the info of %s", entry.Name())
				continue
			}

			perm := info.Mode().Perm()
			// Check if file is executable
			if perm&0111 != 0 {
				pathItems = append(pathItems, entry.Name())
			}
		}
	}
	return pathItems
}
