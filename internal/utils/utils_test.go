package utils

import (
	"testing"
)

func TestUriToPath(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid unix path",
			uri:     "file:///tmp/test.txt",
			want:    "/tmp/test.txt",
			wantErr: false,
		},
		{
			name:    "valid path with spaces",
			uri:     "file:///home/user/My%20File.txt",
			want:    "/home/user/My File.txt",
			wantErr: false,
		},
		{
			name:    "unsupported scheme",
			uri:     "https://example.com/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid uri",
			uri:     "file://%zz",
			wantErr: true,
		},
		{
			name:    "empty uri",
			uri:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UriToPath(tt.uri)

			if (err != nil) != tt.wantErr {
				t.Fatalf("UriToPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("UriToPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPathToURI(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "absolute unix path",
			path: "/tmp/test.txt",
			want: "file:///tmp/test.txt",
		},
		{
			name: "relative path",
			path: "tmp/test.txt",
			want: "file:///tmp/test.txt",
		},
		{
			name: "empty path",
			path: "",
			want: "file:///",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathToURI(tt.path)

			if got != tt.want {
				t.Errorf("PathToURI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetIndentation(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "spaces",
			line: "    hello",
			want: "    ",
		},
		{
			name: "tabs",
			line: "\t\thello",
			want: "\t\t",
		},
		{
			name: "tabs and spaces",
			line: " \t  hello",
			want: " \t  ",
		},
		{
			name: "no indentation",
			line: "hello",
			want: "",
		},
		{
			name: "only spaces",
			line: "    ",
			want: "    ",
		},
		{
			name: "only tabs",
			line: "\t\t",
			want: "\t\t",
		},
		{
			name: "empty string",
			line: "",
			want: "",
		},
		{
			name: "newline and tab",
			line: "\n\tfoo",
			want: "\n\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIndentation(tt.line)

			if got != tt.want {
				t.Errorf("GetIndentation() = %q, want %q", got, tt.want)
			}
		})
	}
}
