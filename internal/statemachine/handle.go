package statemachine

import (
	"context"
	"fmt"
	"iter"
	"reflect"

	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/log"
)

// Handle to update and query a state machine.
//
// Handle abstracts away the access to a statemachine from its implementation.
// For example, the state machine can be running remotely, and the handle would be a client to the remote system.
// One statemachine can have multiple handles attached to it.
type Handle[Q any, R any, E any] interface {
	// Publish updates the state machine with a new event.
	Publish(ctx context.Context, msg E) error

	// Query retrieves the current state of the state machine.
	Query(ctx context.Context, query Q) (R, error)

	// StateIter returns a stream of state based on a query.
	StateIter(ctx context.Context, query Q) (iter.Seq[R], error)
}

type localHandle[Q any, R any, E any] struct {
	sm Listenable[Q, R, E]
}

var _ Handle[any, any, any] = (*localHandle[any, any, any])(nil)

// NewLocalHandle creates a handle to a local Listenable StateMachine.
func NewLocalHandle[Q any, R any, E any](sm Listenable[Q, R, E]) Handle[Q, R, E] {
	return &localHandle[Q, R, E]{
		sm: sm,
	}
}

func (l *localHandle[Q, R, E]) Publish(ctx context.Context, msg E) error {
	if err := l.sm.Publish(msg); err != nil {
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

func (l *localHandle[Q, R, E]) StateIter(ctx context.Context, query Q) (iter.Seq[R], error) {
	logger := log.FromContext(ctx).Scope("local_statemachine")

	subs, err := l.sm.Subscribe(ctx)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	previous, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return iterops.FlatMap(channels.IterContext(ctx, subs), func(struct{}) []R {
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

// SingleQueryHandle is a handle to a state machine that only supports a single query.
//
// This is a convenience interface to avoid repeating the same constant query.
type SingleQueryHandle[Q any, R any, E any] struct {
	underlying Handle[Q, R, E]
	query      Q
}

func NewSingleQueryHandle[Q any, R any, E any](underlying Handle[Q, R, E], query Q) *SingleQueryHandle[Q, R, E] {
	return &SingleQueryHandle[Q, R, E]{
		underlying: underlying,
		query:      query,
	}
}

// View retrieves the current state of the state machine.
func (h *SingleQueryHandle[Q, R, E]) View(ctx context.Context) (R, error) {
	res, err := h.underlying.Query(ctx, h.query)
	if err != nil {
		return res, fmt.Errorf("query: %w", err)
	}
	return res, nil
}

// Publish updates the state machine with a new event.
func (h *SingleQueryHandle[Q, R, E]) Publish(ctx context.Context, msg E) error {
	err := h.underlying.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

// StateIter returns a stream of state machine based on a query.
//
// The iterator is finished when the context is cancelled.
func (h *SingleQueryHandle[Q, R, E]) StateIter(ctx context.Context) (iter.Seq[R], error) {
	res, err := h.underlying.StateIter(ctx, h.query)
	if err != nil {
		return nil, fmt.Errorf("changes: %w", err)
	}
	return res, nil
}
