package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/types/optional"
)

// EnvarDecorator overlays an existing provider with one that loads values from environment variables in the form
// FTL_<ROLE>_<MODULE>_<NAME>. eg. FTL_CONFIGURATION_ECHO_USERNAME
type EnvarDecorator[R Role] struct {
	Provider[R]
}

var _ BaseProvider[Configuration] = &EnvarDecorator[Configuration]{}

// NewEnvarDecorator overlays an existing provider with one that loads values from environment variables in the form
// FTL_<ROLE>_<MODULE>_<NAME>. eg. FTL_CONFIGURATION_ECHO_USERNAME
func NewEnvarDecorator[R Role](provider Provider[R]) *EnvarDecorator[R] {
	return &EnvarDecorator[R]{Provider: provider}
}

func (e *EnvarDecorator[R]) Store(ctx context.Context, ref Ref, value []byte) error {
	_ = os.Unsetenv(e.envarName(ref))
	err := e.Provider.Store(ctx, ref, value)
	if err != nil {
		return fmt.Errorf("failed to store value: %w", err)
	}
	return nil
}

func (e *EnvarDecorator[R]) Delete(ctx context.Context, ref Ref) error {
	_ = os.Unsetenv(e.envarName(ref))
	err := e.Provider.Delete(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}
	return nil
}

func (e *EnvarDecorator[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	if value, ok := os.LookupEnv(e.envarName(ref)); ok {
		return []byte(value), nil
	}
	value, err := e.Provider.Load(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to load value: %w", err)
	}
	return value, nil
}

func (e *EnvarDecorator[R]) List(ctx context.Context, withValues bool) ([]Value, error) {
	values, err := e.Provider.List(ctx, withValues)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}
	// If values are requested, also try to load from environment variables.
	if withValues {
		for i, value := range values {
			if envValue, ok := os.LookupEnv(e.envarName(value.Ref)); ok {
				values[i].Value = optional.Some([]byte(envValue))
			}
		}
	}
	return values, nil
}

func (e *EnvarDecorator[R]) envarName(ref Ref) string {
	prefix := fmt.Sprintf("FTL_%s", strings.ToUpper(e.Role().String()))
	if module, ok := ref.Module.Get(); ok {
		return fmt.Sprintf("%s_%s_%s", prefix, strings.ToUpper(module), ref.Name)
	}
	return fmt.Sprintf("%s_%s", prefix, strings.ToUpper(ref.Name))
}

func (e *EnvarDecorator[R]) Close(ctx context.Context) error {
	return nil
}
