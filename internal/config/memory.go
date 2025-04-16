package config

import (
	"context"
	"slices"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
)

const MemoryProviderKind ProviderKind = "memory"

type MemoryProvider[R Role] struct {
	config *xsync.MapOf[Ref, []byte]
}

var _ Provider[Configuration] = &MemoryProvider[Configuration]{}

func NewMemoryProviderFactory[R Role]() (ProviderKind, Factory[R]) {
	return MemoryProviderKind, func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[R], error) {
		return NewMemoryProvider[R](), nil
	}
}

func (m *MemoryProvider[R]) Key() ProviderKey {
	return NewProviderKey(MemoryProviderKind)
}
func (m *MemoryProvider[R]) Role() R                         { return R{} }
func (m *MemoryProvider[R]) Close(ctx context.Context) error { return nil }

func (m *MemoryProvider[R]) Delete(ctx context.Context, ref Ref) error {
	m.config.Delete(ref)
	return nil
}

func (m *MemoryProvider[R]) Store(ctx context.Context, ref Ref, value []byte) error {
	m.config.Store(ref, value)
	return nil
}

func (m *MemoryProvider[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	data, ok := m.config.Load(ref)
	if !ok {
		return nil, errors.Wrapf(ErrNotFound, "could not load %s", ref)
	}
	return data, nil
}

func (m *MemoryProvider[R]) List(ctx context.Context, withValues bool) ([]Value, error) {
	refs := make([]Value, 0, m.config.Size())
	m.config.Range(func(ref Ref, data []byte) bool {
		value := Value{Ref: ref}
		if withValues {
			value.Value = optional.Some(data)
		}
		refs = append(refs, value)
		return true
	})
	slices.SortFunc(refs, func(a, b Value) int {
		return strings.Compare(a.Ref.String(), b.Ref.String())
	})
	return refs, nil
}

func NewMemoryProvider[R Role]() *MemoryProvider[R] {
	return &MemoryProvider[R]{
		config: xsync.NewMapOf[Ref, []byte](),
	}
}
