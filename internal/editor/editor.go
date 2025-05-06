package editor

import (
	"context"
	"fmt"
	"net/url"
	osexec "os/exec" //nolint:depguard
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
)

const (
	VSCode   = "vscode"
	IntelliJ = "intellij"
	Cursor   = "cursor"
	Zed      = "zed"
	GitHub   = "github"
)

var SupportedEditors = []string{VSCode, IntelliJ, Cursor, Zed, GitHub}

// OpenFileInEditor opens the given file path at the specified position in the selected editor.
func OpenFileInEditor(ctx context.Context, editor string, pos schema.Position, projectRoot string) error {
	executable, args, workDir, err := getEditorCmdArgs(ctx, editor, pos, projectRoot)
	if err != nil {
		return err
	}

	err = exec.Command(ctx, log.Debug, workDir, executable, args...).RunBuffered(ctx)
	if err != nil {
		return errors.Wrapf(err, "could not open file in %s using command '%s %s'", editor, executable, strings.Join(args, " "))
	}
	return nil
}

func getEditorCmdArgs(ctx context.Context, editor string, pos schema.Position, projectRoot string) (executable string, args []string, workDir string, err error) {
	switch editor {
	case VSCode:
		executable = "code"
		path := formatPathWithPosition(pos, true)
		args = []string{projectRoot, "--goto", path}
		workDir = "."

	case IntelliJ:
		executable = "idea"
		args = []string{projectRoot, "--line", strconv.Itoa(pos.Line), "--column", strconv.Itoa(pos.Column), pos.Filename}
		workDir = "."

	case Cursor:
		executable = "cursor"
		path := formatPathWithPosition(pos, true)
		args = []string{projectRoot, "--goto", path}
		workDir = "."

	case Zed:
		executable = "zed"
		path := formatPathWithPosition(pos, true)
		args = []string{path}
		workDir = projectRoot

	case GitHub:
		var err error
		executable, args, workDir, err = buildGitHubLinkCommand(ctx, pos, projectRoot)
		if err != nil {
			return "", nil, "", errors.WithStack(err)
		}

	default:
		err = errors.Errorf("unsupported editor %q, expected one of %s", editor, strings.Join(SupportedEditors, ", "))
	}
	return executable, args, workDir, err
}

func formatPathWithPosition(pos schema.Position, includeColumn bool) string {
	path := pos.Filename + ":" + fmt.Sprint(pos.Line)
	if includeColumn && pos.Column > 0 {
		path += ":" + fmt.Sprint(pos.Column)
	}
	return path
}

// buildGitHubLinkCommand constructs the command needed to open a file link on GitHub.
// It determines the git repository root, remote origin URL, commit hash, and relative file path
// to construct the appropriate GitHub blob URL.
func buildGitHubLinkCommand(ctx context.Context, pos schema.Position, projectRoot string) (executable string, args []string, workDir string, err error) {
	absFilename := filepath.Join(projectRoot, pos.Filename)

	repoRootCmd := osexec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	repoRootCmd.Dir = filepath.Dir(absFilename)
	repoRootBytes, err := repoRootCmd.Output()
	if err != nil {
		return "", nil, "", errors.Wrapf(err, "failed to get git repo root for %q: ensure it is in a git repository and git is installed", absFilename)
	}
	repoRoot := strings.TrimSpace(string(repoRootBytes))

	remoteURLCmd := osexec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
	remoteURLCmd.Dir = repoRoot
	remoteURLBytes, err := remoteURLCmd.Output()
	if err != nil {
		// Attempt to find remote URL from the file's directory if not found at root (might be a submodule)
		remoteURLCmd.Dir = filepath.Dir(absFilename)
		remoteURLBytes, err = remoteURLCmd.Output()
		if err != nil {
			return "", nil, "", errors.Wrap(err, "failed to get git remote origin URL: ensure 'origin' remote is configured")
		}
	}
	remoteURL := strings.TrimSpace(string(remoteURLBytes))

	commitHashCmd := osexec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	commitHashCmd.Dir = repoRoot
	commitHashBytes, err := commitHashCmd.Output()
	if err != nil {
		return "", nil, "", errors.Wrap(err, "failed to get git commit hash")
	}
	commitHash := strings.TrimSpace(string(commitHashBytes))

	org, repo, err := parseGitHubRemoteURL(remoteURL)
	if err != nil {
		return "", nil, "", errors.Wrapf(err, "could not parse GitHub remote URL %q", remoteURL)
	}

	relativePath, err := filepath.Rel(repoRoot, absFilename)
	if err != nil {
		return "", nil, "", errors.Wrapf(err, "failed to get relative path for %q from root %q", absFilename, repoRoot)
	}

	githubURL := fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s#L%d",
		org, repo, commitHash, relativePath, pos.Line)

	switch runtime.GOOS {
	case "darwin":
		executable = "open"
	case "linux":
		executable = "xdg-open"
	default:
		return "", nil, "", errors.Errorf("unsupported operating system for github editor: %s", runtime.GOOS)
	}
	args = []string{githubURL}
	workDir = "."
	return executable, args, workDir, nil
}

func parseGitHubRemoteURL(remoteURL string) (org string, repo string, err error) {
	parsedURL, err := url.Parse(remoteURL)
	if err == nil && (parsedURL.Scheme == "https" || parsedURL.Scheme == "http") && strings.Contains(parsedURL.Host, "github.com") {
		parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = strings.TrimSuffix(parts[1], ".git")
			return org, repo, nil
		}
	}

	sshRegex := regexp.MustCompile(`(?:git@|ssh://git@)github\.com[:/]([^/]+)/([^/]+?)(\.git)?$`)
	matches := sshRegex.FindStringSubmatch(remoteURL)
	if len(matches) >= 3 {
		org = matches[1]
		repo = matches[2]
		return org, repo, nil
	}

	return "", "", errors.Errorf("unsupported GitHub remote URL format: %s", remoteURL)
}
