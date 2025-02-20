package lsp

import (
	_ "embed"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) textDocumentHover() protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		uri := params.TextDocument.URI
		position := params.Position

		doc, ok := s.documents.get(uri)
		if !ok {
			s.logger.Debugf("No document found for URI: %s", uri)
			return nil, nil
		}

		line := int(position.Line)
		if line >= len(doc.lines) {
			s.logger.Debugf("Line %d is out of range for document", line)
			return nil, nil
		}

		lineContent := doc.lines[line]
		character := int(position.Character)
		if character > len(lineContent) {
			character = len(lineContent)
		}
		s.logger.Debugf("Hover request - Line: %d, Character: %d, Content: %q", line, character, lineContent)

		// Determine the language based on file extension
		var lang string
		if strings.HasSuffix(uri, ".kt") {
			lang = "kotlin"
		} else if strings.HasSuffix(uri, ".go") {
			lang = "go"
		} else if strings.HasSuffix(uri, ".java") {
			lang = "java"
		} else {
			lang = "go" // Default to Go if unknown
		}
		s.logger.Debugf("File language: %s", lang)

		for hoverString, hoverContent := range hoverMap {
			s.logger.Debugf("Checking hover string: %q", hoverString)

			// Simple exact string matching for all patterns
			startIndex := strings.Index(lineContent, hoverString)
			if startIndex != -1 && startIndex <= character && character <= startIndex+len(hoverString) {
				s.logger.Debugf("Found match for %q at position %d", hoverString, startIndex)

				// Get language-specific content
				content, ok := hoverContent[lang]
				if !ok {
					// Fallback to any available language if specific one not found
					for _, c := range hoverContent {
						content = c
						break
					}
				}

				return &protocol.Hover{
					Contents: &protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: content,
					},
				}, nil
			}
		}

		s.logger.Debugf("No hover match found")
		return nil, nil
	}
}
