package server

import (
	"fmt"

	"github.com/matkrin/bashd/internal/ast"
	"github.com/matkrin/bashd/internal/lsp"
	"mvdan.cc/sh/v3/syntax"
)

// Handler for `textDocument/completion`
func handleCompletion(request *lsp.CompletionRequest, state *State) *lsp.CompletionResponse {
	completionList := []lsp.CompletionItem{}

	uri := request.Params.TextDocument.URI
	document := state.Documents[uri].Text
	fileAst, err := ast.ParseDocument(document, uri)
	if err != nil {
		response := lsp.NewCompletionResponse(request.ID, completionList)
		return &response
	}

	triggerChar := request.Params.Context.TriggerCharacter
	if triggerChar != nil && (*triggerChar == "$" || *triggerChar == "{") {
		completionList = append(completionList, completeDollar(fileAst, state)...)
	} else {
		completionList = append(completionList, completionKeywords()...)
		completionList = append(completionList, completionFunctions(fileAst)...)
		completionList = append(completionList, completionPathItem(state)...)
	}

	response := lsp.NewCompletionResponse(request.ID, completionList)
	return &response
}

// Handler for `completionItem/resolve`
func handleCompletionItemResolve(
	request *lsp.CompletionItemResolveRequest,
) *lsp.CompletionItemResolveResponse {
	completionItem := request.Params.CompletionItem
	documentation := getDocumentation(completionItem.Label)
	mdDocumentation := fmt.Sprintf("```man\n%s\n```", documentation)

	completionItem.Documentation = &lsp.MarkupContent{
		Kind:  lsp.MarkupKindMarkdown,
		Value: mdDocumentation,
	}

	response := lsp.CompletionItemResolveResponse{
		Response: lsp.Response{
			RPC: lsp.RPC_VERSION,
			ID:  &request.ID,
		},
		Result: completionItem,
	}
	return &response
}

// Completion for variables defined in Document and environment variables
func completeDollar(ast *ast.Ast, state *State) []lsp.CompletionItem {
	var result []lsp.CompletionItem

	// Variables
	syntax.Walk(ast.File, func(node syntax.Node) bool {
		assign, ok := node.(*syntax.Assign)
		if !ok {
			return true
		}
		if assign.Name != nil {
			result = append(result, lsp.CompletionItem{
				Label:         assign.Name.Value,
				Kind:          lsp.CompletionVariable,
				Detail:        "",
				Documentation: nil,
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
			Documentation: nil,
		})
	}

	return result
}

// Completion for keywords
func completionKeywords() []lsp.CompletionItem {
	var result []lsp.CompletionItem
	for _, keyword := range BASH_KEYWORDS {
		completionItem := lsp.CompletionItem{
			Label:         keyword,
			Kind:          lsp.CompletionKeyword,
			Detail:        "",
			Documentation: nil,
		}
		result = append(result, completionItem)
	}

	return result
}

// Completion for function names
func completionFunctions(ast *ast.Ast) []lsp.CompletionItem {
	var result []lsp.CompletionItem

	syntax.Walk(ast.File, func(node syntax.Node) bool {
		funcDecl, ok := node.(*syntax.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Name != nil {
			result = append(result, lsp.CompletionItem{
				Label:         funcDecl.Name.Value,
				Kind:          lsp.CompletionFunction,
				Detail:        "",
				Documentation: nil,
			})
		}

		return true
	})

	return result
}

// Completion for executables in PATH
func completionPathItem(state *State) []lsp.CompletionItem {
	var result []lsp.CompletionItem
	for _, pathItem := range state.PathItems {
		completionItem := lsp.CompletionItem{
			Label:         pathItem,
			Kind:          lsp.CompletionFunction,
			Detail:        "",
			Documentation: nil,
		}
		result = append(result, completionItem)
	}
	return result
}
