package mcp

import "github.com/mark3labs/mcp-go/mcp"

func annotateTextContent(r mcp.TextContent, audience []mcp.Role, priority float64) mcp.TextContent {
	r.Annotations = &struct {
		Audience []mcp.Role `json:"audience,omitempty"`
		Priority float64    `json:"priority,omitempty"`
	}{
		Audience: audience,
		Priority: priority,
	}
	return r
}
