// This program generates hover items for the FTL LSP server.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/tliron/kutil/terminal"
)

type CLI struct {
	Config  string `type:"filepath" default:"internal/lsp/hover.json" help:"Path to the hover configuration file"`
	DocRoot string `type:"dirpath" default:"docs/docs" help:"Path to the config referenced markdowns"`
	Output  string `type:"filepath" default:"internal/lsp/hoveritems.go" help:"Path to the generated Go file"`
}

var cli CLI

type hover struct {
	// Match these texts for triggering this hover, e.g. ["//ftl:typealias", "@TypeAlias"]
	Matches []string `json:"matches"`

	// Source file to read from.
	Source string `json:"source"`

	// Select these heading to use for the docs. If omitted, the entire markdown file is used.
	// Headings are included in the output.
	Select []string `json:"select,omitempty"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`Generate hover items for FTL LSP. See lsp/hover.go`),
	)

	hovers, err := parseHoverConfig(cli.Config)
	kctx.FatalIfErrorf(err)

	items, err := scrapeDocs(hovers)
	kctx.FatalIfErrorf(err)

	err = writeGoFile(cli.Output, items)
	kctx.FatalIfErrorf(err)
}

func scrapeDocs(hovers []hover) (map[string]map[string]string, error) {
	// Map of hover string to language-specific content
	items := make(map[string]map[string]string)
	for _, hover := range hovers {
		path := filepath.Join(cli.DocRoot, hover.Source)
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", path, err)
		}

		doc, err := getMarkdownWithTitle(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		var content string
		if len(hover.Select) > 0 {
			for _, sel := range hover.Select {
				chunk, err := selector(doc.Content, sel)
				if err != nil {
					return nil, fmt.Errorf("failed to select %s from %s: %w", sel, path, err)
				}
				content += chunk
			}
		} else {
			content = doc.Content
		}

		// Extract language-specific sections
		langContent := extractLanguageSections(content)

		// Add the same content for each match pattern
		for _, match := range hover.Matches {
			items[match] = langContent
		}

		// get term width or 80 if not available
		width := 80
		if w, _, err := terminal.GetSize(); err == nil {
			width = w
		}
		line := strings.Repeat("-", width)
		fmt.Print("\x1b[1;34m")
		fmt.Println(line)
		fmt.Printf("Matches: %v\n", hover.Matches)
		fmt.Println(line)
		fmt.Print("\x1b[0m")
		fmt.Println(content)
	}
	return items, nil
}

func parseHoverConfig(path string) ([]hover, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	var hovers []hover
	err = json.NewDecoder(file).Decode(&hovers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON %s: %w", path, err)
	}

	return hovers, nil
}

type Doc struct {
	Content string
}

// getMarkdownWithTitle reads a Docusaurus markdown file and returns the full markdown content and the title.
func getMarkdownWithTitle(file *os.File) (*Doc, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", file.Name(), err)
	}

	// Find the frontmatter boundaries
	lines := strings.Split(string(content), "\n")
	var frontmatterEnd int
	foundStart := false

	for i, line := range lines {
		if line == "---" {
			if !foundStart {
				foundStart = true
			} else {
				frontmatterEnd = i
				break
			}
		}
	}

	if !foundStart || frontmatterEnd == 0 {
		return nil, fmt.Errorf("file %s does not contain frontmatter delimiters", file.Name())
	}

	// Join the content after frontmatter
	contentLines := lines[frontmatterEnd+1:]
	return &Doc{Content: strings.Join(contentLines, "\n")}, nil
}

var HTMLCommentRegex = regexp.MustCompile(`<!--\w*(.*?)\w*-->`)

func selector(content, selector string) (string, error) {
	// Split the content into lines.
	lines := strings.Split(content, "\n")
	collected := []string{}

	// If the selector starts with ## (the only type of heading we have):
	// Find the line, include it, and all lines until the next heading.
	if !strings.HasPrefix(selector, "##") {
		return "", fmt.Errorf("unsupported selector %s", selector)
	}
	include := false
	for _, line := range lines {
		if include {
			// We have found another heading. Abort!
			if strings.HasPrefix(line, "##") {
				break
			}

			// We also stop at a line break, because we don't want to include footnotes.
			// See the end of docs/content/docs/reference/types.md for an example.
			if line == "---" {
				break
			}

			collected = append(collected, line)
		}

		// Start collecting
		if strings.HasPrefix(line, selector) {
			include = true
			collected = append(collected, line)
		}
	}

	if len(collected) == 0 {
		return "", fmt.Errorf("no content found for selector %s", selector)
	}

	return strings.TrimSpace(strings.Join(collected, "\n")) + "\n", nil
}

func writeGoFile(path string, items map[string]map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}
	defer file.Close()

	tmpl, err := template.New("").Parse(`// Code generated by 'just lsp-generate'. DO NOT EDIT.
package lsp

var hoverMap = map[string]map[string]string{
{{- range $match, $content := . }}
	{{ printf "%q" $match }}: {
		{{- range $lang, $content := $content }}
			{{ printf "%q" $lang }}: {{ printf "%q" $content }},
		{{- end }}
	},
{{- end }}
}
`)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = tmpl.Execute(file, items)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Generated %s\n", path)
	return nil
}

// extractLanguageSections extracts content for different languages from markdown tabs
func extractLanguageSections(content string) map[string]string {
	sections := make(map[string]string)

	var currentLang string
	var currentContent []string
	inTabs := false

	// Initialize sections with common content
	sections["go"] = ""
	sections["kotlin"] = ""
	sections["java"] = ""

	content = removeDocusaurusImports(content)
	lines := strings.Split(content, "\n")

	for i := range lines {
		line := lines[i]

		if strings.Contains(line, "<Tabs") {
			inTabs = true
			continue
		}

		if strings.Contains(line, "</Tabs>") {
			inTabs = false
			// When we exit a tab section, append any accumulated content for the current language
			if currentLang != "" {
				sections[currentLang] += strings.Join(currentContent, "\n") + "\n"
				currentContent = nil
				currentLang = ""
			}
			continue
		}

		// Look for language tab items
		if strings.Contains(line, "<TabItem value=") {
			// Store any accumulated content for the previous language
			if currentLang != "" {
				sections[currentLang] += strings.Join(currentContent, "\n") + "\n"
				currentContent = nil
			}

			// Extract language from TabItem
			if start := strings.Index(line, `value="`); start != -1 {
				if end := strings.Index(line[start+7:], `"`); end != -1 {
					currentLang = line[start+7 : start+7+end]
				}
			}
			continue
		}

		if strings.Contains(line, "</TabItem>") {
			if currentLang != "" {
				// Store the current language's content when we reach the end of its tab
				sections[currentLang] += strings.Join(currentContent, "\n") + "\n"
				currentContent = nil
				currentLang = ""
			}
			continue
		}

		if inTabs {
			if currentLang != "" {
				currentContent = append(currentContent, line)
			}
		} else {
			// Content outside of any tabs is common to all languages
			for lang := range sections {
				sections[lang] += line + "\n"
			}
		}
	}

	// Handle the last language section if we haven't already
	if currentLang != "" && len(currentContent) > 0 {
		sections[currentLang] += strings.Join(currentContent, "\n") + "\n"
	}

	return sections
}

// removeDocusaurusImports removes docusaurus-specific import statements
func removeDocusaurusImports(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "import Tabs from '@theme/Tabs'") ||
			strings.HasPrefix(line, "import TabItem from '@theme/TabItem'") {
			continue
		}
		result = append(result, lines[i])
	}

	return strings.Join(result, "\n")
}
