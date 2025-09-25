package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

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

func PathToURI(path string) string {
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	return uri.String()
}
