package main

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/internal/profiles"
	"github.com/block/ftl/internal/realm"
)

type realmCmd struct {
	Add realmAddCmd `cmd:"" help:"Add a new external realm."`
}

type realmAddCmd struct {
	GitRepo   string `help:"The git repository to use for the realm." required:""`
	GitBranch string `help:"The git branch to use for the realm." required:""`
	GitPath   string `help:"The path to the schema file in the git repository." default:"ftl-external-schema.json"`
}

func (c *realmAddCmd) Run(ctx context.Context, project *profiles.Project) error {
	realm, commitSHA, err := realm.LatestExternalRealm(ctx, profiles.ExternalRealmConfig{
		GitRepo:   c.GitRepo,
		GitBranch: c.GitBranch,
		GitPath:   c.GitPath,
	})
	if err != nil {
		return errors.Wrap(err, "failed to fetch external realm")
	}

	projectConfig := project.Config()
	if projectConfig.ExternalRealms == nil {
		projectConfig.ExternalRealms = make(map[string]profiles.ExternalRealmConfig)
	}
	projectConfig.ExternalRealms[realm.Name] = profiles.ExternalRealmConfig{
		GitRepo:   c.GitRepo,
		GitBranch: c.GitBranch,
		GitPath:   c.GitPath,
		GitCommit: commitSHA,
	}
	if err := project.Reconfigure(projectConfig); err != nil {
		return errors.Wrap(err, "failed to save project config")
	}
	return nil
}
