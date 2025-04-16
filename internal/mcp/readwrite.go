package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
)

type readResult struct {
	Explanation            string `json:"explanation,omitempty"`
	FileContent            string `json:"fileContent"`
	WriteVerificationToken string `json:"writeVerificationToken"`
}

func ReadTool() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"Read",
			mcp.WithDescription(`Read a text file within an FTL module. This is better than other ways of reading these files as it can be used safely with FTL's Write tool with the verification token that this tool returns.
				Make sure you read every file before you overwrite it so you do not accidentally alter data or code.`),
			mcp.WithString("path", mcp.Description("Path to the file to read")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, ok := request.Params.Arguments["path"].(string)
			if !ok {
				return nil, errors.Errorf("path is required")
			}
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return nil, errors.Wrap(err, "could not read file")
			}
			token, err := tokenForFileContent(fileContent)
			if err != nil {
				return nil, errors.Wrap(err, "could not generate verification token")
			}
			readResult, err := newReadResult(fileContent, token, false, "")
			if err != nil {
				return nil, errors.Wrap(err, "could not create read result")
			}
			readResult.Content = append(readResult.Content, annotateTextContent(mcp.NewTextContent("Read contents of "+path), []mcp.Role{mcp.RoleUser}, 0.3))
			return readResult, nil
		}
}

func newReadResult(fileContent []byte, token string, isError bool, explanation string) (*mcp.CallToolResult, error) {
	readResult := &readResult{
		FileContent:            string(fileContent),
		WriteVerificationToken: token,
		Explanation:            explanation,
	}
	outputBytes, err := json.Marshal(readResult)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal read result")
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			annotateTextContent(mcp.NewTextContent(string(outputBytes)), []mcp.Role{mcp.RoleAssistant}, 1.0),
		},
		IsError: isError,
	}, nil
}

type writeResult struct {
	StatusExplanation    string       `json:"statusExplanation,omitempty"`
	Status               StatusOutput `json:"status,omitempty"`
	TokenExplanation     string       `json:"tokenExplanation,omitempty"`
	NewVerificationToken string       `json:"newVerificationToken"`
}

func WriteTool(serverCtx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"Write",
			mcp.WithDescription(`Write a file withiin an FTL module. This is better than other ways of writing these files as it also prevents writing files in the wrong places and returns the FTL status after the change.
			Be careful with existing files! This is a full overwrite, so you must include everything - not just sections you are modifying.`),
			mcp.WithString("path", mcp.Description("Path to the file to write")),
			mcp.WithString("content", mcp.Description("Data to write to the file")),
			mcp.WithString("verificationToken", mcp.Description(`Obtained by the Read tool to verify that the existing content of the file has been read and understood before being replaced. Not required for new files.`)),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, ok := request.Params.Arguments["path"].(string)
			if !ok {
				return nil, errors.Errorf("path is required")
			}
			fileContent, ok := request.Params.Arguments["content"].(string)
			if !ok {
				return nil, errors.Errorf("content is required")
			}
			moduleDir, ok := detectModulePath(path).Get()
			if !ok {
				return nil, errors.Errorf("this tool can only write to files within FTL modules")
			}
			originalGeneratedFiles := getGeneratedFileContent(moduleDir, path)

			originalContent, err := os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					originalContent = []byte{}
				} else {
					return nil, errors.Wrap(err, "could access existing file")
				}
			}
			if len(originalContent) > 0 {
				token, err := tokenForFileContent(originalContent)
				if err != nil {
					return nil, errors.Wrap(err, "could not generate verification token")
				}
				expectedToken, ok := request.Params.Arguments["verificationToken"].(string)
				if !ok || expectedToken != token {
					// File was not read (or file has changed). Return an error response with an explanation and the original content
					return errors.WithStack2(newReadResult(originalContent, token, true, `The file was not read before it was written or it has changed since it was last read.
		The file has been read and provided here. Make sure you understand the existing content before using the Write tool so you do not accidentally alter data or code that you did not mean to change.`))
				}
			}

			// Atomically write the file.
			dir, filename := filepath.Split(path)
			tmpFile, err := os.CreateTemp(dir, filename+"-")
			if err != nil {
				return nil, errors.Wrapf(err, "could not create temp file for %s", path)
			}
			defer os.Remove(tmpFile.Name()) // Delete the temp file if we error.
			defer tmpFile.Close()
			if _, err := tmpFile.WriteString(fileContent); err != nil {
				return nil, errors.Wrapf(err, "could not write to tmp file for %s", path)
			}
			if err := os.Rename(tmpFile.Name(), path); err != nil {
				return nil, errors.Wrapf(err, "could not replace file with tmp file %s", path)
			}

			var userResult mcp.TextContent
			if len(originalContent) == 0 {
				userResult = annotateTextContent(mcp.NewTextContent("###CODEBLOCK### "+path+"\n```\n"+fileContent+"\n<\\...>\n```"), []mcp.Role{mcp.RoleUser}, 0.2)
			} else {
				userResult = annotateTextContent(mcp.NewTextContent(prettyDiff(path, string(originalContent), fileContent)), []mcp.Role{mcp.RoleUser}, 0.2)
			}

			content := []mcp.Content{
				userResult,
			}

			assistantResult := writeResult{}
			assistantResult.NewVerificationToken, err = tokenForFileContent([]byte(fileContent))
			if err == nil {
				assistantResult.TokenExplanation = "The file has been updated. A new verification token is provided if you need to update the file again."
			}

			if status, err := GetStatusOutput(serverCtx, buildEngineClient, adminClient); err == nil {
				assistantResult.StatusExplanation = "The FTL status after the change is also provided."
				assistantResult.Status = status
			}
			assistantResultJSON, err := json.Marshal(assistantResult)
			if err != nil {
				return nil, errors.Wrap(err, "could not marshal assistant result")
			}
			content = append(content, annotateTextContent(mcp.NewTextContent(string(assistantResultJSON)), []mcp.Role{mcp.RoleAssistant}, 1.0))

			latestGeneratedFiles := getGeneratedFileContent(moduleDir, path)
			for path, latest := range latestGeneratedFiles {
				original, ok := originalGeneratedFiles[path]
				if ok && original.WriteVerificationToken == latest.WriteVerificationToken {
					continue
				}
				latest.WriteVerificationToken = "" // Do not allow assistant to write to these files
				generatedFileUpdateJSON, err := json.Marshal(latest)
				if err != nil {
					return nil, errors.Wrap(err, "could not marshal read result for generated file")
				}
				content = append(content, annotateTextContent(mcp.NewTextContent(string(generatedFileUpdateJSON)), []mcp.Role{mcp.RoleAssistant}, 0.5))
			}
			return &mcp.CallToolResult{
				Content: content,
				IsError: false,
			}, nil
		}
}

func prettyDiff(path, original, latest string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, latest, false)
	diffs = dmp.DiffCleanupSemantic(diffs)
	includedLines := 10
	for i, d := range diffs {
		if d.Type != diffmatchpatch.DiffEqual {
			continue
		}
		lines := strings.Split(d.Text, "\n")
		if i == 0 {
			// First diff
			if len(diffs) == 1 {
				return "###CODEBLOCK### " + path + "\nNo changes were made."
			}
			if len(lines) > includedLines+1 {
				lines = append([]string{"<...>"}, lines[len(lines)-1-includedLines:]...)
			}
		} else if i == len(diffs)-1 {
			// Last diff
			if len(lines) > includedLines+1 {
				lines = append(lines[:includedLines], "<...>")
			}
		} else {
			// Middle diff
			if len(lines) > includedLines*2+1 {
				originalLines := slices.Clone(lines)
				lines = append(lines[:includedLines], "<...>")
				lines = append(lines, originalLines[len(originalLines)-1-includedLines:]...)
			}
		}
		d.Text = strings.Join(lines, "\n")
		diffs[i] = d
	}
	return "###CODEBLOCK###" + path + "\n\n" + dmp.DiffPrettyText(diffs) + "\n\n<\\...>\n"

}

func tokenForFileContent(content []byte) (string, error) {
	hasher := sha256.New()
	if _, err := hasher.Write(content); err != nil {
		return "", errors.Wrap(err, "could not hash file content")
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func detectModulePath(path string) optional.Option[string] {
	path, err := filepath.Abs(path)
	if err != nil {
		return optional.None[string]()
	}
	for dir := filepath.Dir(path); dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "ftl.toml")); err == nil {
			return optional.Some(dir)
		}
	}
	return optional.None[string]()
}

// getGeneratedFileContent returns the content of generated files that are relevant to the file being written.
//
// Currently this is hardcoded to only support `*.sql` -> `queries.ftl.go`.
func getGeneratedFileContent(moduleDir string, writePath string) map[string]*readResult {
	// TODO: make this generic with language plugin support
	generatedPaths := []string{}
	if strings.HasSuffix(writePath, ".sql") {
		generatedPaths = append(generatedPaths, "queries.ftl.go")
	}
	out := map[string]*readResult{}
	for _, path := range generatedPaths {
		path = filepath.Join(moduleDir, path)
		fileContent, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		token, err := tokenForFileContent(fileContent)
		if err != nil {
			continue
		}
		out[path] = &readResult{
			FileContent:            string(fileContent),
			WriteVerificationToken: token,
			Explanation:            fmt.Sprintf("Generated file has changed at %s", path),
		}
	}
	return out
}
