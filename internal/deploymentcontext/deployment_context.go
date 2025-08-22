package deploymentcontext

import (
	"context"
	"encoding/json"
	"iter"
	"maps"
	"os"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/rpc"
)

type DeploymentContextProvider func(ctx context.Context) <-chan DeploymentContext

type SecretsProvider func(ctx context.Context) (map[string][]byte, error)
type ConfigProvider func(ctx context.Context) (map[string][]byte, error)

type RouteProvider interface {
	Subscribe(ctx context.Context) <-chan string
	Route(module string) string
}

// Verb is a function that takes a request and returns a response but is not constrained by request/response type like ftl.Verb
//
// It is used for definitions of mock verbs as well as real implementations of verbs to directly execute
type Verb func(ctx context.Context, req any) (resp any, err error)

// DeploymentContext holds the context needed for a module, including configs, secrets and DSNs
//
// DeploymentContext is immutable
type DeploymentContext struct {
	module    string
	configs   map[string][]byte
	secrets   map[string][]byte
	routes    map[string]string
	databases map[string]Database
	egress    map[string]string

	isTesting                     bool
	mockVerbs                     map[schema.RefKey]Verb
	allowDirectVerbBehaviorGlobal bool
	allowDirectVerb               schema.RefKey
	leaseClient                   optional.Option[LeaseClient]
	allowDirectQueryVerbs         bool
}

// DynamicDeploymentContext provides up-to-date DeploymentContext instances supplied by the controller
type DynamicDeploymentContext struct {
	current atomic.Value[DeploymentContext]
}

// Builder is used to build a DeploymentContext
type Builder DeploymentContext

type contextKeyDynamicDeploymentContext struct{}

func Empty(module string) DeploymentContext {
	return NewBuilder(module).Build()
}

// NewBuilder creates a new blank Builder for the given module.
func NewBuilder(module string) *Builder {
	return &Builder{
		module:    module,
		configs:   map[string][]byte{},
		secrets:   map[string][]byte{},
		databases: map[string]Database{},
		mockVerbs: map[schema.RefKey]Verb{},
		routes:    map[string]string{},
		egress:    map[string]string{},
	}
}

func NewBuilderFromContext(ctx DeploymentContext) *Builder {
	return &Builder{
		module:                        ctx.module,
		configs:                       ctx.configs,
		secrets:                       ctx.secrets,
		databases:                     ctx.databases,
		isTesting:                     ctx.isTesting,
		mockVerbs:                     ctx.mockVerbs,
		allowDirectVerbBehaviorGlobal: ctx.allowDirectVerbBehaviorGlobal,
		allowDirectVerb:               ctx.allowDirectVerb,
		leaseClient:                   ctx.leaseClient,
		allowDirectQueryVerbs:         ctx.allowDirectQueryVerbs,
		egress:                        ctx.egress,
	}
}

// AddConfigs adds configuration values (as bytes) to the builder
func (b *Builder) AddConfigs(configs map[string][]byte) *Builder {
	for name, data := range configs {
		b.configs[name] = data
	}
	return b
}

// AddSecrets adds secrets values (as bytes) to the builder
func (b *Builder) AddSecrets(secrets map[string][]byte) *Builder {
	for name, data := range secrets {
		b.secrets[name] = data
	}
	return b
}

// AddEgress adds egress values to the builder
func (b *Builder) AddEgress(egress map[string]string) *Builder {
	for name, data := range egress {
		b.egress[name] = data
	}
	return b
}

func (b *Builder) AddRoutes(routes map[string]string) *Builder {
	for name, data := range routes {
		b.routes[name] = data
	}
	return b
}

// AddDatabases adds databases to the builder
func (b *Builder) AddDatabases(databases map[string]Database) *Builder {
	for name, db := range databases {
		b.databases[name] = db
	}
	return b
}

// AddAllowedDirectVerb adds a verb that can be called directly within the current context
func (b *Builder) AddAllowedDirectVerb(ref reflection.Ref) *Builder {
	b.allowDirectVerb = schema.RefKey(ref)
	return b
}

// UpdateForTesting marks the builder as part of a test environment and adds mock verbs and flags for other test features.
func (b *Builder) UpdateForTesting(mockVerbs map[schema.RefKey]Verb, allowDirectVerbBehavior bool, allowDirectQueryVerbs bool, leaseClient LeaseClient) *Builder {
	b.isTesting = true
	for name, verb := range mockVerbs {
		b.mockVerbs[name] = verb
	}
	b.allowDirectVerbBehaviorGlobal = allowDirectVerbBehavior
	b.allowDirectQueryVerbs = allowDirectQueryVerbs
	b.leaseClient = optional.Some[LeaseClient](leaseClient)
	return b
}

func (b *Builder) Build() DeploymentContext {
	return DeploymentContext(reflect.DeepCopy(*b))
}

// GetConfig reads a configuration value for the module.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m DeploymentContext) GetConfig(name string, value any) error {
	data, ok := m.configs[name]
	if !ok {
		return errors.Errorf("no config value for %q", name)
	}
	return errors.WithStack(json.Unmarshal(data, value))
}

// GetSecret reads a secret value for the module.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m DeploymentContext) GetSecret(name string, value any) error {
	data, ok := m.secrets[name]
	if !ok {
		return errors.Errorf("no secret value for %q", name)
	}
	return errors.WithStack(json.Unmarshal(data, value))
}

func (m DeploymentContext) GetEgress(name string) string {
	return m.egress[name]
}

func (m DeploymentContext) GetRoutes() iter.Seq[string] {
	return maps.Keys(m.routes)
}

func (m DeploymentContext) GetRoute(name string) string {
	return m.routes[name]
}

func (m DeploymentContext) GetModule() string {
	return m.module
}

// GetDatabase gets a database DSN by name and type.
//
// Returns an error if no database with that name is found or it is not the
// expected type. When in a testing context (via ftltest), an error is returned
// if the database is not a test database.
func (m DeploymentContext) GetDatabase(name string, dbType DBType) (string, bool, error) {
	db, ok := m.databases[name]
	// TODO: Remove databases from the context once we have a way to inject test dbs in some other way
	if !ok {
		if dbType == DBTypePostgres {
			proxyAddress := os.Getenv("FTL_PROXY_POSTGRES_ADDRESS")
			return "postgres://" + proxyAddress + "/" + name, false, nil
		} else if dbType == DBTypeMySQL {
			proxyAddress := os.Getenv("FTL_PROXY_MYSQL_ADDRESS_" + strings.ToUpper(name))
			return "ftl:ftl@tcp(" + proxyAddress + ")/" + name, false, nil
		}
		return "", false, errors.Errorf("missing DSN for database %s", name)
	}
	if db.DBType != dbType {
		return "", false, errors.Errorf("database %s does not match expected type of %s", name, dbType)
	}
	if m.isTesting && !db.isTestDB {
		return "", false, errors.Errorf("accessing non-test database %q while testing: try adding ftltest.WithDatabase[MyConfig]() as an option with ftltest.Context(...)", name)
	}
	return db.DSN, db.isTestDB, nil
}

// LeaseClient is the interface for acquiring, heartbeating and releasing leases
type LeaseClient interface {
	// Returns ResourceExhausted if the lease is held.
	Acquire(ctx context.Context, module string, key []string, ttl time.Duration) error
	Heartbeat(ctx context.Context, module string, key []string, ttl time.Duration) error
	Release(ctx context.Context, key []string) error
}

// MockLeaseClient provides a mock lease client when testing
func (m DeploymentContext) MockLeaseClient() optional.Option[LeaseClient] {
	return m.leaseClient
}

// BehaviorForVerb returns what to do to execute a verb
//
// This allows module context to dictate behavior based on testing options
// Returning optional.Nil indicates the verb should be executed normally via the controller
func (m DeploymentContext) BehaviorForVerb(ref schema.Ref) (optional.Option[VerbBehavior], error) {
	if mock, ok := m.mockVerbs[ref.ToRefKey()]; ok {
		return optional.Some(VerbBehavior(MockBehavior{Mock: mock})), nil
	} else if m.isTesting && reflection.IsQueryVerb(reflection.Ref{Module: ref.Module, Name: ref.Name}) {
		if m.allowDirectQueryVerbs {
			return optional.Some(VerbBehavior(DirectBehavior{})), nil
		}
		if ref.Module == m.module {
			return optional.None[VerbBehavior](), errors.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s, ...) or enable all SQL verbs to execute directly with ftltest.WithSQLVerbsEnabled()", strings.ToUpper(ref.Name[:1])+ref.Name[1:])
		}
		return optional.None[VerbBehavior](), errors.Errorf("no mock found: query verbs must be mocked with ftltest.WhenVerb(%s.%s, ...)", ref.Module, strings.ToUpper(ref.Name[:1])+ref.Name[1:])
	} else if (m.allowDirectVerbBehaviorGlobal || m.allowDirectVerb == ref.ToRefKey()) && ref.Module == m.module {
		return optional.Some(VerbBehavior(DirectBehavior{})), nil
	} else if m.isTesting {
		if ref.Module == m.module {
			return optional.None[VerbBehavior](), errors.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s, ...) or enable all calls within the module with ftltest.WithCallsAllowedWithinModule()", strings.ToUpper(ref.Name[:1])+ref.Name[1:])
		}
		return optional.None[VerbBehavior](), errors.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s.%s, ...)", ref.Module, strings.ToUpper(ref.Name[:1])+ref.Name[1:])
	}
	return optional.None[VerbBehavior](), nil
}

type DeploymentContextSupplier interface {
	Subscribe(ctx context.Context, moduleName string, sink func(ctx context.Context, moduleContext DeploymentContext), errorRetryCallback func(err error) bool)
}

type grpcDeploymentContextSupplier struct {
	client ftlv1connect.DeploymentContextServiceClient
}

func NewDeploymentContextSupplier(client ftlv1connect.DeploymentContextServiceClient) DeploymentContextSupplier {
	return DeploymentContextSupplier(grpcDeploymentContextSupplier{client})
}

func (g grpcDeploymentContextSupplier) Subscribe(ctx context.Context, deploymentName string, sink func(ctx context.Context, moduleContext DeploymentContext), errorRetryCallback func(err error) bool) {
	request := &ftlv1.GetDeploymentContextRequest{Deployment: deploymentName}
	callback := func(_ context.Context, resp *ftlv1.GetDeploymentContextResponse) error {
		mc, err := FromProto(resp)
		if err != nil {
			return errors.WithStack(err)
		}
		sink(ctx, mc)
		return nil
	}
	go rpc.RetryStreamingServerStream(ctx, "module-context", backoff.Backoff{}, request, g.client.GetDeploymentContext, callback, errorRetryCallback)
}

// NewDynamicContext creates a new DynamicDeploymentContext. This operation blocks
// until the first DeploymentContext is supplied by the controller.
//
// The DynamicDeploymentContext will continually update as updated DeploymentContext's
// are streamed from the controller. This operation may time out if the first
// module context is not supplied quickly enough (fixed at 5 seconds).
func NewDynamicContext(ctx context.Context, supplier DeploymentContextSupplier, deploymentName string) (*DynamicDeploymentContext, error) {
	result := &DynamicDeploymentContext{}

	await := sync.WaitGroup{}
	await.Add(1)
	releaseOnce := sync.Once{}

	ctx, cancel := context.WithCancelCause(ctx)
	deadline, timeoutCancel := context.WithTimeout(ctx, 5*time.Second)
	g, _ := errgroup.WithContext(deadline)
	defer timeoutCancel()

	// asynchronously consumes a subscription of DeploymentContext changes and signals the arrival of the first
	supplier.Subscribe(
		ctx,
		deploymentName,
		func(ctx context.Context, moduleContext DeploymentContext) {
			result.current.Store(moduleContext)
			releaseOnce.Do(func() {
				await.Done()
			})
		},
		func(err error) bool {
			var connectErr *connect.Error

			if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeInternal {
				cancel(errors.Join(err, context.Canceled))
				releaseOnce.Do(func() {
					await.Done()
				})
				return false
			}

			return true
		})

	// await the WaitGroup's completion which either signals the availability of the
	// first DeploymentContext or an error
	g.Go(func() error {
		await.Wait()
		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		default:
			return nil
		}
	})

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "error waiting for first DeploymentContext")
	}

	return result, nil
}

// CurrentContext immediately returns the most recently updated DeploymentContext
func (m *DynamicDeploymentContext) CurrentContext() DeploymentContext {
	return m.current.Load()
}

// FromContext returns the DynamicDeploymentContext attached to a context.
func FromContext(ctx context.Context) *DynamicDeploymentContext {
	m, ok := ctx.Value(contextKeyDynamicDeploymentContext{}).(*DynamicDeploymentContext)
	if !ok {
		panic("no DeploymentContext in context")
	}
	return m
}

// ApplyToContext returns a Go context.Context with DynamicDeploymentContext added.
func (m *DynamicDeploymentContext) ApplyToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyDynamicDeploymentContext{}, m)
}

// VerbBehavior indicates how to execute a verb
type VerbBehavior interface {
	Call(ctx context.Context, verb Verb, request any) (any, error)
}

// DirectBehavior indicates that the verb should be executed by calling the function directly (for testing)
type DirectBehavior struct{}

func (DirectBehavior) Call(ctx context.Context, verb Verb, req any) (any, error) {
	return errors.WithStack2(verb(ctx, req))
}

var _ VerbBehavior = DirectBehavior{}

// MockBehavior indicates the verb has a mock implementation
type MockBehavior struct {
	Mock Verb
}

func (b MockBehavior) Call(ctx context.Context, _ Verb, req any) (any, error) {
	return errors.WithStack2(b.Mock(ctx, req))
}
