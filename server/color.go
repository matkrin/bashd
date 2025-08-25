package server

import (
	"log/slog"

	"github.com/matkrin/bashd/lsp"
)

func handleDocumentColor(request *lsp.DocumentColorRequest, state *State) *lsp.DocumentColorResponse {
	slog.Info("COLOR", "params", request.Params)

	color := lsp.Color{
		Red:   1,
		Green: 1,
		Blue:  1,
		Alpha: 1,
	}
	colorInformation := lsp.ColorInformation{
		Range: lsp.NewRange(0, 0, 0, 5),
		Color: color,
	}

	response := &lsp.DocumentColorResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &request.ID,
		},
		Result:   []lsp.ColorInformation{colorInformation},
	}
	return response
}
