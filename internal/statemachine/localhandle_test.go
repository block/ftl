package statemachine

import (
	"context"
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

	broadcaster *channels.Broadcaster[struct{}]
	mu          sync.Mutex
	runningCtx  context.Context
}

func newMockStateMachine(ctx context.Context, initial string) *mockStateMachine {
	broadcaster := &channels.Broadcaster[struct{}]{}
	go broadcaster.Run(ctx)

	return &mockStateMachine{
		runningCtx:  ctx,
		value:       initial,
		broadcaster: broadcaster,
	}
}

func (m *mockStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return m.broadcaster.Subscribe(), nil
}

func (m *mockStateMachine) Update(msg string) error {
	m.mu.Lock()
	m.value = msg
	m.updates = append(m.updates, msg)
	m.mu.Unlock()

	m.broadcaster.Broadcast(m.runningCtx, struct{}{})
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
		handle := LocalHandle(mock)

		// Test Update
		err := handle.Update(context.Background(), "new value")
		assert.NoError(t, err)
		assert.Equal(t, "new value", mock.value)

		// Test Query
		result, err := handle.Query(context.Background(), "any query")
		assert.NoError(t, err)
		assert.Equal(t, "new value", result)
	})

	t.Run("changes channel", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		mock := newMockStateMachine(ctx, "initial")
		handle1 := LocalHandle(mock)
		handle2 := LocalHandle(mock)

		changes1, err := handle1.Changes(ctx, "any query")
		assert.NoError(t, err)
		changes2, err := handle2.Changes(ctx, "any query")
		assert.NoError(t, err)

		assert.NoError(t, mock.Update("updated value"))

		select {
		case newValue := <-changes1:
			assert.Equal(t, "updated value", newValue)
		case <-ctx.Done():
			t.Error("Timed out waiting for changes")
		}
		select {
		case newValue := <-changes2:
			assert.Equal(t, "updated value", newValue)
		case <-ctx.Done():
			t.Error("Timed out waiting for changes")
		}
	})
}
