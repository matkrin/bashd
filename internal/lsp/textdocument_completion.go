package lsp

type CompletionRequest struct {
	Request
	Params CompletionParams `json:"params"`
}

type CompletionParams struct {
	TextDocumentPositionParams
	Context CompletionContext `json:"context"`
}

type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter *string               `json:"triggerCharacter"`
}

type CompletionTriggerKind int

const (
	CompletionInvoked                         CompletionTriggerKind = 1
	CompletionTriggerCharacter                CompletionTriggerKind = 2
	CompletionTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

type CompletionResponse struct {
	Response
	Result []CompletionItem `json:"result"`
}

func NewCompletionResponse(id int, completionList []CompletionItem) CompletionResponse {
	return CompletionResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: completionList,
	}
}

type CompletionItem struct {
	Label        string                      `json:"label"`
	LabelDetails *CompletionItemLabelDetails `json:"labelDetails"`
	Kind         CompletionItemKind          `json:"kind"`
	// Tags
	Detail        string         `json:"detail"`
	Documentation *MarkupContent `json:"documentation"` // Can make this markdown
	// Deprecated
	// Preselect
	// SortText
	// FilterText
	// InsertText
	// InsertTextFormat
	// InsertTextMode
	// TextEdit
	// TextEditText
	// AddtionalTextEdits
	// CommitCharacters
	// Command
	// Data
}

type CompletionItemLabelDetails struct {
	Detail      *string `json:"detail"`
	Description *string `json:"description"`
}

type CompletionItemKind int

const (
	CompletionText CompletionItemKind = iota + 1
	CompletionMethod
	CompletionFunction
	CompletionConstructor
	CompletionField
	CompletionVariable
	CompletionClass
	CompletionInterface
	CompletionModule
	CompletionProperty
	CompletionUnit
	CompletionValue
	CompletionEnum
	CompletionKeyword
	CompletionSnippet
	CompletionColor
	CompletionFile
	CompletionReference
	CompletionFolder
	CompletionEnumMember
	CompletionConstant
	CompletionStruct
	CompletionEvent
	CompletionOperator
	CompletionTypeParameter
)

type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)
