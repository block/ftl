package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/config"
)

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

	t.Cleanup(func() {
		err := provider.Close(ctx)
		assert.NoError(t, err)
	})

	values, err := provider.List(ctx, true, None[string]())
	assert.NoError(t, err)
	assert.Equal(t, []config.Value{}, values)

	ref1 := config.Ref{Module: Some("echo"), Name: "name"}
	value1 := []byte(`"Alice"`)

	ref2 := config.Ref{Name: "age"}
	value2 := []byte(`30`)

	ref3 := config.Ref{Name: "name"}
	value3 := []byte(`"Bob"`)

	err = provider.Store(ctx, ref3, value3)
	assert.NoError(t, err)

	_, err = provider.Load(ctx, ref1)
	assert.IsError(t, err, config.ErrNotFound, "providers should not fall back to global, use GlobalFallbackDecorator for this purpose")

	err = provider.Store(ctx, ref1, value1)
	assert.NoError(t, err)

	values, err = provider.List(ctx, true, None[string]())
	assert.NoError(t, err)
	assert.Equal(t, []config.Value{
		{
			Ref:   ref1,
			Value: Some(value1),
		},
		{
			Ref:   ref3,
			Value: Some(value3),
		},
	}, values, "%s", repr.String(values))

	data, err := provider.Load(ctx, ref1)
	assert.NoError(t, err)
	assert.Equal(t, value1, data)

	err = provider.Store(ctx, ref2, value2)
	assert.NoError(t, err)

	values, err = provider.List(ctx, true, None[string]())
	assert.NoError(t, err)
	assert.Equal(t, []config.Value{
		{
			Ref:   ref2,
			Value: Some(value2),
		},
		{
			Ref:   ref1,
			Value: Some(value1),
		},
		{
			Ref:   ref3,
			Value: Some(value3),
		},
	}, values, "%s", repr.String(values))

	err = provider.Delete(ctx, ref2)
	assert.NoError(t, err)

	values, err = provider.List(ctx, true, None[string]())
	assert.NoError(t, err)
	assert.Equal(t, []config.Value{
		{
			Ref:   ref1,
			Value: Some(value1),
		},
		{
			Ref:   ref3,
			Value: Some(value3),
		},
	}, values)

	// Test global fallbacks work correctly.
	globalFallback := config.NewGlobalFallbackDecorator(provider)
	data, err = globalFallback.Load(ctx, ref1)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"Alice"`), data)

	err = globalFallback.Delete(ctx, ref1)
	assert.NoError(t, err)

	data, err = globalFallback.Load(ctx, ref1)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"Bob"`), data)
}
