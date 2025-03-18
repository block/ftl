package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

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
				return nil, fmt.Errorf("path is required")
			}
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("could not read file: %w", err)
			}
			token, err := tokenForFileContent(fileContent)
			if err != nil {
				return nil, fmt.Errorf("could not generate verification token: %w", err)
			}
			readResult, err := newReadResult(fileContent, token, false, "")
			if err != nil {
				return nil, fmt.Errorf("could not create read result: %w", err)
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
		return nil, fmt.Errorf("could not marshal read result: %w", err)
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
	Status               statusOutput `json:"status,omitempty"`
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
				return nil, fmt.Errorf("path is required")
			}
			fileContent, ok := request.Params.Arguments["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content is required")
			}
			// TODO: validate if path is allowed

			originalContent, err := os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					originalContent = []byte{}
				} else {
					return nil, fmt.Errorf("could access existing file: %w", err)
				}
			}
			if len(originalContent) > 0 {
				token, err := tokenForFileContent(originalContent)
				if err != nil {
					return nil, fmt.Errorf("could not generate verification token: %w", err)
				}
				expectedToken, ok := request.Params.Arguments["verificationToken"].(string)
				if !ok || expectedToken != token {
					// File was not read (or file has changed). Return an error response with an explanation and the original content
					return newReadResult(originalContent, token, true, `The file was not read before it was written or it has changed since it was last read.
		The file has been read and provided here. Make sure you understand the existing content before using the Write tool so you do not accidentally alter data or code that you did not mean to change.`)
				}

			}
			if err := os.WriteFile(path, []byte(fileContent), 0600); err != nil {
				return nil, fmt.Errorf("could not write file: %w", err)
			}
			var userResult mcp.TextContent
			if len(originalContent) == 0 {
				userResult = annotateTextContent(mcp.NewTextContent("### "+path+"\n```\n"+fileContent+"\n```"), []mcp.Role{mcp.RoleUser}, 0.2)
			} else {
				userResult = annotateTextContent(mcp.NewTextContent(diff(path, string(originalContent), fileContent)), []mcp.Role{mcp.RoleUser}, 0.2)
			}

			content := []mcp.Content{
				userResult,
			}

			assistantResult := writeResult{}
			assistantResult.NewVerificationToken, err = tokenForFileContent([]byte(fileContent))
			if err == nil {
				assistantResult.TokenExplanation = "The file has been updated. A new verification token is provided if you need to update the file again."
			}

			if status, err := getStatusOutput(serverCtx, buildEngineClient, adminClient); err == nil {
				assistantResult.StatusExplanation = "The FTL status after the change is also provided."
				assistantResult.Status = status
			}
			assistantResultJSON, err := json.Marshal(assistantResult)
			if err != nil {
				return nil, fmt.Errorf("could not marshal assistant result: %w", err)
			}
			content = append(content, annotateTextContent(mcp.NewTextContent(string(assistantResultJSON)), []mcp.Role{mcp.RoleAssistant}, 1.0))
			return &mcp.CallToolResult{
				Content: content,
				IsError: false,
			}, nil
		}
}

func diff(path, original, new string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(original), new, false)
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
				return "### " + path + "\nNo changes were made."
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
	return "###" + path + "\n\n" + dmp.DiffPrettyText(diffs) + "\n"

}

func tokenForFileContent(content []byte) (string, error) {
	hasher := sha256.New()
	if _, err := hasher.Write(content); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
