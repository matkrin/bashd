package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
)

// Converts a URI to an absolute path
func UriToPath(uri string) (string, error) {
	if !strings.HasPrefix(uri, "file://") {
		return "", fmt.Errorf("unsupported URI scheme")
	}
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}

// Converts an absolute path to and URI
func PathToURI(path string) string {
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "file://" + path
}

// Get the leading whitespace (spaces or tabs) from a line
func GetIndentation(line string) string {
    index := strings.IndexFunc(line, func(r rune) bool {
        return !unicode.IsSpace(r)
    })

    if index == -1 {
        // Line is all whitespace
        return line
    }
    return line[:index]
}
