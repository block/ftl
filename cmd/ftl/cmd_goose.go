package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/terminal"
)

//go:embed goose_instructions_intro.txt
var gooseIntroInstructions string

//go:embed goose_instructions_first_prompt.txt
var gooseFirstPromptInstructions string

type gooseCmd struct {
	Prompt []string `arg:"" required:"" help:"Ask Goose for help"`
}

var first = true

func (c *gooseCmd) Run(ctx context.Context) error {
	gooseName := "🔮 Goose"
	terminal.UpdateModuleState(ctx, gooseName, terminal.BuildStateBuilding)
	defer terminal.UpdateModuleState(ctx, gooseName, terminal.BuildStateTerminated)

	logger := log.FromContext(ctx)

	data := strings.Join(c.Prompt, " ")
	args := []string{"run"}
	if first {
		logger.Infof("Setting up goose")
		first = false

		// Run introduction instructions
		// This is separate from the next command because goose may skip an instruction if it is part of a larger input.
		docs, err := downloadDocs(ctx)
		if err != nil {
			return err
		}

		help, err := exec.Capture(ctx, ".", "ftl", "dump-help",
			"--ignored-commands",
			"ping,status,init,profile,ps,bench,replay,update,kill,schema diff,schema generate,build,deploy,download,release,lsp,goose,interactive,dev,serve,completion",
			"--ignored-flags",
			"help,endpoint,provisioner-endpoint,timeline-endpoint,lease-endpoint,admin-endpoint,schema-endpoint,authenticators,log-level,log-json,log-timestamps,log-color,plain,opvault,insecure,version,config")
		if err != nil {
			return fmt.Errorf("failed to dump help: %w", err)
		}

		fullIntroInstructions := gooseIntroInstructions + "\n\nOutput of all FTL commands with --help\n\n" + string(help) + "\n\nAll FTL Docs:\n\n" + strings.Join(docs, "\n\n")

		cmd := exec.Command(ctx, log.Debug, ".", "goose", "run", "--text", fullIntroInstructions)
		out := &output{}
		cmd.Stderr = out
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("goose failed: %w", err)
		}

		// Second command includes final instructions and the user's input
		data = gooseFirstPromptInstructions + data
	}
	args = append(args, "--resume", "--text", data)

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

var _ io.Writer = &output{}

type output struct {
}

func (o output) Write(p []byte) (n int, err error) {
	if !strings.HasPrefix(string(p), "Closing session.") {
		fmt.Printf("%s", string(p))
	}
	return len(p), nil
}
