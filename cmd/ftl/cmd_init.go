package main

import (
	"archive/zip"
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/block/scaffolder"

	"github.com/block/ftl"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/profiles"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/projectinit"
)

//go:embed dependency-versions.txt
var userHermitPackages string

type initCmd struct {
	Name        string   `arg:"" help:"Name of the project."`
	Dir         string   `arg:"" optional:"" help:"Directory to initialize the project in. If not specified, creates a new directory with the project name."`
	Hermit      bool     `help:"Include Hermit language-specific toolchain binaries." negatable:"" default:"true"`
	ModuleDirs  []string `help:"Child directories of existing modules."`
	ModuleRoots []string `help:"Root directories of existing modules."`
	Git         bool     `help:"Commit generated files to git." negatable:"" default:"true"`
	Startup     string   `help:"Command to run on startup."`
}

func (i initCmd) Help() string {
	return `
Examples:
  ftl init myproject        # Creates a new folder named "myproject" and initializes it
  ftl init myproject .      # Initializes the current directory as "myproject"
  ftl init myproject custom # Creates a folder named "custom" and initializes it as "myproject"`
}

func (i initCmd) Run(
	ctx context.Context,
	logger *log.Logger,
	configRegistry *config.Registry[config.Configuration],
	secretsRegistry *config.Registry[config.Secrets],
) error {
	// If the directory is not specified, use the project name as the directory name.
	if i.Dir == "" {
		i.Dir = i.Name
	}

	logger.Debugf("Initializing FTL project in %s", i.Dir)
	if err := os.MkdirAll(i.Dir, 0750); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	if err := scaffold(ctx, i.Hermit, projectinit.Files(), i.Dir, i); err != nil {
		return errors.WithStack(err)
	}

	config := projectconfig.Config{
		Name:          i.Name,
		Hermit:        i.Hermit,
		NoGit:         !i.Git,
		FTLMinVersion: ftl.Version,
		ModuleDirs:    i.ModuleDirs,
		Commands: projectconfig.Commands{
			Startup: []string{i.Startup},
		},
	}
	if err := projectconfig.Create(ctx, config, i.Dir); err != nil {
		return errors.WithStack(err)
	}

	_, err := profiles.Init(profiles.ProjectConfig{
		Realm:         i.Name,
		FTLMinVersion: ftl.Version,
		ModuleRoots:   i.ModuleRoots,
		Git:           !i.Git,
		Root:          i.Dir,
	}, secretsRegistry, configRegistry)
	if err != nil {
		return errors.Wrap(err, "initialize project")
	}
	if i.Hermit {
		if err := installHermitFTL(ctx, i.Dir); err != nil {
			return errors.Wrap(err, "initialize Hermit FTL")
		}
	}

	if i.Git {
		err := maybeGitInit(ctx, i.Dir)
		if err != nil {
			return errors.Wrap(err, "running git init")
		}
		logger.Debugf("Updating .gitignore")
		if err := updateGitIgnore(ctx, i.Dir); err != nil {
			return errors.Wrap(err, "update .gitignore")
		}
		if err := maybeGitAdd(ctx, i.Dir, ".ftl-project"); err != nil {
			return errors.Wrap(err, "git add .ftl-project")
		}
		if err := maybeGitAdd(ctx, i.Dir, "ftl-project.toml"); err != nil {
			return errors.Wrap(err, "git add ftl-project.toml")
		}
		if err := maybeGitAdd(ctx, i.Dir, "README.md"); err != nil {
			return errors.Wrap(err, "git add README.md")
		}
		if i.Hermit {
			if err := maybeGitAdd(ctx, i.Dir, "bin"); err != nil {
				return errors.Wrap(err, "git add bin")
			}
		}
	}

	fmt.Printf("Successfully created FTL project '%s' in %s\n\n", i.Name, i.Dir)
	fmt.Printf("To get started:\n")
	if i.Dir != "." {
		fmt.Printf("  cd %s\n", i.Dir)
	}
	fmt.Printf("  ftl dev\n")
	return nil
}

func maybeGitAdd(ctx context.Context, dir string, paths ...string) error {
	args := append([]string{"add"}, paths...)
	if err := exec.Command(ctx, log.Debug, dir, "git", args...).RunBuffered(ctx); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func installHermitFTL(ctx context.Context, dir string) error {
	for _, install := range strings.Fields(userHermitPackages) {
		args := []string{"install", install}
		if err := exec.Command(ctx, log.Debug, dir, "./bin/hermit", args...).RunBuffered(ctx); err != nil {
			return errors.Wrapf(err, "unable to install hermit package %s", install)
		}
	}
	ftlVersion := ftl.Version

	if !ftl.IsRelease(ftlVersion) {
		ftlVersion = "@latest"
	} else {
		ftlVersion = "-" + ftlVersion
	}
	args := []string{"install", "ftl" + ftlVersion}
	if err := exec.Command(ctx, log.Debug, dir, "./bin/hermit", args...).RunBuffered(ctx); err != nil {
		return errors.Wrap(err, "unable to install hermit package ft")
	}
	return nil
}

func maybeGitInit(ctx context.Context, dir string) error {
	args := []string{"init"}
	if err := exec.Command(ctx, log.Debug, dir, "git", args...).RunBuffered(ctx); err != nil {
		return errors.Wrap(err, "git init")
	}
	return nil
}

func updateGitIgnore(ctx context.Context, gitRoot string) error {
	f, err := os.OpenFile(path.Join(gitRoot, ".gitignore"), os.O_RDWR|os.O_CREATE, 0644) //nolint:gosec
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close() //nolint:gosec

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "**/.ftl" {
			return nil
		}
	}

	if scanner.Err() != nil {
		return errors.WithStack(scanner.Err())
	}

	// append if not already present
	if _, err = f.WriteString("**/.ftl\n"); err != nil {
		return errors.WithStack(err)
	}

	// Add .gitignore to git
	return errors.WithStack(maybeGitAdd(ctx, gitRoot, ".gitignore"))
}

func scaffold(ctx context.Context, includeBinDir bool, source *zip.Reader, destination string, sctx any, options ...scaffolder.Option) error {
	logger := log.FromContext(ctx)
	opts := []scaffolder.Option{scaffolder.Exclude("^go.mod$")}
	if !includeBinDir {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}
	opts = append(opts, options...)
	if err := internal.ScaffoldZip(source, destination, sctx, opts...); err != nil {
		return errors.Wrap(err, "failed to scaffold")
	}
	return nil
}
