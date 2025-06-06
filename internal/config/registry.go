package config

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
)

// NewConfigurationRegistry creates a new configuration registry.
//
// If adminClient is provided, a remote provider will be registered
func NewConfigurationRegistry(adminClient Option[adminpbconnect.AdminServiceClient]) *Registry[Configuration] {
	registry := NewRegistry[Configuration]()
	registry.Register(NewFileProviderFactory[Configuration]())
	registry.Register(NewMemoryProviderFactory[Configuration]())
	if adminClient, ok := adminClient.Get(); ok {
		registry.Register(NewRemoteProviderFactory[Configuration](adminClient))
	}
	return registry
}

func NewSecretsRegistry(adminClient Option[adminpbconnect.AdminServiceClient]) *Registry[Secrets] {
	registry := NewRegistry[Secrets]()
	registry.Register(NewFileProviderFactory[Secrets]())
	registry.Register(NewMemoryProviderFactory[Secrets]())
	registry.Register(NewOnePasswordProviderFactory())
	if adminClient, ok := adminClient.Get(); ok {
		registry.Register(NewRemoteProviderFactory[Secrets](adminClient))
	}
	return registry
}

// Factory is a function that creates a Provider.
//
// "projectRoot" is the root directory of the project. "key" is the provider key, including any configuration options.
type Factory[R Role] func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[R], error)

// Registry that lazily constructs configuration providers.
type Registry[R Role] struct {
	factories map[ProviderKind]Factory[R]
}

func NewRegistry[R Role]() *Registry[R] {
	return &Registry[R]{
		factories: map[ProviderKind]Factory[R]{},
	}
}

// Providers returns the list of registered provider keys.
func (r *Registry[R]) Providers() []ProviderKind {
	return slices.Collect(maps.Keys(r.factories))
}

func (r *Registry[R]) Register(name ProviderKind, factory Factory[R]) {
	r.factories[name] = factory
}

func (r *Registry[R]) Get(ctx context.Context, projectRoot string, key ProviderKey) (Provider[R], error) {
	factory, ok := r.factories[key.Kind()]
	if !ok {
		var role R
		return nil, errors.Errorf("%s: %s provider not found", key, role)
	}
	provider, err := factory(ctx, projectRoot, key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct %s provider", key)
	}

	// If the provider is asynchronous, wrap it in a cache decorator.
	var syncProvider Provider[R]
	switch provider := provider.(type) {
	case AsynchronousProvider[R]:
		syncProvider = NewCacheDecorator(ctx, provider)
	case Provider[R]:
		syncProvider = provider
	default:
		// Not ideal that this fails at runtime, but I haven't figured out a better way to handle this yet.
		panic(fmt.Sprintf("provider %s must be either a SynchronousProvider or an AsynchronousProvider", key))
	}

	// Wrap the provider in an environment variable decorator.
	syncProvider = NewEnvarDecorator(syncProvider)
	// Wrap in global fallback decorator.
	syncProvider = NewGlobalFallbackDecorator(syncProvider)
	return syncProvider, nil
}
