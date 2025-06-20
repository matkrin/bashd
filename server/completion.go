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
	completionList := []lsp.CompletionItem{}
	completionList = append(completionList, completionKeywords()...)

	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text
	fileAst, err := parseDocument(document, uri)
	if err != nil {
		response := lsp.NewCompletionResponse(request.ID, completionList)
		return &response
	}

	triggerChar := request.Params.Context.TriggerCharacter
	if triggerChar != nil && *triggerChar == "$" {
		completionList = append(completionList, completeDollar(fileAst, state)...)
	} else {
		completionList = append(completionList, completionFunctions(fileAst)...)
	}

	response := lsp.NewCompletionResponse(request.ID, completionList)
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

func completionKeywords() []lsp.CompletionItem {
	var result []lsp.CompletionItem
	for _, keyword := range KEYWORDS {
		completionItem := lsp.CompletionItem{
			Label:         keyword,
			Kind:          lsp.CompletionKeyword,
			Detail:        "",
			Documentation: "",
		}
		result = append(result, completionItem)
	}

	return result
}

func completionFunctions(file *syntax.File) []lsp.CompletionItem {
	var result []lsp.CompletionItem

	syntax.Walk(file, func(node syntax.Node) bool {
		funcDecl, ok := node.(*syntax.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Name != nil {
			result = append(result, lsp.CompletionItem{
				Label:         funcDecl.Name.Value,
				Kind:          lsp.CompletionFunction,
				Detail:        "",
				Documentation: "",
			})
		}

		return true
	})

	return result
}
