package schema

import "strings"

// formatTokens formats a list of tokens into a string, joining them with spaces.
// It skips empty tokens.
func formatTokens(sb *strings.Builder, tokens ...string) {
	first := true
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if !first {
			sb.WriteString(" ")
		}
		sb.WriteString(token)
		first = false
	}
}
