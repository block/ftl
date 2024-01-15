package main

import (
	"bufio"
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type moduleFolderInfo struct {
	NumFiles    int
	LastModTime time.Time
}

type devCmd struct {
	BaseDir      string        `arg:"" help:"Directory to watch for FTL modules" type:"existingdir" default:"."`
	Watch        time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	FailureDelay time.Duration `help:"Delay before retrying a failed deploy." default:"5s"`
	modules      map[string]moduleFolderInfo
	client       ftlv1connect.ControllerServiceClient
}

func (d *devCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)
	logger.Infof("Watching %s for FTL modules", d.BaseDir)

	d.modules = make(map[string]moduleFolderInfo)
	d.client = client

	lastScanTime := time.Now()
	for {
		delay := d.Watch
		iterationStartTime := time.Now()

		tomls, err := d.getTomls(ctx)
		if err != nil {
			return err
		}

		d.addOrRemoveModules(tomls)

		for dir := range d.modules {
			currentModule := d.modules[dir]
			err := d.updateFileInfo(ctx, dir)
			if err != nil {
				return err
			}

			if currentModule.NumFiles != d.modules[dir].NumFiles || d.modules[dir].LastModTime.After(lastScanTime) {
				deploy := deployCmd{
					Replicas:  1,
					ModuleDir: dir,
				}
				err = deploy.Run(ctx, client)
				if err != nil {
					logger.Errorf(err, "Error deploying module %s. Will retry", dir)
					delete(d.modules, dir)
					// Increase delay when there's a compile failure.
					delay = d.FailureDelay
				}
			}
		}

		lastScanTime = iterationStartTime
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil
		}
	}
}

func (d *devCmd) getTomls(ctx context.Context) ([]string, error) {
	baseDir := d.BaseDir
	ignores := initGitIgnore(ctx, baseDir)
	tomls := []string{}

	err := walkDir(baseDir, ignores, func(srcPath string, d fs.DirEntry) error {
		if filepath.Base(srcPath) == "ftl.toml" {
			tomls = append(tomls, srcPath)
			return errSkip // Return errSkip to stop recursion in this branch
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tomls, nil
}

func (d *devCmd) addOrRemoveModules(tomls []string) {
	for _, toml := range tomls {
		dir := filepath.Dir(toml)
		if _, ok := d.modules[dir]; !ok {
			d.modules[dir] = moduleFolderInfo{
				LastModTime: time.Now(),
			}
		}
	}

	for dir := range d.modules {
		found := false
		for _, toml := range tomls {
			if filepath.Dir(toml) == dir {
				found = true
				break
			}
		}
		if !found {
			delete(d.modules, dir) // Remove deleted module from d.modules
		}
	}
}

func (d *devCmd) updateFileInfo(ctx context.Context, dir string) error {
	config, err := moduleconfig.LoadConfig(dir)
	if err != nil {
		return err
	}

	ignores := initGitIgnore(ctx, dir)
	d.modules[dir] = moduleFolderInfo{}

	err = walkDir(dir, ignores, func(srcPath string, entry fs.DirEntry) error {
		for _, pattern := range config.Watch {
			relativePath, err := filepath.Rel(dir, srcPath)
			if err != nil {
				return err
			}

			match, err := doublestar.PathMatch(pattern, relativePath)
			if err != nil {
				return err
			}

			if match && !entry.IsDir() {
				fileInfo, err := entry.Info()
				if err != nil {
					return err
				}

				module := d.modules[dir]
				module.NumFiles++
				if fileInfo.ModTime().After(module.LastModTime) {
					module.LastModTime = fileInfo.ModTime()
				}
				d.modules[dir] = module
			}
		}

		return nil
	})

	return err
}

// errSkip is returned by walkDir to skip a file or directory.
var errSkip = errors.New("skip directory")

// Depth-first walk of dir executing fn after each entry.
func walkDir(dir string, ignores []string, fn func(path string, d fs.DirEntry) error) error {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if err = fn(dir, fs.FileInfoToDirEntry(dirInfo)); err != nil {
		if errors.Is(err, errSkip) {
			return nil
		}
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var dirs []os.DirEntry

	// Process files first, then recurse into directories.
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		// Check if the path matches any ignore pattern
		shouldIgnore := false
		for _, pattern := range ignores {
			match, err := doublestar.PathMatch(pattern, fullPath)
			if err != nil {
				return err
			}
			if match {
				shouldIgnore = true
				break
			}
		}

		if shouldIgnore {
			continue // Skip this entry
		}

		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			if err = fn(fullPath, entry); err != nil {
				if errors.Is(err, errSkip) {
					// If errSkip is found in a file, skip the remaining files in this directory
					return nil
				}
				return err
			}
		}
	}

	// Then, recurse into subdirectories
	for _, dirEntry := range dirs {
		dirPath := filepath.Join(dir, dirEntry.Name())
		ignores = append(ignores, loadGitIgnore(dirPath)...)
		if err := walkDir(dirPath, ignores, fn); err != nil {
			if errors.Is(err, errSkip) {
				return errSkip // Propagate errSkip upwards to stop this branch of recursion
			}
			return err
		}
	}
	return nil
}

func initGitIgnore(ctx context.Context, dir string) []string {
	ignore := []string{
		"**/.*",
		"**/.*/**",
	}
	home, err := os.UserHomeDir()
	if err == nil {
		ignore = append(ignore, loadGitIgnore(home)...)
	}
	gitRootBytes, err := exec.Capture(ctx, dir, "git", "rev-parse", "--show-toplevel")
	if err == nil {
		gitRoot := strings.TrimSpace(string(gitRootBytes))
		for current := dir; strings.HasPrefix(current, gitRoot); current = path.Dir(current) {
			ignore = append(ignore, loadGitIgnore(current)...)
		}
	}
	return ignore
}

func loadGitIgnore(dir string) []string {
	r, err := os.Open(path.Join(dir, ".gitignore"))
	if err != nil {
		return nil
	}
	ignore := []string{}
	lr := bufio.NewScanner(r)
	for lr.Scan() {
		line := lr.Text()
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' || line[0] == '!' { // We don't support negation.
			continue
		}
		if strings.HasSuffix(line, "/") {
			line = path.Join("**", line, "**/*")
		} else if !strings.ContainsRune(line, '/') {
			line = path.Join("**", line)
		}
		ignore = append(ignore, line)
	}
	return ignore
}
