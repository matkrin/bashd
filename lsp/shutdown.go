package lsp

type ShutdownRequest struct {
	Request
}

type ShutdownResponse struct {
	Response
	Result *any
}
