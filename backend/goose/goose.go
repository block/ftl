package goose

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec" //nolint:depguard
	"regexp"
	"strings"
	"sync"

	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/log"
)

// MessageSource represents the source of a message from Goose
type MessageSource int

const (
	// SourceStdout represents messages from standard output
	SourceStdout MessageSource = iota
	// SourceStderr represents messages from standard error
	SourceStderr
	// SourceCompletion represents a completion signal
	SourceCompletion
)

type Message struct {
	Content string
	Source  MessageSource
}

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

// Execute runs a Goose command with the given prompt and streams cleaned messages
// to the provided callback function. The callback is called for each cleaned message
// and when execution completes.
//
// For now this uses the Goose CLI via FTL CLI to execute the command, but we should switch to
// native Goose API calls when available.
func (c *Client) Execute(ctx context.Context, prompt string, callback func(Message)) error {
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	logger := log.FromContext(ctx).Scope("goose")

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	// Only close writers in defer as cleanup in case of early returns
	defer func() {
		if stdoutWriter != nil {
			stdoutWriter.Close()
		}
		if stderrWriter != nil {
			stderrWriter.Close()
		}
	}()

	gitRoot, ok := internal.GitRoot(os.Getenv("FTL_DIR")).Get()
	if !ok {
		return fmt.Errorf("failed to find Git root")
	}

	cmd := exec.CommandContext(ctx, "ftl", "goose", prompt)
	cmd.Dir = gitRoot
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	scanOutput := func(reader *io.PipeReader, source MessageSource) error {
		defer reader.Close()
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if shouldFilterLine(line) {
				continue
			}
			if cleaned := cleanMessage(line); cleaned != "" {
				callback(Message{
					Content: cleaned,
					Source:  source,
				})
			}
		}
		return scanner.Err()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start goose command: %w", err)
	}

	go func() {
		defer wg.Done()
		if err := scanOutput(stdoutReader, SourceStdout); err != nil {
			logger.Warnf("Error reading stdout: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := scanOutput(stderrReader, SourceStderr); err != nil {
			logger.Warnf("Error reading stderr: %v", err)
		}
	}()

	err := cmd.Wait()

	// Close writers to signal EOF to readers
	stdoutWriter.Close()
	stderrWriter.Close()
	// Set to nil so defer doesn't try to close again
	stdoutWriter = nil
	stderrWriter = nil

	wg.Wait()
	if err != nil {
		return fmt.Errorf("failed to execute goose command: %w", err)
	}

	callback(Message{
		Content: "",
		Source:  SourceCompletion,
	})
	return nil
}

// shouldFilterLine returns true if the line should be filtered out
func shouldFilterLine(line string) bool {
	// Filter out session management lines and other typical output that should be removed
	return strings.HasPrefix(line, "Closing session.") ||
		strings.Contains(line, "resuming session") ||
		strings.Contains(line, "Session:") ||
		strings.Contains(line, "working directory:") ||
		strings.Contains(line, "logging to")
}

// cleanMessage cleans a message from Goose by:
// - Removing ANSI escape sequences
// - Removing tool output headers and separators
// - Cleaning up paths and repeated info
// - Preserving important line breaks while removing excessive ones
// - Handling markdown elements
// - Removing duplicate sections while preserving structure
func cleanMessage(s string) string {
	s = stripAnsiCodes(s)

	toolPattern := `─+\s*(Status|Timeline|Read|Write|NewModule|CallVerb|ResetSubscription|NewMySQLDatabase|NewMySQLMigration|NewPostgresDatabase|NewPostgresMigration|SubscriptionInfo)\s*\|\s*\w+\s*─+`
	hashPattern := `###[^#]+###`
	pathPattern := `(?:path|content|verificationToken): \.{3}`
	readPattern := `Read contents of [^\n]+\n`
	newlinePattern := `\n{3,}`
	trimPattern := `^\n+|\n+$`
	s = regexp.MustCompile(toolPattern).ReplaceAllString(s, "")
	s = regexp.MustCompile(hashPattern).ReplaceAllString(s, "")
	s = regexp.MustCompile(pathPattern).ReplaceAllString(s, "")
	s = regexp.MustCompile(readPattern).ReplaceAllString(s, "")
	s = regexp.MustCompile(newlinePattern).ReplaceAllString(s, "\n\n")
	s = regexp.MustCompile(trimPattern).ReplaceAllString(s, "")

	result := strings.Join(splitIntoBlocks(s), "\n\n")
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}

// splitIntoBlocks splits content into logical blocks while preserving formatting
func splitIntoBlocks(s string) []string {
	var blocks []string
	var currentBlock []string

	lines := strings.Split(s, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start a new block if:
		// 1. Current line is empty and we have content in currentBlock
		// 2. Current line starts a new markdown block
		// 3. Current line starts a new list item and previous wasn't a list
		if len(currentBlock) > 0 && (trimmed == "" ||
			isMarkdownBlockStart(trimmed) ||
			(isListItem(trimmed) && !isListItem(strings.TrimSpace(currentBlock[len(currentBlock)-1])))) {

			if blockContent := strings.TrimSpace(strings.Join(currentBlock, "\n")); blockContent != "" {
				blocks = append(blocks, blockContent)
			}
			currentBlock = nil

			// Skip empty lines between blocks
			if trimmed == "" {
				continue
			}
		}

		currentBlock = append(currentBlock, line)
	}

	if len(currentBlock) > 0 {
		if blockContent := strings.TrimSpace(strings.Join(currentBlock, "\n")); blockContent != "" {
			blocks = append(blocks, blockContent)
		}
	}

	return blocks
}

// isMarkdownBlockStart checks if a line starts a new markdown block
func isMarkdownBlockStart(line string) bool {
	return strings.HasPrefix(line, "#") || // Headers
		strings.HasPrefix(line, "```") || // Code blocks
		strings.HasPrefix(line, ">") || // Blockquotes
		strings.HasPrefix(line, "- ") || // Unordered lists
		strings.HasPrefix(line, "* ") || // Alternative unordered lists
		regexp.MustCompile(`^\d+\.\s`).MatchString(line) // Ordered lists
}

// isListItem checks if a line is a list item
func isListItem(line string) bool {
	return strings.HasPrefix(line, "- ") ||
		strings.HasPrefix(line, "* ") ||
		regexp.MustCompile(`^\d+\.\s`).MatchString(line)
}

// stripAnsiCodes removes ANSI escape sequences from a string
func stripAnsiCodes(s string) string {
	ansiColorPattern := `\x1b\[[0-9;]*[mK]`
	ansiOSCPattern := `\x1b\][0-9];.*?\x1b\\`
	ansiOSCHyperlinkPattern := `\x1b\]8;;.*?\x1b\\`

	result := regexp.MustCompile(ansiColorPattern).ReplaceAllString(s, "")
	result = regexp.MustCompile(ansiOSCPattern).ReplaceAllString(result, "")
	result = regexp.MustCompile(ansiOSCHyperlinkPattern).ReplaceAllString(result, "")
	result = strings.ReplaceAll(result, "0m", "")
	result = strings.TrimRight(result, " \t\n\r")

	return result
}
