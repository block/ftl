package realm

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/git"
	"github.com/block/ftl/internal/projectconfig"
)

// LatestExternalRealm returns the latest external realm that the config refers to, with the commit SHA.
func LatestExternalRealm(ctx context.Context, cfg projectconfig.ExternalRealmConfig) (realm *schema.Realm, commitSHA string, err error) {
	tmpdir, err := os.MkdirTemp("", "ftl-realm")
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create temp directory")
	}
	defer os.RemoveAll(tmpdir)

	git := git.NewGitCLI(tmpdir)
	if err := git.Init(ctx, cfg.GitRepo); err != nil {
		return nil, "", errors.Wrap(err, "failed to init git repo")
	}
	if err := git.Fetch(ctx, cfg.GitBranch, true); err != nil {
		return nil, "", errors.Wrap(err, "failed to fetch git repo")
	}
	if err := git.Checkout(ctx, cfg.GitBranch); err != nil {
		return nil, "", errors.Wrap(err, "failed to checkout git commit")
	}
	commitSHA, err = git.GetCommitSHA(ctx)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get commit SHA")
	}

	content, err := git.ReadFile(ctx, cfg.GitPath)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to read "+cfg.GitPath)
	}
	realm, err = parseExternal(content)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse external realm")
	}
	return realm, commitSHA, nil
}

// GetExternalRealm returns the external realm that the config refers to, with the commit SHA.
// It will cache the realm in the given directory if it does not yet exist.
func GetExternalRealm(ctx context.Context, cacheDir, name string, cfg projectconfig.ExternalRealmConfig) (*schema.Realm, error) {
	logger := log.FromContext(ctx)

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create external realm dir")
	}

	var content []byte
	cachedFile := filepath.Join(cacheDir, cfg.GitCommit+".json")
	_, err := os.Stat(cachedFile)
	if errors.Is(err, os.ErrNotExist) {
		logger.Infof("Fetching external realm %s from %s (%s)", name, cfg.GitRepo, cfg.GitCommit) //nolint:forbidigo

		content, err = fetchExternalSchemaFile(ctx, cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch external realm")
		}

		if err := os.WriteFile(cachedFile, content, 0600); err != nil {
			return nil, errors.Wrap(err, "failed to write external realm to local cache")
		}
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to stat external realm")
	} else {
		content, err = os.ReadFile(cachedFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read external realm from cache")
		}
	}

	return errors.WithStack2(parseExternal(content))
}

func fetchExternalSchemaFile(ctx context.Context, cfg projectconfig.ExternalRealmConfig) (content []byte, err error) {
	tmpdir, err := os.MkdirTemp("", "ftl-external-realm")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create temp directory")
	}
	defer os.RemoveAll(tmpdir)

	git := git.NewGitCLI(tmpdir)
	if err := git.Init(ctx, cfg.GitRepo); err != nil {
		return nil, errors.Wrap(err, "failed to init git repo")
	}
	if err := git.Fetch(ctx, cfg.GitBranch, true); err != nil {
		return nil, errors.Wrap(err, "failed to fetch git repo")
	}
	if err := git.Checkout(ctx, cfg.GitCommit); err != nil {
		return nil, errors.Wrap(err, "failed to checkout git branch")
	}

	content, err = git.ReadFile(ctx, cfg.GitPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read "+cfg.GitPath)
	}
	return content, nil
}

func parseExternal(content []byte) (*schema.Realm, error) {
	proto := &schemapb.Schema{}
	if err := protojson.Unmarshal(content, proto); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schema")
	}
	sch, err := schema.FromProto(proto)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schema")
	}
	if len(sch.InternalRealms()) != 1 {
		return nil, errors.New("expected exactly one internal realm")
	}
	result := sch.InternalRealms()[0]
	result.External = true
	return result, nil
}
