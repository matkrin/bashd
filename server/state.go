package server

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/fileutil"
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

var excludeDirs = [...]string{
	".git",
	".venv",
	"node_modules",
}

func (s *State) WorkspaceShFiles() []string {
	shFiles := []string{}
	for _, folder := range s.WorkspaceFolders {
		filepath.WalkDir(folder.Name, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && slices.Contains(excludeDirs[:], d.Name()) {
				return fs.SkipDir
			}
			fileext := filepath.Ext(path)
			if fileext == ".sh" || fileutil.CouldBeScript2(d) == fileutil.ConfIfShebang {
				shFiles = append(shFiles, path)
			}
			return nil
		})
	}

	return shFiles
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
				slog.Error("Error getting file into", "file", entry.Name())
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
