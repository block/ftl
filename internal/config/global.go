package config

import (
	"context"

	"github.com/alecthomas/errors"
)

// GlobalFallbackDecorator is a [Provider] that overloads [Provider.Load] to return global config if the key is not
// available in the module.
type GlobalFallbackDecorator[R Role] struct {
	Provider[R]
}

func NewGlobalFallbackDecorator[R Role](provider Provider[R]) *GlobalFallbackDecorator[R] {
	return &GlobalFallbackDecorator[R]{Provider: provider}
}

func (g *GlobalFallbackDecorator[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	value, err := g.Provider.Load(ctx, ref)
	if errors.Is(err, ErrNotFound) && ref.Module.Ok() {
		// Global fallback.
		return errors.WithStack2(g.Provider.Load(ctx, ref.WithoutModule()))
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	return value, nil
}
