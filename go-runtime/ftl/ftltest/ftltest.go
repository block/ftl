// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"

	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/runner/query"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/internal"
	"github.com/block/ftl/go-runtime/server"
	queryclient "github.com/block/ftl/go-runtime/server/query"
	"github.com/block/ftl/go-runtime/server/rpccontext"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/deploymentcontext"
	pc "github.com/block/ftl/internal/projectconfig"
	mcu "github.com/block/ftl/internal/testutils/modulecontext"
)

// Allows tests to mock module reflection
var moduleGetter = reflection.Module

type OptionsState struct {
	databases               map[string]deploymentcontext.Database
	mockVerbs               map[schema.RefKey]deploymentcontext.Verb
	allowDirectVerbBehavior bool
	allowDirectSQLVerbs     bool
}

type optionRank int

const (
	profile optionRank = iota
	other
)

type Option struct {
	rank  optionRank
	apply func(context.Context, *OptionsState) error
}

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...Option) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	module := moduleGetter()
	return newContext(ctx, module, options...)
}

func newContext(ctx context.Context, module string, options ...Option) context.Context {
	state := &OptionsState{
		databases: make(map[string]deploymentcontext.Database),
		mockVerbs: make(map[schema.RefKey]deploymentcontext.Verb),
	}

	ctx = contextWithFakeFTL(ctx, options...)

	sort.Slice(options, func(i, j int) bool {
		return options[i].rank < options[j].rank
	})

	for _, option := range options {
		err := option.apply(ctx, state)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}

	if state.allowDirectSQLVerbs {
		querySvc, err := query.New(ctx, &schema.Module{Name: module}, xsync.NewMapOf[string, string]())
		if err != nil {
			panic(errors.Wrap(err, "failed to create in-process query service to execute query verbs"))
		}
		for name, db := range state.databases {
			err := querySvc.AddQueryConn(ctx, name, db)
			if err != nil {
				panic(errors.Wrapf(err, "failed to create DB connection for %s", name))
			}
		}
		ctx = rpccontext.ContextWithClient[queryconnect.QueryServiceClient, ftlv1.PingRequest, ftlv1.PingResponse, *ftlv1.PingResponse](ctx, queryclient.NewInlineQueryClient(querySvc)) // yuck
	}

	builder := deploymentcontext.NewBuilder(module).AddDatabases(state.databases)
	builder = builder.UpdateForTesting(state.mockVerbs, state.allowDirectVerbBehavior, state.allowDirectSQLVerbs, newFakeLeaseClient())

	return mcu.MakeDynamic(ctx, builder.Build()).ApplyToContext(ctx)
}

// SubContext applies the given options to the given context, creating a new
// context extending the previous one.
//
// Does not modify the existing context
func SubContext(ctx context.Context, options ...Option) context.Context {
	oldFtl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
	module := moduleGetter()
	return newContext(ctx, module, append(oldFtl.options, options...)...)
}

// WithDefaultProjectFile loads config and secrets from the default project
// file, which is either the FTL_CONFIG environment variable or the
// ftl-project.toml file in the git root.
func WithDefaultProjectFile() Option {
	return WithProjectFile("")
}

// WithProjectFile loads config and secrets from a project file
//
// Takes a path to an FTL project file. If an empty path is provided, the path
// is inferred from the FTL_CONFIG environment variable. If that is not found,
// the ftl-project.toml file in the git root is used. If a project file is not
// found, an error is returned.
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithProjectFile("path/to/ftl-project.yaml"),
//		// ... other options
//	)
func WithProjectFile(path string) Option {
	// Convert to absolute path immediately in case working directory changes
	var preprocessingErr error
	if path == "" {
		var ok bool
		path, ok = pc.DefaultConfigPath().Get()
		if !ok {
			preprocessingErr = errors.Errorf("could not find default project file in $FTL_CONFIG or git")
		}
	}
	return Option{
		rank: profile,
		apply: func(ctx context.Context, state *OptionsState) error {
			if preprocessingErr != nil {
				return errors.WithStack(preprocessingErr)
			}
			if _, err := os.Stat(path); err != nil {
				return errors.Wrap(err, "error accessing project file")
			}
			projectConfig, err := pc.Load(ctx, optional.Some(path))
			if err != nil {
				return errors.Wrap(err, "project")
			}

			cr := config.NewConfigurationRegistry(nil)
			cm, err := cr.Get(ctx, projectConfig.Root(), projectConfig.ConfigProvider)
			if err != nil {
				return errors.Wrap(err, "could not set up configs")
			}
			configs, err := config.MapForModule(ctx, cm, moduleGetter())
			if err != nil {
				return errors.Wrap(err, "could not read configs")
			}

			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			for name, data := range configs {
				if err := fftl.setConfig(name, json.RawMessage(data)); err != nil {
					return errors.WithStack(err)
				}
			}

			sr := config.NewSecretsRegistry(nil)
			sm, err := sr.Get(ctx, projectConfig.Root(), projectConfig.SecretsProvider)
			if err != nil {
				return errors.Wrap(err, "could not set up secrets")
			}
			secrets, err := config.MapForModule(ctx, sm, moduleGetter())
			if err != nil {
				return errors.Wrap(err, "could not read secrets")
			}
			for name, data := range secrets {
				if err := fftl.setSecret(name, json.RawMessage(data)); err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		},
	}

}

// WithConfig sets a configuration for the current module
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithConfig(exampleEndpoint, "https://example.com"),
//		// ... other options
//	)
func WithConfig[T ftl.ConfigType](config ftl.Config[T], value T) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			if config.Module != moduleGetter() {
				return errors.Errorf("config %v does not match current module %s", config.Module, moduleGetter())
			}
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			if err := fftl.setConfig(config.Name, value); err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}
}

// WithSecret sets a secret for the current module
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithSecret(privateKey, "abc123"),
//		// ... other options
//	)
func WithSecret[T ftl.SecretType](secret ftl.Secret[T], value T) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			if secret.Module != moduleGetter() {
				return errors.Errorf("secret %v does not match current module %s", secret.Module, moduleGetter())
			}
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			if err := fftl.setSecret(secret.Name, value); err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}
}

// WhenVerb replaces an implementation for a verb
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenVerb[example.VerbClient](func(ctx context.Context, req example.Req) (example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenVerb[VerbClient, Req, Resp any](fake ftl.Verb[Req, Resp]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[VerbClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				request, ok := req.(Req)
				if !ok {
					return nil, errors.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
				}
				return errors.WithStack2(fake(ctx, request))
			}
			return nil
		},
	}
}

// WhenSource replaces an implementation for a verb with no request
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenSource[example.SourceClient](func(ctx context.Context) (example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenSource[SourceClient, Resp any](fake ftl.Source[Resp]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[SourceClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				return errors.WithStack2(fake(ctx))
			}
			return nil
		},
	}
}

// WhenSink replaces an implementation for a verb with no response
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenSink[example.SinkClient](func(ctx context.Context, req example.Req) error {
//	    	...
//		}),
//		// ... other options
//	)
func WhenSink[SinkClient, Req any](fake ftl.Sink[Req]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[SinkClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				request, ok := req.(Req)
				if !ok {
					return nil, errors.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
				}
				return ftl.Unit{}, errors.WithStack(fake(ctx, request))
			}
			return nil
		},
	}
}

// WhenEmpty replaces an implementation for a verb with no request or response
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenEmpty[example.EmptyClient](func(ctx context.Context) error {
//	    	...
//		}),
//	)
func WhenEmpty[EmptyClient any](fake ftl.Empty) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[EmptyClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				return ftl.Unit{}, errors.WithStack(fake(ctx))
			}
			return nil
		},
	}
}

// WithCallsAllowedWithinModule allows tests to enable calls to all verbs within the current module
//
// Any overrides provided by calling WhenVerb(...) will take precedence
func WithCallsAllowedWithinModule() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			state.allowDirectVerbBehavior = true
			return nil
		},
	}
}

// WhenMap injects a fake implementation of a Mapping function
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenMap(Example.MapHandle, func(ctx context.Context) (U, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenMap[T, U any](mapper *ftl.MapHandle[T, U], fake func(context.Context) (U, error)) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			addMapMock(fftl, mapper, fake)
			return nil
		},
	}
}

// WithMapsAllowed allows all `ftl.Map` calls to pass through to their original
// implementation.
//
// Any overrides provided by calling WhenMap(...) will take precedence.
func WithMapsAllowed() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			fftl.startAllowingMapCalls()
			return nil
		},
	}
}

// dsnSecretKey returns the key for the secret that is expected to hold the DSN for a database.
//
// The format is FTL_DSN_<MODULE>_<DBNAME>
func dsnSecretKey(module, name string) string {
	return fmt.Sprintf("FTL_DSN_%s_%s", strings.ToUpper(module), strings.ToUpper(name))
}

// getDSNFromSecret returns the DSN for a database from the relevant secret
func getDSNFromSecret(ftl internal.FTL, module, name string) (string, error) {
	key := dsnSecretKey(module, name)
	var dsn string
	if err := ftl.GetSecret(context.Background(), key, &dsn); err != nil {
		return "", errors.Wrapf(err, "could not get DSN for database %q from secret %q", name, key)
	}
	return dsn, nil
}

// Call a Verb inline, applying resources and test behavior.
func Call[VerbClient, Req, Resp any](ctx context.Context, req Req) (Resp, error) {
	return errors.WithStack2(call[VerbClient, Req, Resp](ctx, req))
}

// CallSource calls a Source inline, applying resources and test behavior.
func CallSource[VerbClient, Resp any](ctx context.Context) (Resp, error) {
	return errors.WithStack2(call[VerbClient, ftl.Unit, Resp](ctx, ftl.Unit{}))
}

// CallSink calls a Sink inline, applying resources and test behavior.
func CallSink[VerbClient, Req any](ctx context.Context, req Req) error {
	_, err := call[VerbClient, Req, ftl.Unit](ctx, req)
	return errors.WithStack(err)
}

// CallEmpty calls an Empty inline, applying resources and test behavior.
func CallEmpty[VerbClient any](ctx context.Context) error {
	_, err := call[VerbClient, ftl.Unit, ftl.Unit](ctx, ftl.Unit{})
	return errors.WithStack(err)
}

// GetDatabaseHandle returns a database handle using the given database config.
func GetDatabaseHandle[T any]() (ftl.DatabaseHandle[T], error) {
	reflectedDB := reflection.GetDatabase[T]()
	if reflectedDB == nil {
		return ftl.DatabaseHandle[T]{}, errors.Errorf("could not find database for config")
	}

	var dbType ftl.DatabaseType
	switch reflectedDB.DBType {
	case "postgres":
		dbType = ftl.DatabaseTypePostgres
	case "mysql":
		dbType = ftl.DatabaseTypeMysql
	default:
		return ftl.DatabaseHandle[T]{}, errors.Errorf("unsupported database type %v", reflectedDB.DBType)
	}
	return ftl.NewDatabaseHandle[T](reflectedDB.Name, dbType, reflectedDB.DB), nil
}

func call[VerbClient, Req, Resp any](ctx context.Context, req Req) (resp Resp, err error) {
	ref := reflection.ClientRef[VerbClient]()
	// always allow direct behavior for the verb triggered by this call
	moduleCtx := deploymentcontext.NewBuilderFromContext(
		deploymentcontext.FromContext(ctx).CurrentContext(),
	).AddAllowedDirectVerb(ref).Build()
	ctx = mcu.MakeDynamic(ctx, moduleCtx).ApplyToContext(ctx)

	inline := server.InvokeVerb[Req, Resp](ref)
	override, err := moduleCtx.BehaviorForVerb(schema.Ref{Module: ref.Module, Name: ref.Name})
	if err != nil {
		return resp, errors.Wrapf(err, "test harness failed to retrieve behavior for verb %s", ref)
	}
	if behavior, ok := override.Get(); ok {
		uncheckedResp, err := behavior.Call(ctx, deploymentcontext.Verb(widenVerb(inline)), req)
		if err != nil {
			return resp, errors.Wrapf(err, "test harness failed to call verb %s", ref)
		}
		if r, ok := uncheckedResp.(Resp); ok {
			return r, nil
		}
		return resp, errors.Errorf("%s: overridden verb had invalid response type %T, expected %v", ref, uncheckedResp, reflect.TypeFor[Resp]())
	}
	return errors.WithStack2(inline(ctx, req))
}

func widenVerb[Req, Resp any](verb ftl.Verb[Req, Resp]) ftl.Verb[any, any] {
	return func(ctx context.Context, uncheckedReq any) (any, error) {
		req, ok := uncheckedReq.(Req)
		if !ok {
			return nil, errors.Errorf("invalid request type %T for %v, expected %v", uncheckedReq, reflection.FuncRef(verb), reflect.TypeFor[Req]())
		}
		return errors.WithStack2(verb(ctx, req))
	}
}
