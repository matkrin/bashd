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

const (
	checkShebangWorkers = 8
	shebangBufferSize   = 256
)

type Document struct {
	Text         string
	SourcedFiles []Document
}

type Config struct {
	ExcludeDirs                  []string
	DiagnosticDebounceTime       time.Duration
	ShellCheckOptions            shellcheck.Options
	FormatOptions                FormatOptions
	WorkspaceDiagnosticsOnUpdate bool
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
	var candidates []string

	for _, folder := range s.WorkspaceFolders {
		dirpath, err := utils.UriToPath(folder.URI)
		if err != nil {
			continue
		}

		err = filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() && slices.Contains(s.Config.ExcludeDirs, d.Name()) {
				return fs.SkipDir
			}

			switch fileutil.CouldBeScript2(d) {
			case fileutil.ConfIsScript:
				shFiles = append(shFiles, path)

			case fileutil.ConfIfShebang:
				candidates = append(candidates, path)
			}

			return nil
		})
		_ = err
	}

	return append(shFiles, checkShebangFiles(candidates)...)
}

func checkShebangFiles(paths []string) []string {
	filesChan := make(chan string, checkShebangWorkers*2)
	resultsChan := make(chan string, len(paths))

	var wg sync.WaitGroup

	for range checkShebangWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for path := range filesChan {
				if hasShebangFile(path) {
					resultsChan <- path
				}
			}
		}()
	}

	for _, path := range paths {
		filesChan <- path
	}
	close(filesChan)

	wg.Wait()
	close(resultsChan)

	var shFiles []string
	for path := range resultsChan {
		shFiles = append(shFiles, path)
	}

	return shFiles
}

func hasShebangFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, shebangBufferSize))
	return err == nil && fileutil.HasShebang(data)
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
