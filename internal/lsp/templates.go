package lsp

import (
	_ "embed"
	"fmt"
	"strings"

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

//go:embed markdown/completion/go/config.md
var configCompletionDocs string

//go:embed markdown/completion/go/secret.md
var secretCompletionDocs string

//go:embed markdown/completion/go/pubSubTopic.md
var pubSubTopicCompletionDocs string

//go:embed markdown/completion/go/pubSubSubscription.md
var pubSubSubscriptionCompletionDocs string

//go:embed markdown/completion/go/logger.md
var loggerCompletionDocs string

// Markdown is split by "---". First half is completion docs, second half is insert text.
var goCompletionItems = []protocol.CompletionItem{
	completionItem("ftl:config", "Create a new configuration value", configCompletionDocs),
	completionItem("ftl:cron", "FTL Cron", cronCompletionDocs),
	completionItem("ftl:cron:expression", "FTL Cron with expression", cronExpressionCompletionDocs),
	completionItem("ftl:enum:sumtype", "FTL Enum (sum type)", enumTypeCompletionDocs),
	completionItem("ftl:enum:value", "FTL Enum (value type)", enumValueCompletionDocs),
	completionItem("ftl:ingress", "FTL Ingress", ingressCompletionDocs),
	completionItem("ftl:pubsub:subscription", "Create a PubSub subscription", pubSubSubscriptionCompletionDocs),
	completionItem("ftl:pubsub:topic", "Create a PubSub topic", pubSubTopicCompletionDocs),
	completionItem("ftl:retry", "FTL Retry", retryCompletionDocs),
	completionItem("ftl:secret", "Create a new secret value", secretCompletionDocs),
	completionItem("ftl:typealias", "FTL Type Alias", typeAliasCompletionDocs),
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocs),
	completionItem("ftl:logger", "FTL Logger", loggerCompletionDocs),
}

//go:embed markdown/completion/kotlin/verb.md
var verbCompletionDocsKotlin string

//go:embed markdown/completion/kotlin/config.md
var configCompletionDocsKotlin string

//go:embed markdown/completion/kotlin/ingress.md
var ingressCompletionDocsKotlin string

//go:embed markdown/completion/kotlin/cron.md
var cronCompletionDocsKotlin string

//go:embed markdown/completion/kotlin/pubSubTopic.md
var pubSubTopicCompletionDocsKotlin string

//go:embed markdown/completion/kotlin/pubSubSubscription.md
var pubSubSubscriptionCompletionDocsKotlin string

//go:embed markdown/completion/kotlin/secret.md
var secretCompletionDocsKotlin string

var kotlinCompletionItems = []protocol.CompletionItem{
	completionItem("ftl:config", "Create a new configuration value", configCompletionDocsKotlin),
	completionItem("ftl:cron", "FTL Cron", cronCompletionDocsKotlin),
	completionItem("ftl:ingress", "FTL Ingress", ingressCompletionDocsKotlin),
	completionItem("ftl:pubsub:subscription", "Create a PubSub subscription", pubSubSubscriptionCompletionDocsKotlin),
	completionItem("ftl:pubsub:topic", "Create a PubSub topic", pubSubTopicCompletionDocsKotlin),
	completionItem("ftl:secret", "Create a new secret value", secretCompletionDocsKotlin),
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocsKotlin),
}

//go:embed markdown/completion/java/config.md
var configCompletionDocsJava string

//go:embed markdown/completion/java/verb.md
var verbCompletionDocsJava string

//go:embed markdown/completion/java/cron.md
var cronCompletionDocsJava string

//go:embed markdown/completion/java/ingress.md
var ingressCompletionDocsJava string

//go:embed markdown/completion/java/pubSubTopic.md
var pubSubTopicCompletionDocsJava string

//go:embed markdown/completion/java/pubSubSubscription.md
var pubSubSubscriptionCompletionDocsJava string

//go:embed markdown/completion/java/secret.md
var secretCompletionDocsJava string

var javaCompletionItems = []protocol.CompletionItem{
	completionItem("ftl:config", "Create a new configuration value", configCompletionDocsJava),
	completionItem("ftl:cron", "FTL Cron", cronCompletionDocsJava),
	completionItem("ftl:ingress", "FTL Ingress", ingressCompletionDocsJava),
	completionItem("ftl:pubsub:subscription", "Create a PubSub subscription", pubSubSubscriptionCompletionDocsJava),
	completionItem("ftl:pubsub:topic", "Create a PubSub topic", pubSubTopicCompletionDocsJava),
	completionItem("ftl:secret", "Create a new secret value", secretCompletionDocsJava),
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocsJava),
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
