package profiles_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/profiles"
)

func TestProfile(t *testing.T) {
	root := t.TempDir()

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	projectConfig := profiles.ProjectConfig{
		Root:          root,
		Realm:         "test",
		FTLMinVersion: ftl.Version,
		ModuleRoots:   []string{"."},
	}
	sr := config.NewSecretsRegistry(optional.None[adminpbconnect.AdminServiceClient]())
	sr.Register(config.NewMemoryProviderFactory[config.Secrets]())
	sr.Register(config.NewFileProviderFactory[config.Secrets]())
	cr := config.NewRegistry[config.Configuration]()
	cr.Register(config.NewMemoryProviderFactory[config.Configuration]())
	cr.Register(config.NewFileProviderFactory[config.Configuration]())

	_, err := profiles.Init(projectConfig, sr, cr)
	assert.NoError(t, err)

	project, err := profiles.Open(root, sr, cr)
	assert.NoError(t, err)

	profile, err := project.Load(ctx, "local")
	assert.NoError(t, err)

	assert.Equal(t, "local", profile.Name())
	assert.Equal(t, must.Get(url.Parse("http://localhost:8892")), profile.Endpoint())

	assert.Equal(t, profiles.ProjectConfig{
		Root:           root,
		Realm:          "test",
		FTLMinVersion:  ftl.Version,
		ModuleRoots:    []string{"."},
		DefaultProfile: "local",
	}, profile.ProjectConfig())

	cm := profile.ConfigurationManager()
	passwordKey := config.NewRef(optional.Some("echo"), "password")
	err = config.Store(ctx, cm, passwordKey, "hello")
	assert.NoError(t, err)

	passwordValue, err := config.Load[string](ctx, cm, passwordKey)
	assert.NoError(t, err)

	assert.Equal(t, "hello", passwordValue)
}
