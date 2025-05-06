package app

import (
	"context"

	errors "github.com/alecthomas/errors"
	"github.com/block/ftl/internal/projectconfig"
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

func (c *realmAddCmd) Run(ctx context.Context, p projectconfig.Config) error {
	realm, commitSHA, err := realm.LatestExternalRealm(ctx, projectconfig.ExternalRealmConfig{
		GitRepo:   c.GitRepo,
		GitBranch: c.GitBranch,
		GitPath:   c.GitPath,
	})
	if err != nil {
		return errors.Wrap(err, "failed to fetch external realm")
	}

	if p.ExternalRealms == nil {
		p.ExternalRealms = make(map[string]projectconfig.ExternalRealmConfig)
	}
	p.ExternalRealms[realm.Name] = projectconfig.ExternalRealmConfig{
		GitRepo:   c.GitRepo,
		GitBranch: c.GitBranch,
		GitPath:   c.GitPath,
		GitCommit: commitSHA,
	}
	if err := projectconfig.Save(p); err != nil {
		return errors.Wrap(err, "failed to save project config")
	}
	return nil
}
