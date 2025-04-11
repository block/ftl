package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/mcp"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/terminal"
)

//go:embed goose_instructions.txt
var gooseInstructions string

type gooseCmd struct {
	Chat  gooseChatCmd  `cmd:"" default:"withargs" help:"Ask Goose for help"`
	Reset gooseResetCmd `cmd:"" help:"Reset Goose's context"`
}

type gooseResetCmd struct {
}

func (c *gooseResetCmd) Run(projectConfig projectconfig.Config) error {
	logPath := logPath(projectConfig)
	if err := os.Remove(logPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove Goose logs: %w", err)
	}
	return nil
}

func logPath(projectConfig projectconfig.Config) string {
	return filepath.Join(projectConfig.Root(), ".ftl", "goose-logs.json")
}

type gooseChatCmd struct {
	Prompt []string `arg:"" required:"" help:"Prompt for Goose"`
}

func (c *gooseChatCmd) Run(ctx context.Context, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient) error {
	gooseName := "🔮 Goose"
	terminal.UpdateModuleState(ctx, gooseName, terminal.BuildStateBuilding)
	defer terminal.UpdateModuleState(ctx, gooseName, terminal.BuildStateTerminated)

	logger := log.FromContext(ctx)

	logPath := logPath(projectConfig)
	logExists := false
	if _, err := os.Stat(logPath); err == nil {
		logExists = true
	}
	var prompt string
	userPrompt := strings.Join(c.Prompt, " ")
	args := []string{"run", "--with-extension", "ftl mcp", "--path", logPath}
	if !logExists {
		var docs []string
		var status string
		wg := &errgroup.Group{}
		wg.Go(func() error {
			var err error
			docs, err = downloadDocs(ctx)
			return err
		})
		wg.Go(func() error {
			var err error
			statusObj, err := mcp.GetStatusOutput(ctx, buildEngineClient, adminClient)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}
			statusBytes, err := json.Marshal(statusObj)
			if err != nil {
				return fmt.Errorf("failed to marshal status: %w", err)
			}
			status = string(statusBytes)
			return nil
		})
		if err := wg.Wait(); err != nil {
			return err //nolint:wrapcheck
		}

		// Run introduction instructions
		// This is separate from the next command because goose may skip an instruction if it is part of a larger input.

		components := []string{
			"You are working with a system called FTL (Faster than Light) within an existing project. I want you to learn about FTL before I give you the user's prompt.",
			"All FTL Docs:",
			strings.Join(docs, "\n\n"),
		}

		// Only include Go runtime docs if they are available (Go may not be installed)
		goRuntimeDocs, err := getGoRuntimeDocs(ctx)
		if err != nil {
			logger.Debugf("Failed to get Go runtime docs (skipping this step): %v", err)
		} else {
			components = append(components, "FTL Go Package Docs:", goRuntimeDocs)
		}

		components = append(components,
			gooseInstructions,
			"The initial FTL Status has been automatically fetched:",
			status,
			"The user's prompt:",
			userPrompt,
		)
		prompt = strings.Join(components, "\n\n")
	} else {
		prompt = userPrompt
		args = append(args, "--resume")
	}
	args = append(args, "--text", prompt)

	cmd := exec.Command(ctx, log.Debug, ".", "goose", args...)
	out := &output{}
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("goose failed: %w", err)
	}
	return nil
}

func downloadDocs(ctx context.Context) ([]string, error) {
	baseReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/block/ftl/contents/docs/docs/reference", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(baseReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch docs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch docs: status code %d", resp.StatusCode)
	}

	var referenceList []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&referenceList)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	errGroup := &errgroup.Group{}
	pages := make(chan string, len(referenceList))
	for _, i := range referenceList {
		name, ok := i["name"].(string)
		if !ok || !strings.HasSuffix(name, ".md") {
			continue
		}
		errGroup.Go(func() error {
			urlPath, ok := i["download_url"].(string)
			if !ok {
				return fmt.Errorf("failed to parse response: %v", i)
			}
			contentReq, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			contentResp, err := http.DefaultClient.Do(contentReq)
			if err != nil {
				return fmt.Errorf("failed to fetch doc: %w", err)
			}
			defer contentResp.Body.Close()

			if contentResp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to fetch doc %s: status code %d", urlPath, contentResp.StatusCode)
			}

			body, err := io.ReadAll(contentResp.Body)
			if err != nil {
				return fmt.Errorf("failed to read doc body: %w", err)
			}

			pages <- string(body)
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch docs: %w", err)
	}
	pagesSlice := make([]string, 0, len(referenceList))
	close(pages)
	for p := range pages {
		pagesSlice = append(pagesSlice, p)
	}
	getSidebarPosition := func(s string) int {
		lines := strings.Split(s, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "sidebar_position:") {
				var pos int
				if _, err := fmt.Sscanf(line, "sidebar_position: %d", &pos); err != nil {
					return -1
				}
				return pos
			}
		}
		return 0
	}
	slices.SortFunc(pagesSlice, func(i, j string) int {
		return getSidebarPosition(i) - getSidebarPosition(j)
	})
	return pagesSlice, nil
}

func getGoRuntimeDocs(ctx context.Context) (string, error) {
	output, err := exec.Capture(ctx, ".", "go", "doc", "github.com/block/ftl/go-runtime/ftl")
	if err != nil {
		return "", fmt.Errorf("failed to get Go runtime docs: %w", err)
	}
	return string(output), nil
}

var _ io.Writer = &output{}

type output struct {
}

func (o output) Write(p []byte) (n int, err error) {
	if !strings.HasPrefix(string(p), "Closing session.") {
		fmt.Printf("%s", string(p))
	}
	return len(p), nil
}
