package lsp

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Track which directives are //ftl: prefixed, so the we can autocomplete them via `/`.
// This is built at init time and does not change during runtime.
var directiveItems = map[string]bool{}

func init() {
	// Build directiveItems map from all completion items
	for _, item := range goCompletionItems {
		if strings.Contains(*item.InsertText, "//ftl:") {
			directiveItems[item.Label] = true
		}
	}
	for _, item := range javaCompletionItems {
		if strings.Contains(*item.InsertText, "//ftl:") {
			directiveItems[item.Label] = true
		}
	}
	for _, item := range kotlinCompletionItems {
		if strings.Contains(*item.InsertText, "//ftl:") {
			directiveItems[item.Label] = true
		}
	}
}

func (s *Server) textDocumentCompletion() protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (interface{}, error) {
		uri := params.TextDocument.URI
		position := params.Position

		doc, ok := s.documents.get(uri)
		if !ok {
			return nil, nil
		}

		// Line and Character are 0-based, however the cursor can be after the last character in the line.
		line := int(position.Line)
		if line >= len(doc.lines) {
			return nil, nil
		}
		lineContent := doc.lines[line]
		character := int(position.Character)
		if character > len(lineContent) {
			character = len(lineContent)
		}

		// Get file extension to determine language
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(uri), "."))

		// Check if we're at the end of the content (ignoring trailing whitespace)
		trimmedLine := strings.TrimRight(lineContent, " \t")
		isAtEOL := character >= len(trimmedLine)
		if !isAtEOL {
			return nil, nil
		}

		// Trim leading whitespace to handle indentation
		trimmedContent := strings.TrimLeft(lineContent, " \t")
		if trimmedContent == "" {
			return nil, nil
		}

		// If there is a single `/` at the start of the trimmed content, we can autocomplete directives
		isSlashed := strings.HasPrefix(trimmedContent, "/")
		if isSlashed {
			trimmedContent = strings.TrimPrefix(trimmedContent, "/")
		}

		// Get the appropriate completion items based on file extension
		var items []protocol.CompletionItem
		switch ext {
		case "go":
			items = goCompletionItems
		case "java":
			items = javaCompletionItems
		case "kt":
			items = kotlinCompletionItems
		default:
			return nil, nil
		}

		// Calculate the indentation level
		indentLength := len(lineContent) - len(trimmedContent)

		// Filter completion items based on the line content and if it is a directive
		var filteredItems []protocol.CompletionItem
		for _, item := range items {
			if !strings.Contains(item.Label, trimmedContent) {
				continue
			}

			if isSlashed && !directiveItems[item.Label] {
				continue
			}

			if isSlashed {
				// Remove the / from the start of the line
				item.AdditionalTextEdits = []protocol.TextEdit{
					{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(line),
								Character: uint32(indentLength),
							},
							End: protocol.Position{
								Line:      uint32(line),
								Character: uint32(indentLength + 1),
							},
						},
						NewText: "",
					},
				}
			}

			filteredItems = append(filteredItems, item)
		}

		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        filteredItems,
		}, nil
	}
}

func (s *Server) completionItemResolve() protocol.CompletionItemResolveFunc {
	return func(context *glsp.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
		if path, ok := params.Data.(string); ok {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			params.Documentation = &protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: string(content),
			}
		}

		return params, nil
	}
}
