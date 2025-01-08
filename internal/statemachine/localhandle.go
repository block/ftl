package statemachine

import (
	"context"
	"fmt"
	"reflect"

	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
)

type localHandle[Q any, R any, E any] struct {
	sm ListenableStateMachine[Q, R, E]
}

// LocalHandle creates a handle to a state machine that is local to the current process.
func LocalHandle[Q any, R any, E any](sm ListenableStateMachine[Q, R, E]) StateMachineHandle[Q, R, E] {
	return &localHandle[Q, R, E]{
		sm: sm,
	}
}

func (l *localHandle[Q, R, E]) Update(ctx context.Context, msg E) error {
	if err := l.sm.Update(msg); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

func (l *localHandle[Q, R, E]) Query(ctx context.Context, query Q) (R, error) {
	var zero R

	r, err := l.sm.Lookup(query)
	if err != nil {
		return zero, fmt.Errorf("query: %w", err)
	}
	return r, nil
}

func (l *localHandle[Q, R, E]) Changes(ctx context.Context, query Q) (<-chan R, error) {
	logger := log.FromContext(ctx).Scope("local_statemachine")

	subs, err := l.sm.Subscribe(ctx)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	previous, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return channels.FlatMap(subs, func(struct{}) []R {
		r, err := l.Query(ctx, query)
		if err != nil {
			logger.Warnf("query for changes failed: %s", err)
			return nil
		}
		if reflect.DeepEqual(previous, r) {
			return nil
		}
		previous = r
		return []R{r}
	}), nil
}
