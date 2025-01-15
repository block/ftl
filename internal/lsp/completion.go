package lsp

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

//go:embed markdown/completion/go/verb.md
var verbCompletionDocs string

//go:embed markdown/completion/go/enumType.md
var enumTypeCompletionDocs string

//go:embed markdown/completion/go/enumValue.md
var enumValueCompletionDocs string

//go:embed markdown/completion/go/typeAlias.md
var typeAliasCompletionDocs string

//go:embed markdown/completion/go/ingress.md
var ingressCompletionDocs string

//go:embed markdown/completion/go/cron.md
var cronCompletionDocs string

//go:embed markdown/completion/go/cronExpression.md
var cronExpressionCompletionDocs string

//go:embed markdown/completion/go/retry.md
var retryCompletionDocs string

//go:embed markdown/completion/go/retryWithCatch.md
var retryWithCatchCompletionDocs string

//go:embed markdown/completion/go/config.md
var configCompletionDocs string

//go:embed markdown/completion/go/secret.md
var secretCompletionDocs string

//go:embed markdown/completion/go/pubSubTopic.md
var pubSubTopicCompletionDocs string

//go:embed markdown/completion/go/pubSubSubscription.md
var pubSubSubscriptionCompletionDocs string

//go:embed markdown/completion/java/verb.md
var verbCompletionDocsJava string

//go:embed markdown/completion/kotlin/verb.md
var verbCompletionDocsKotlin string

// Markdown is split by "---". First half is completion docs, second half is insert text.
var goCompletionItems = []protocol.CompletionItem{
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocs),
	completionItem("ftl:enum:sumtype", "FTL Enum (sum type)", enumTypeCompletionDocs),
	completionItem("ftl:enum:value", "FTL Enum (value type)", enumValueCompletionDocs),
	completionItem("ftl:typealias", "FTL Type Alias", typeAliasCompletionDocs),
	completionItem("ftl:ingress", "FTL Ingress", ingressCompletionDocs),
	completionItem("ftl:cron", "FTL Cron", cronCompletionDocs),
	completionItem("ftl:cron:expression", "FTL Cron with expression", cronExpressionCompletionDocs),
	completionItem("ftl:retry", "FTL Retry", retryCompletionDocs),
	completionItem("ftl:retry:catch", "FTL Retry with catch", retryWithCatchCompletionDocs),
	completionItem("ftl:config", "Create a new configuration value", configCompletionDocs),
	completionItem("ftl:secret", "Create a new secret value", secretCompletionDocs),
	completionItem("ftl:pubsub:topic", "Create a PubSub topic", pubSubTopicCompletionDocs),
	completionItem("ftl:pubsub:subscription", "Create a PubSub subscription", pubSubSubscriptionCompletionDocs),
}

var javaCompletionItems = []protocol.CompletionItem{
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocsJava),
}

var kotlinCompletionItems = []protocol.CompletionItem{
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocsKotlin),
}

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

func completionItem(label, detail, markdown string) protocol.CompletionItem {
	snippetKind := protocol.CompletionItemKindSnippet
	insertTextFormat := protocol.InsertTextFormatSnippet

	parts := strings.Split(markdown, "---")
	if len(parts) != 2 {
		panic(fmt.Sprintf("completion item %q: invalid markdown. must contain exactly one '---' to separate completion docs from insert text", label))
	}

	insertText := strings.TrimSpace(parts[1])
	// Warn if we see two spaces in the insert text.
	if strings.Contains(insertText, "  ") {
		panic(fmt.Sprintf("completion item %q: contains two spaces in the insert text. Use tabs instead!", label))
	}

	// If there is a `//ftl:` this can be autocompleted when the user types `/`.
	if strings.Contains(insertText, "//ftl:") {
		directiveItems[label] = true
	}

	return protocol.CompletionItem{
		Label:      label,
		Kind:       &snippetKind,
		Detail:     &detail,
		InsertText: &insertText,
		Documentation: &protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: strings.TrimSpace(parts[0]),
		},
		InsertTextFormat: &insertTextFormat,
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

		// Currently all completions are in global scope, so the completion must be triggered at the beginning of the line.
		// To do this, check to the start of the line and if there is any whitespace, it is not completing a whole word from the start.
		// We also want to check that the cursor is at the end of the line so we dont let stray chars shoved at the end of the completion.
		isAtEOL := character == len(lineContent)
		if !isAtEOL {
			return nil, nil
		}

		// Is not completing from the start of the line.
		if strings.ContainsAny(lineContent, " \t") {
			return nil, nil
		}

		// If there is a single `/` at the start of the line, we can autocomplete directives. eg `/f`.
		// This is a hint to the user that these are ftl directives.
		// Note that what I can tell, VSCode won't trigger completion on and after `//` so we can only complete on half of a comment.
		isSlashed := strings.HasPrefix(lineContent, "/")
		if isSlashed {
			lineContent = strings.TrimPrefix(lineContent, "/")
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

		// Filter completion items based on the line content and if it is a directive
		var filteredItems []protocol.CompletionItem
		for _, item := range items {
			if !strings.Contains(item.Label, lineContent) {
				continue
			}

			if isSlashed && !directiveItems[item.Label] {
				continue
			}

			if isSlashed {
				// Remove that / from the start of the line, so that the completion doesn't have `///`.
				// VSCode doesn't seem to want to remove the `/` for us.
				item.AdditionalTextEdits = []protocol.TextEdit{
					{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(line),
								Character: 0,
							},
							End: protocol.Position{
								Line:      uint32(line),
								Character: 1,
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
