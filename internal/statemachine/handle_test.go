package statemachine

import (
	"context"
	"iter"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
)

// Mock state machine for testing
type mockStateMachine struct {
	value    string
	updates  []string
	queryErr error

	notifier   *channels.Notifier
	mu         sync.Mutex
	runningCtx context.Context
}

func newMockStateMachine(ctx context.Context, initial string) *mockStateMachine {
	broadcaster := channels.NewNotifier(ctx)

	return &mockStateMachine{
		runningCtx: ctx,
		value:      initial,
		notifier:   broadcaster,
	}
}

func (m *mockStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return m.notifier.Subscribe(), nil
}

func (m *mockStateMachine) Publish(msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value = msg
	m.updates = append(m.updates, msg)
	m.notifier.Notify(m.runningCtx)

	return nil
}

func (m *mockStateMachine) Lookup(_ string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.queryErr != nil {
		return "", m.queryErr
	}
	return m.value, nil
}

func TestLocalHandle(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		mock := newMockStateMachine(ctx, "nitial")
		handle := NewLocalHandle(mock)

		// Test Update
		err := handle.Publish(context.Background(), "new value")
		assert.NoError(t, err)
		assert.Equal(t, "new value", mock.value)

		// Test Query
		result, err := handle.Query(context.Background(), "any query")
		assert.NoError(t, err)
		assert.Equal(t, "new value", result)
	})

	t.Run("StateIter", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		mock := newMockStateMachine(ctx, "initial")
		handle1 := NewLocalHandle(mock)
		handle2 := NewLocalHandle(mock)

		changes1, err := handle1.StateIter(ctx, "any query")
		assert.NoError(t, err)
		changes2, err := handle2.StateIter(ctx, "any query")
		assert.NoError(t, err)

		assert.NoError(t, mock.Publish("updated value"))

		pull1, _ := iter.Pull[string](changes1)
		pull2, _ := iter.Pull[string](changes2)

		v1, ok := pull1()
		assert.True(t, ok)
		assert.Equal(t, "initial", v1)

		v2, ok := pull2()
		assert.True(t, ok)
		assert.Equal(t, "initial", v2)

		v1, ok = pull1()
		assert.True(t, ok)
		assert.Equal(t, "updated value", v1)

		v2, ok = pull2()
		assert.True(t, ok)
		assert.Equal(t, "updated value", v2)
	})
}
