package server

import (
	"github.com/matkrin/bashd/lsp"
	"mvdan.cc/sh/v3/syntax"
)

var KEYWORDS = [...]string{
	"if",
	"then",
	"elif",
	"else",
	"fi",
	"for",
	"in",
	"do",
	"done",
	"case",
	"esac",
	"select",
	"function",
	"{",
	"}",
	"[[",
	"]]",
	"!",
	"time",
	"until",
	"while",
	"coproc",
}

func handleCompletion(request *lsp.CompletionRequest, state *State) *lsp.CompletionResponse {
	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		return nil
	}

	completionList := []lsp.CompletionItem{}

	triggerChar := request.Params.Context.TriggerCharacter
	if triggerChar != nil && *triggerChar == "$" {
		completionList = append(completionList, completeDollar(fileAst, state)...)
	} else {
		for _, keyword := range KEYWORDS {
			completionItem := lsp.CompletionItem{
				Label:         keyword,
				Kind:          lsp.CompletionKeyword,
				Detail:        "detail",
				Documentation: "doc",
			}
			completionList = append(completionList, completionItem)
		}

		for _, pathItem := range state.PathItems {
			completionItem := lsp.CompletionItem{
				Label:         pathItem,
				Kind:          lsp.CompletionFunction,
				Detail:        "detail",
				Documentation: "doc",
			}
			completionList = append(completionList, completionItem)
		}

		if fileAst, err := parseDocument(document, uri); err == nil {
			astCompletionItems := findCompletionItems(fileAst)
			completionList = append(completionList, astCompletionItems...)
		}
	}

	response := lsp.CompletionResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result: completionList,
	}

	return &response
}

// Variables defined in Document and Environment Variables
func completeDollar(file *syntax.File, state *State) []lsp.CompletionItem {
	var result []lsp.CompletionItem

	// Variables
	syntax.Walk(file, func(node syntax.Node) bool {
		assign, ok := node.(*syntax.Assign)
		if !ok {
			return true
		}
		if assign.Name != nil {
			result = append(result, lsp.CompletionItem{
				Label:         assign.Name.Value,
				Kind:          lsp.CompletionVariable,
				Detail:        "",
				Documentation: "Variable",
			})
		}

		return true
	})

	// Environment variables
	for envVarName, envVarValue := range state.EnvVars {
		result = append(result, lsp.CompletionItem{
			Label:         envVarName,
			Kind:          lsp.CompletionConstant,
			Detail:        envVarValue,
			Documentation: "",
		})
	}

	return result
}

func findCompletionItems(file *syntax.File) []lsp.CompletionItem {
	var result []lsp.CompletionItem

	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			return false
		}

		var name string
		var kind lsp.CompletionItemKind

		switch n := node.(type) {
		case *syntax.Assign:
			if n.Name != nil {
				name = n.Name.Value
				kind = lsp.CompletionVariable
			}
		case *syntax.FuncDecl:
			if n.Name != nil {
				name = n.Name.Value
				kind = lsp.CompletionFunction
			}

		}

		if name != "" {
			result = append(result, lsp.CompletionItem{
				Label:         name,
				Kind:          kind,
				Detail:        "",
				Documentation: "",
			})
		}

		return true
	})

	return result
}
