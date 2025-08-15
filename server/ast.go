package server

import (
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

type Cursor struct {
	Line uint
	Col  uint
}

// Otherwise I will mess that up for sure. In the LSP 0-based, in the parser 1-based
func newCursor(lspLine, lspCol uint) Cursor {
	return Cursor{Line: lspLine + 1, Col: lspCol + 1}
}

func parseDocument(document string, documentName string) (*syntax.File, error) {
	reader := strings.NewReader(document)
	parser := syntax.NewParser(syntax.KeepComments(true))
	file, err := parser.Parse(reader, documentName)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func findNodeUnderCursor(file *syntax.File, cursor Cursor) syntax.Node {
	var found syntax.Node

	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			return true
		}
		start, end := node.Pos(), node.End()
		if isCursorInNode(cursor, start, end) {
			found = node
			// Continue walking to find deepest node containing cursor
			return true
		}
		return true
	})

	return found
}

func findAllSourcedFiles(
	file *syntax.File,
	env map[string]string,
	baseDir string,
	visited map[string]bool,
) []string {
	var sourcedFiles []string
	for _, sourcedFile := range findSourceStatments(file, env) {
		path := sourcedFile.Name
		resolved := path
		if !filepath.IsAbs(path) {
			resolved = filepath.Join(baseDir, path)
		}
		resolved = filepath.Clean(resolved)

		if visited[resolved] {
			continue
		}
		visited[resolved] = true

		sourcedFiles = append(sourcedFiles, resolved)

		// Recurse
		if content, err := os.ReadFile(resolved); err == nil {
			parser := syntax.NewParser()
			if parsed, err := parser.Parse(strings.NewReader(string(content)), ""); err == nil {
				subFiles := findAllSourcedFiles(parsed, env, filepath.Dir(resolved), visited)
				sourcedFiles = append(sourcedFiles, subFiles...)
			}
		}
	}
	return sourcedFiles
}

type SourcedFile struct {
	Name  string
	Start syntax.Pos
	End   syntax.Pos
}

func findSourceStatments(file *syntax.File, env map[string]string) []SourcedFile {
	sourcedFiles := []SourcedFile{}
	syntax.Walk(file, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}

		cmdName := extractWord(call.Args[0], env)
		if cmdName != "source" && cmdName != "." {
			return true
		}

		path := extractWord(call.Args[1], env)
		if path == "" {
			return true
		}
		sourcedFiles = append(sourcedFiles, SourcedFile{
			Name:  path,
			Start: node.Pos(),
			End:   node.End(),
		})
		return true
	})
	return sourcedFiles
}

func extractWord(word *syntax.Word, env map[string]string) string {
	var b strings.Builder
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(p.Value)

		case *syntax.ParamExp:
			val := env[p.Param.Value]
			b.WriteString(val)

		case *syntax.SglQuoted:
			b.WriteString(p.Value)

		case *syntax.DblQuoted:
			for _, qpart := range p.Parts {
				switch qp := qpart.(type) {
				case *syntax.Lit:
					b.WriteString(qp.Value)
				case *syntax.ParamExp:
					val := env[qp.Param.Value]
					b.WriteString(val)
				}
			}
		}
	}
	// Expand things like $HOME and ${VAR}
	return os.Expand(b.String(), func(key string) string {
		return env[key]
	})
}

func isCursorInNode(cursor Cursor, start, end syntax.Pos) bool {
	startLine := start.Line()
	startCol := start.Col()
	endLine := end.Line()
	endCol := end.Col()

	// Compare lines first
	if cursor.Line < startLine || cursor.Line > endLine {
		return false
	}

	if start.Line() == end.Line() {
		// Node is on a single line
		return cursor.Line == startLine &&
			cursor.Col >= startCol && cursor.Col <= endCol
	}

	// Multi-line node
	switch cursor.Line {
	case startLine:
		// On first line of node: col must be >= start.Col()
		return cursor.Col >= startCol
	case endLine:
		// On last line of node: col must be <= end.Col()
		return cursor.Col <= endCol
	default:
		// Any line in between start and end line is inside
		return true
	}
}

func extractIdentifier(node syntax.Node) string {
	switch n := node.(type) {
	case *syntax.Lit:
		return n.Value
	case *syntax.ParamExp:
		if n.Param != nil {
			return n.Param.Value
		}
	case *syntax.Word:
		if len(n.Parts) == 1 {
			switch p := n.Parts[0].(type) {
			case *syntax.Lit:
				return p.Value
			}
		}
	case *syntax.Assign:
		if n.Name != nil {
			return n.Name.Value
		}
	case *syntax.FuncDecl:
		if n.Name != nil {
			return n.Name.Value
		}
	}
	return ""
}
