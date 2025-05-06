package languageplugin

import (
	"context"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/atomic"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/moduleconfig"
)

type testBuildContext struct {
	BuildContext
	IsRebuild bool
}

type mockPluginClient struct {
	// atomic.Value does not allow us to atomically publish, close and replace the chan
	buildEventsLock *sync.Mutex

	latestBuildContext atomic.Value[testBuildContext]
	buildResult        *connect.Response[langpb.BuildResponse]

	cmdError chan error
}

var _ pluginClient = &mockPluginClient{}

func newMockPluginClient() *mockPluginClient {
	return &mockPluginClient{
		buildEventsLock: &sync.Mutex{},
		cmdError:        make(chan error),
	}
}

func (p *mockPluginClient) getDependencies(context.Context, *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	panic("not implemented")
}

func buildContextFromProto(proto *langpb.BuildContext) (BuildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return BuildContext{}, errors.Wrap(err, "could not load schema from build context proto")
	}
	return BuildContext{
		Schema:       sch,
		Dependencies: proto.Dependencies,
		Config: moduleconfig.ModuleConfig{
			Dir:            proto.ModuleConfig.Dir,
			Language:       "test",
			Realm:          "test",
			Module:         proto.ModuleConfig.Name,
			Build:          optional.Ptr(proto.ModuleConfig.Build).Default(""),
			DevModeBuild:   optional.Ptr(proto.ModuleConfig.DevModeBuild).Default(""),
			DeployDir:      proto.ModuleConfig.DeployDir,
			Watch:          proto.ModuleConfig.Watch,
			LanguageConfig: proto.ModuleConfig.LanguageConfig.AsMap(),
			SQLRootDir:     proto.ModuleConfig.SqlRootDir,
		},
	}, nil
}

func (p *mockPluginClient) generateStubs(context.Context, *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	panic("not implemented")
}

func (p *mockPluginClient) syncStubReferences(context.Context, *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	panic("not implemented")
}

func (p *mockPluginClient) build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (*connect.Response[langpb.BuildResponse], error) {
	p.buildEventsLock.Lock()
	defer p.buildEventsLock.Unlock()

	bctx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	p.latestBuildContext.Store(testBuildContext{
		BuildContext: bctx,
		IsRebuild:    false,
	})
	if p.buildResult == nil {
		<-time.After(time.Second)
	}
	return p.buildResult, nil
}

func (p *mockPluginClient) buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	bctx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	p.latestBuildContext.Store(testBuildContext{
		BuildContext: bctx,
		IsRebuild:    true,
	})
	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

func (p *mockPluginClient) kill() error {
	return nil
}

func (p *mockPluginClient) cmdErr() <-chan error {
	return p.cmdError
}

func setUp() (context.Context, *LanguagePlugin, *mockPluginClient, BuildContext) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	mockImpl := newMockPluginClient()
	plugin := newPluginForTesting(ctx, mockImpl)

	bctx := BuildContext{
		Config: moduleconfig.ModuleConfig{
			Module:   "name",
			Dir:      "test/dir",
			Language: "test-lang",
		},
		Schema:       &schema.Schema{Realms: []*schema.Realm{{Name: "test"}}},
		Dependencies: []string{},
	}
	return ctx, plugin, mockImpl, bctx
}

func TestNewModuleFlags(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		protoFlags    []*langpb.GetNewModuleFlagsResponse_Flag
		expectedFlags []*kong.Flag
		expectedError optional.Option[string]
	}{
		{
			protoFlags: []*langpb.GetNewModuleFlagsResponse_Flag{
				{
					Name:        "full-flag",
					Help:        "This has all the fields set",
					Envar:       optional.Some("full-flag").Ptr(),
					Short:       optional.Some("f").Ptr(),
					Placeholder: optional.Some("placeholder").Ptr(),
					Default:     optional.Some("defaultValue").Ptr(),
				},
				{
					Name: "sparse-flag",
					Help: "This has only the minimum fields set",
				},
			},
			expectedFlags: []*kong.Flag{
				{
					Value: &kong.Value{
						Name:       "full-flag",
						Help:       "This has all the fields set",
						HasDefault: true,
						Default:    "defaultValue",
						Tag: &kong.Tag{
							Envs: []string{
								"full-flag",
							},
						},
					},
					PlaceHolder: "placeholder",
					Short:       'f',
				},
				{
					Value: &kong.Value{
						Name: "sparse-flag",
						Help: "This has only the minimum fields set",
						Tag:  &kong.Tag{},
					},
				},
			},
		},
		{
			protoFlags: []*langpb.GetNewModuleFlagsResponse_Flag{
				{
					Name:  "multi-char-short",
					Help:  "This has all the fields set",
					Short: optional.Some("multi").Ptr(),
				},
			},
			expectedError: optional.Some(`invalid flag declared: short flag "multi" for multi-char-short must be a single character`),
		},
		{
			protoFlags: []*langpb.GetNewModuleFlagsResponse_Flag{
				{
					Name:  "dupe-short-1",
					Help:  "Short must be unique",
					Short: optional.Some("d").Ptr(),
				},
				{
					Name:  "dupe-short-2",
					Help:  "Short must be unique",
					Short: optional.Some("d").Ptr(),
				},
			},
			expectedError: optional.Some(`multiple flags declared with the same short name: dupe-short-1 and dupe-short-2`),
		},
	} {
		t.Run(tt.protoFlags[0].Name, func(t *testing.T) {
			t.Parallel()
			kongFlags, err := kongFlagsFromProto(tt.protoFlags)
			if expectedError, ok := tt.expectedError.Get(); ok {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFlags, kongFlags)
		})
	}
}
