package profiles

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
)

func TestProfile(t *testing.T) {
	root := t.TempDir()

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	projectConfig := ProjectConfig{
		Realm:         "test",
		FTLMinVersion: ftl.Version,
		ModuleRoots:   []string{"."},
	}
	sr := config.NewSecretsRegistry(None[adminpbconnect.AdminServiceClient]())
	cr := config.NewConfigurationRegistry(None[adminpbconnect.AdminServiceClient]())

	_, err := Init(root, projectConfig, sr, cr)
	assert.NoError(t, err)

	project, err := Open(root, sr, cr)
	assert.NoError(t, err)

	err = project.New(NewProfileConfig{
		Name: "sandbox",
		Config: LocalProfileConfig{
			SecretsProvider: config.NewProfileProviderKey("sandbox"),
			ConfigProvider:  config.NewProfileProviderKey("sandbox"),
		},
	})
	assert.NoError(t, err)

	profile, err := project.ActiveProfile(ctx)
	assert.NoError(t, err)

	assert.Equal(t, "local", profile.Name())
	assert.Equal(t, must.Get(url.Parse("http://localhost:8892")), profile.Endpoint())

	assert.Equal(t, ProjectConfig{
		root:           root,
		Realm:          "test",
		FTLMinVersion:  ftl.Version,
		ModuleRoots:    []string{"."},
		DefaultProfile: "local",
	}, profile.ProjectConfig())

	cm := profile.ConfigurationManager()
	passwordKey := config.NewRef(Some("echo"), "password")
	err = config.Store(ctx, cm, passwordKey, "hello")
	assert.NoError(t, err)

	passwordValue, err := config.Load[string](ctx, cm, passwordKey)
	assert.NoError(t, err)

	assert.Equal(t, "hello", passwordValue)

	err = project.Switch("sandbox", true)
	assert.NoError(t, err)

	profile, err = project.ActiveProfile(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "sandbox", profile.Name())
}
