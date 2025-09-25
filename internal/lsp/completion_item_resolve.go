package lsp

type CompletionItemResolveRequest struct {
	Request
	Params CompletionItemResolveParams `json:"params"`
}

type CompletionItemResolveParams struct {
	CompletionItem
}

type CompletionItemResolveResponse struct {
	Response
	Result CompletionItem `json:"result"`
}
