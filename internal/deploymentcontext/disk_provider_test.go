package deploymentcontext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"google.golang.org/protobuf/proto"
)

func TestDiskProvider(t *testing.T) {
	tempDir := t.TempDir()

	// Setup secrets directory and file
	secretsDir := filepath.Join(tempDir, "secrets")
	err := os.Mkdir(secretsDir, 0700)
	assert.NoError(t, err)
	secretFilePath := filepath.Join(secretsDir, "test-secret")
	err = os.WriteFile(secretFilePath, []byte("secret-value"), 0600)
	assert.NoError(t, err)

	// Setup configs directory and file
	configsDir := filepath.Join(tempDir, "configs")
	err = os.Mkdir(configsDir, 0700)
	assert.NoError(t, err)
	configFilePath := filepath.Join(configsDir, "test-config")
	err = os.WriteFile(configFilePath, []byte("config-value"), 0600)
	assert.NoError(t, err)

	testSchema := &schema.Schema{
		Realms: []*schema.Realm{{
			Name: "testrealm",
			Modules: []*schema.Module{
				schema.Builtins(),
				{
					Name: "testmodule",
					Decls: []schema.Decl{
						&schema.Data{
							Name: "PathParams",
							Fields: []*schema.Field{
								{Name: "name", Type: &schema.String{}},
							},
							Visibility: schema.VisibilityScopeModule,
						},
						&schema.Data{
							Name: "Response",
							Fields: []*schema.Field{
								{Name: "message", Type: &schema.String{}},
							},
							Visibility: schema.VisibilityScopeModule,
						},
						&schema.Verb{
							Name: "hello",
							Request: &schema.Ref{
								Module: "builtin",
								Name:   "HttpRequest",
								TypeParameters: []schema.Type{
									&schema.Unit{},
									&schema.Ref{Module: "testmodule", Name: "PathParams"},
									&schema.Unit{},
								},
							},
							Response: &schema.Ref{
								Module: "builtin",
								Name:   "HttpResponse",
								TypeParameters: []schema.Type{
									&schema.Ref{Module: "testmodule", Name: "Response"},
									&schema.Unit{},
								},
							},
							Metadata: []schema.Metadata{
								&schema.MetadataIngress{
									Type:   "GET",
									Method: "GET",
									Path: []schema.IngressPathComponent{
										&schema.IngressPathLiteral{Text: "hello"},
										&schema.IngressPathParameter{Name: "name"},
									},
								},
								&schema.MetadataCalls{
									Calls: []*schema.Ref{{Module: "othermodule", Name: "Greeting"}},
								},
							},
							Visibility: schema.VisibilityScopeModule,
						},
						&schema.Database{
							Name: "testdb",
							Type: "postgres",
						},
						&schema.Verb{
							Name:       "sendMessage",
							Request:    &schema.String{},
							Response:   &schema.String{},
							Visibility: schema.VisibilityScopeRealm,
							Runtime: &schema.VerbRuntime{
								EgressRuntime: &schema.EgressRuntime{
									Targets: []schema.EgressTarget{
										{Expression: "external-service", Target: "http://external-service"},
									},
								},
							},
						},
					},
				},
				{
					Name: "othermodule",
					Decls: []schema.Decl{
						&schema.Verb{
							Name:       "Greeting",
							Request:    &schema.String{},
							Response:   &schema.String{},
							Visibility: schema.VisibilityScopeRealm,
						},
					},
				},
			},
		}},
	}

	schemaPath := filepath.Join(tempDir, "schema.ftl")
	schemaBytes, err := proto.Marshal(testSchema.ToProto())
	assert.NoError(t, err)
	err = os.WriteFile(schemaPath, schemaBytes, 0600)
	assert.NoError(t, err)

	t.Run("test providers", func(t *testing.T) {
		ctx := log.ContextWithLogger(t.Context(), log.Configure(os.Stderr, log.Config{Level: log.Info}))

		secretsProvider := NewDiskProvider(secretsDir)
		secretsMap, err := secretsProvider(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "secret-value", string(secretsMap["test-secret"]))

		configProvider := NewDiskProvider(configsDir)
		configsMap, err := configProvider(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "config-value", string(configsMap["test-config"]))
	})
}
