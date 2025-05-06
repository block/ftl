package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
)

type Git interface {
	Init(ctx context.Context, repo string) error
	Fetch(ctx context.Context, branch string, shallow bool) error
	Checkout(ctx context.Context, ref string) error
	GetCommitSHA(ctx context.Context) (string, error)
	ReadFile(ctx context.Context, path string) ([]byte, error)
}

type gitCLI struct {
	dir string
}

func NewGitCLI(dir string) Git {
	return &gitCLI{dir: dir}
}

func (g *gitCLI) Init(ctx context.Context, repo string) error {
	if err := exec.Command(ctx, log.Debug, g.dir, "git", "init", ".").Run(); err != nil {
		return errors.Wrapf(err, "failed to init git repo")
	}
	if err := exec.Command(ctx, log.Debug, g.dir, "git", "remote", "add", "origin", repo).Run(); err != nil {
		return errors.Wrap(err, "failed to add remote to git repo")
	}
	return nil
}

func (g *gitCLI) Fetch(ctx context.Context, branch string, shallow bool) error {
	if shallow {
		if err := exec.Command(ctx, log.Debug, g.dir, "git", "fetch", "--depth", "1", "origin", branch).Run(); err != nil {
			return errors.Wrap(err, "failed to fetch shallow git repo")
		}
	} else {
		if err := exec.Command(ctx, log.Debug, g.dir, "git", "fetch", "origin", branch).Run(); err != nil {
			return errors.Wrap(err, "failed to fetch git repo")
		}
	}
	return nil
}

func (g *gitCLI) Checkout(ctx context.Context, ref string) error {
	if err := exec.Command(ctx, log.Debug, g.dir, "git", "checkout", ref).Run(); err != nil {
		return errors.Wrapf(err, "failed to checkout git ref %s", ref)
	}
	return nil
}

func (g *gitCLI) GetCommitSHA(ctx context.Context) (string, error) {
	commitSHA, err := exec.Capture(ctx, g.dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", errors.Wrap(err, "failed to get commit SHA")
	}
	return strings.TrimSpace(string(commitSHA)), nil
}

func (g *gitCLI) ReadFile(ctx context.Context, path string) ([]byte, error) {
	content, err := os.ReadFile(filepath.Join(g.dir, path))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read external schema")
	}
	return content, nil
}
