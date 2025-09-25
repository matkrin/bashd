package lsp

type DocumentSymbolsRequest struct {
	Request
	Params DocumentSymbolsParams `json:"params"`
}

type DocumentSymbolsParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbolResponse struct {
	Response
	Result []DocumentSymbol `json:"result"`
}

func NewDocumentSymbolResponse(
	id int,
	documentSymbols []DocumentSymbol,
) DocumentSymbolResponse {
	return DocumentSymbolResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: documentSymbols,
	}
}

type DocumentSymbol struct {
	Name string `json:"name"`
	// Detail string `json:"detail"`
	Kind SymbolKind `json:"kind"`
	// Tags
	// Deprecated
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children"`
}

type SymbolKind int

const (
	SymbolFile SymbolKind = iota + 1
	SymbolModule
	SymbolNamespace
	SymbolPackage
	SymbolClass
	SymbolMethod
	SymbolProperty
	SymbolField
	SymbolConstructor
	SymbolEnum
	SymbolInterface
	SymbolFunction
	SymbolVariable
	SymbolConstant
	SymbolString
	SymbolNumber
	SymbolBoolean
	SymbolArray
	SymbolObject
	SymbolKey
	SymbolNull
	SymbolEnumMember
	SymbolStruct
	SymbolEvent
	SymbolOperator
	SymbolTypeParameter
)
