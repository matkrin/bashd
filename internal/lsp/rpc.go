package lsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

func EncodeMessage(msg any) string {
	content, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(content), content)
}

type BaseMessage struct {
	Method string `json:"method"`
}

func DecodeMessage(msg []byte) (string, []byte, error) {
	header, content, found := bytes.Cut(msg, []byte{'\r', '\n', '\r', '\n'})
	if !found {
		return "", nil, errors.New("Not found separator")
	}

	contentLengthBytes := header[len("Content-Length: "):]
	contentLength, err := strconv.Atoi(string(contentLengthBytes))
	if err != nil {
		return "", nil, err
	}

	var basemessage BaseMessage
	if err := json.Unmarshal(content[:contentLength], &basemessage); err != nil {
		return "", nil, err
	}

	return basemessage.Method, content[:contentLength], nil
}

func Split(data []byte, _ bool) (advance int, token []byte, err error) {
	header, content, found := bytes.Cut(data, []byte{'\r', '\n', '\r', '\n'})
	if !found {
		return 0, nil, nil
	}

	contentLengthBytes := header[len("Content-Length: "):]
	contentLength, err := strconv.Atoi(string(contentLengthBytes))
	if err != nil {
		return 0, nil, err
	}

	if len(content) < contentLength {
		return 0, nil, nil
	}

	// 4 is \r\n\r\n
	totalLength := len(header) + 4 + contentLength
	return totalLength, data[:totalLength], nil
}
