package mcp

import "github.com/mark3labs/mcp-go/mcp"

func annotateTextContent(r mcp.TextContent, audience []mcp.Role, priority float64) mcp.TextContent {
	r.Annotations = &mcp.Annotations{
		Audience: audience,
		Priority: priority,
	}
	return r
}
