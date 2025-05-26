package statemachine

import (
	"context"
	"iter"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/iterops"
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

	// StateIter returns an iterator of state based on a query.
	//
	// The current state is returned as the first element of the iterator,
	// followed by a stream of states for each change.
	//
	// The iterator is finished when the context is cancelled.
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
		return errors.Wrap(err, "update")
	}
	return nil
}

func (l *localHandle[Q, R, E]) Query(ctx context.Context, query Q) (R, error) {
	var zero R

	r, err := l.sm.Lookup(query)
	if err != nil {
		return zero, errors.Wrap(err, "query")
	}
	return r, nil
}

func (l *localHandle[Q, R, E]) StateIter(ctx context.Context, query Q) (iter.Seq[R], error) {
	logger := log.FromContext(ctx).Scope("local_statemachine")

	subs, err := l.sm.Subscribe(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "subscribe")
	}
	previous, err := l.Query(ctx, query)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return iterops.Concat(
		iterops.Const(previous),
		iterops.FlatMap(channels.IterContext(ctx, subs), func(struct{}) iter.Seq[R] {
			r, err := l.Query(ctx, query)
			if err != nil {
				logger.Warnf("query for changes failed: %s", err)
				return iterops.Empty[R]()
			}
			return iterops.Const(r)
		}),
	), nil
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
		return res, errors.Wrap(err, "query")
	}
	return res, nil
}

// Publish updates the state machine with a new event.
func (h *SingleQueryHandle[Q, R, E]) Publish(ctx context.Context, msg E) error {
	err := h.underlying.Publish(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "publish")
	}
	return nil
}

// StateIter returns a stream of state machine based on a query.
// The current state is returned as the first element of the iterator,
// followed by a stream of states for each change.
//
// The iterator is finished when the context is cancelled.
func (h *SingleQueryHandle[Q, R, E]) StateIter(ctx context.Context) (iter.Seq[R], error) {
	res, err := h.underlying.StateIter(ctx, h.query)
	if err != nil {
		return nil, errors.Wrap(err, "changes")
	}
	return res, nil
}
