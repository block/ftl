// Package config is the FTL configuration and secret management API.
package config

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/encoding"
)

// ErrNotFound is returned when a configuration entry is not found or cannot be resolved.
var ErrNotFound = errors.New("not found")

// Entry in the configuration store.
type Entry struct {
	Ref
	Accessor *url.URL
}

// A Ref is a reference to a configuration value.
type Ref struct {
	// If [Module] is omitted the Ref is considered to be a global value.
	Module optional.Option[string]
	Name   string
}

// NewRef creates a new Ref.
//
// If [module] is empty, the Ref is considered to be a global configuration value.
func NewRef(module optional.Option[string], name string) Ref {
	return Ref{Module: module, Name: name}
}

// ParseRef parses a string into a Ref.
func ParseRef(s string) Ref {
	ref := Ref{}
	_ = ref.UnmarshalText([]byte(s)) //nolint
	return ref
}

func (k Ref) String() string {
	if m, ok := k.Module.Get(); ok {
		return m + "." + k.Name
	}
	return k.Name
}

func (k *Ref) UnmarshalText(text []byte) error {
	s := string(text)
	if i := strings.Index(s, "."); i != -1 {
		k.Module = optional.Some(s[:i])
		k.Name = s[i+1:]
	} else {
		k.Name = s
	}
	return nil
}

// Role of a Provider.
type Role interface {
	Secrets | Configuration
	String() string
}

type Secrets struct{}

func (Secrets) String() string { return "secrets" }

type Configuration struct{}

func (Configuration) String() string { return "configuration" }

// Value represents a configuration value with its reference.
type Value struct {
	Ref
	Value optional.Option[[]byte]
}

// BaseProvider is the base generic interface for storing and retrieving configuration and secrets.
//
// Note that implementations of this interface should be thread-safe, and also must implement either SynchronousProvider
// or AsynchronousProvider.
type BaseProvider[R Role] interface {
	// Role returns the role of the provider (either Configuration or Secrets)
	Role() R

	// Key returns the key of the provider.
	Key() ProviderKey

	// Store a configuration value and return its key.
	Store(ctx context.Context, ref Ref, value []byte) error
	// Delete a configuration value.
	Delete(ctx context.Context, ref Ref) error
	// Close the provider.
	Close(ctx context.Context) error
}

// Provider is an interface for storing and retrieving configuration and secrets.
type Provider[R Role] interface {
	BaseProvider[R]

	// Load a configuration value.
	Load(ctx context.Context, ref Ref) ([]byte, error)
	// List all configuration keys.
	List(ctx context.Context, withValues bool) ([]Value, error)
}

// AsynchronousProvider is an interface for Provider's that support syncing values.
// This is recommended if the Provider allows batch access, or is expensive to load.
type AsynchronousProvider[R Role] interface {
	BaseProvider[R]

	// SyncInterval returns the desired time between syncs.
	SyncInterval() time.Duration

	// Sync is called periodically to update the cache with the latest values.
	//
	// SyncInterval() provides the desired time between syncs.
	//
	// If Sync() returns an error, sync will be retried with an exponential backoff.
	Sync(ctx context.Context) (map[Ref]SyncedValue, error)
}

type VersionToken any

type SyncedValue struct {
	Value []byte

	// VersionToken is a way of storing a version provided by the source of truth (eg: lastModified)
	// it is nil when:
	// - the owner of the cache is not using version tokens
	// - the cache is updated after writing
	VersionToken optional.Option[VersionToken]
}

// Store a typed configuration value.
func Store[T any, R Role](ctx context.Context, provider Provider[R], ref Ref, value T) error {
	data, err := encoding.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	err = provider.Store(ctx, ref, data)
	if err != nil {
		return fmt.Errorf("failed to store value: %w", err)
	}
	return nil
}

// Load a typed configuration value.
func Load[T any, R Role](ctx context.Context, provider Provider[R], ref Ref) (out T, err error) {
	data, err := provider.Load(ctx, ref)
	if err != nil {
		return out, fmt.Errorf("failed to load value: %w", err)
	}
	err = encoding.Unmarshal(data, &out)
	if err != nil {
		return out, fmt.Errorf("failed to marshal value: %w", err)
	}
	return out, nil
}
