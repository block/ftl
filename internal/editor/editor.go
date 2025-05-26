package editor

import (
	"context"
	"fmt"
	"net/url" //nolint:depguard
	"regexp"
	"runtime"
	"strconv"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/exec"
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
// It now requires the overall schema to look up module-specific Git metadata.
func OpenFileInEditor(ctx context.Context, editor string, pos schema.Position, projectRoot string, module *schema.Module) error {
	executable, args, workDir, err := getEditorCmdArgs(editor, pos, projectRoot, module)
	if err != nil {
		return err
	}

	err = exec.Command(ctx, log.Debug, workDir, executable, args...).RunBuffered(ctx)
	if err != nil {
		return errors.Wrapf(err, "could not open file in %s using command '%s %s'", editor, executable, strings.Join(args, " "))
	}
	return nil
}

func getEditorCmdArgs(editor string, pos schema.Position, projectRoot string, module *schema.Module) (executable string, args []string, workDir string, err error) {
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
		executable, args, workDir, err = buildGitHubLinkCommand(pos, module)
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
// to construct the appropriate GitHub blob URL, using module schema for Git information.
func buildGitHubLinkCommand(pos schema.Position, module *schema.Module) (executable string, args []string, workDir string, err error) {
	if module == nil {
		return "", nil, "", errors.Errorf("module information is required to build GitHub link")
	}

	var commitHash string
	var gitSchemaMeta *schema.MetadataGit

	for _, metaItem := range module.Metadata {
		if concreteGitMeta, ok := metaItem.(*schema.MetadataGit); ok {
			gitSchemaMeta = concreteGitMeta
			break
		}
	}

	if gitSchemaMeta == nil || gitSchemaMeta.Repository == "" {
		return "", nil, "", errors.Errorf("Git metadata (repository URL) not found in module %q schema", module.Name)
	}
	remoteURL := gitSchemaMeta.Repository

	if gitSchemaMeta.Commit == "" {
		return "", nil, "", errors.Errorf("Git commit hash not found in module %q schema", module.Name)
	}
	commitHash = gitSchemaMeta.Commit

	org, repo, err := parseGitHubRemoteURL(remoteURL)
	if err != nil {
		return "", nil, "", errors.Wrapf(err, "could not parse GitHub remote URL %q", remoteURL)
	}

	// Use pos.Filename directly, assuming it is relative to the repo root
	githubURL := fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s#L%d",
		org, repo, commitHash, pos.Filename, pos.Line)

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
