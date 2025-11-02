package server

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/matkrin/bashd/internal/lsp"
	"github.com/matkrin/bashd/internal/shellcheck"
	"github.com/matkrin/bashd/internal/utils"
	"mvdan.cc/sh/v3/fileutil"
)

type Document struct {
	Text         string
	SourcedFiles []Document
}

type Config struct {
	ExcludeDirs            []string
	DiagnosticDebounceTime time.Duration
	ShellCheckOptions      shellcheck.Options
	FormatOptions          FormatOptions
}

type FormatOptions struct {
	BinaryNextLine bool
	CaseIndent     bool
	SpaceRedirects bool
	FuncNextLine   bool
}

type State struct {
	Documents         map[string]Document
	EnvVars           map[string]string
	WorkspaceFolders  []lsp.WorkspaceFolder
	PathItems         []string
	Config            Config
	ShutdownRequested bool
}

func NewState(config Config) State {
	envVars := getEnvVars()
	pathItems := []string{}
	if pathStr, ok := envVars["PATH"]; ok {
		pathItems = getPathItems(pathStr)
	}

	return State{
		Documents:         make(map[string]Document),
		EnvVars:           envVars,
		PathItems:         pathItems,
		Config:            config,
		ShutdownRequested: false,
	}
}

func (s *State) SetDocument(uri, documentText string) {
	s.Documents[uri] = Document{
		Text:         documentText,
		SourcedFiles: []Document{},
	}
}

// Find sh-files and return their filepaths
func (s *State) WorkspaceShFiles() []string {
	var shFiles []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, folder := range s.WorkspaceFolders {
		dirpath, err := utils.UriToPath(folder.URI)
		if err != nil {
			continue
		}

		filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && slices.Contains(s.Config.ExcludeDirs, d.Name()) {
				return fs.SkipDir
			}

			if fileutil.CouldBeScript2(d) == fileutil.ConfIsScript {
				mu.Lock()
				shFiles = append(shFiles, path)
				mu.Unlock()
			} else if fileutil.CouldBeScript2(d) == fileutil.ConfIfShebang {
				wg.Add(1)
				go func(path string) {
					defer wg.Done()
					file, err := os.Open(path)
					if err != nil {
						return
					}
					defer file.Close()
					// TODO: Maybe read not all bytes
					data, err := io.ReadAll(file)
					if err != nil {
						return
					}
					if fileutil.HasShebang(data) {
						mu.Lock()
						shFiles = append(shFiles, path)
						mu.Unlock()
					}
				}(path)
			}
			return nil
		})
	}

	wg.Wait()
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

func getPathItems(pathStr string) []string {
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
				slog.Error("Error getting file info", "file", entry.Name())
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
