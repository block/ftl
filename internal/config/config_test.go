package config_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/config"
)

// sortConfigValues sorts config values by module and name for consistent comparison
func sortConfigValues(values []config.Value) []config.Value {
	sorted := make([]config.Value, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		// First compare modules
		iMod, _ := sorted[i].Ref.Module.Get()
		jMod, _ := sorted[j].Ref.Module.Get()
		if iMod != jMod {
			return iMod < jMod
		}
		// If modules are equal, compare names
		return sorted[i].Ref.Name < sorted[j].Ref.Name
	})
	return sorted
}

type syncIntervalOverride[R config.Role] struct {
	SyncIntervalOverride time.Duration
	config.AsynchronousProvider[R]
}

func (s syncIntervalOverride[R]) SyncInterval() time.Duration { return s.SyncIntervalOverride }

// OverrideSyncInterval creates a new asynchronous provider with a custom sync interval.
func OverrideSyncInterval[R config.Role](provider config.AsynchronousProvider[R], interval time.Duration) config.AsynchronousProvider[R] {
	return syncIntervalOverride[R]{
		AsynchronousProvider: provider,
		SyncIntervalOverride: interval,
	}
}

// Shared test function
func testConfig[R config.Role](t *testing.T, ctx context.Context, provider config.Provider[R]) { //nolint
	ctx, cancel := context.WithCancel(ctx) //nolint:forbidigo
	t.Cleanup(cancel)
	values, err := provider.List(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, []config.Value{}, values)

	ref1 := config.Ref{Module: optional.Some("echo"), Name: "name"}
	value1 := []byte(`"Alice"`)
	ref2 := config.Ref{Name: "age"}
	value2 := []byte(`30`)

	err = provider.Store(ctx, ref1, value1)
	assert.NoError(t, err)

	values, err = provider.List(ctx, true)
	assert.NoError(t, err)
	expected := []config.Value{{
		Ref:   ref1,
		Value: optional.Some(value1),
	}}
	assert.Equal(t, expected, sortConfigValues(values), "%s", repr.String(values))

	data, err := provider.Load(ctx, ref1)
	assert.NoError(t, err)
	assert.Equal(t, value1, data)

	err = provider.Store(ctx, ref2, value2)
	assert.NoError(t, err)

	values, err = provider.List(ctx, true)
	assert.NoError(t, err)
	expected = []config.Value{
		{
			Ref:   ref2,
			Value: optional.Some(value2),
		},
		{
			Ref:   ref1,
			Value: optional.Some(value1),
		},
	}
	assert.Equal(t, sortConfigValues(expected), sortConfigValues(values), "%s", repr.String(values))

	err = provider.Delete(ctx, ref1)
	assert.NoError(t, err)

	values, err = provider.List(ctx, true)
	assert.NoError(t, err)
	expected = []config.Value{
		{
			Ref:   ref2,
			Value: optional.Some(value2),
		},
	}
	assert.Equal(t, sortConfigValues(expected), sortConfigValues(values))
}
