package raft

import (
	"context"
	"io"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/internal/channels"
	sm "github.com/block/ftl/internal/statemachine"
)

// notifyingStateMachine is a StateMachine that emits a signal to a Notifier when the state changes.
type notifyingStateMachine[Q any, R any, E any] struct {
	underlying sm.Snapshotting[Q, R, E]
	Notifier   *channels.Notifier
}

var _ sm.Snapshotting[any, any, any] = (*notifyingStateMachine[any, any, any])(nil)

func newNotifyingStateMachine[Q any, R any, E any](ctx context.Context, underlying sm.Snapshotting[Q, R, E]) *notifyingStateMachine[Q, R, E] {
	return &notifyingStateMachine[Q, R, E]{
		underlying: underlying,
		Notifier:   channels.NewNotifier(ctx),
	}
}

func (s *notifyingStateMachine[Q, R, E]) Lookup(key Q) (R, error) {
	res, err := s.underlying.Lookup(key)
	return res, errors.Wrap(err, "failed to lookup")
}

func (s *notifyingStateMachine[Q, R, E]) Publish(msg E) error {
	err := s.underlying.Publish(msg)
	if err != nil {
		return errors.Wrap(err, "failed to publish")
	}
	s.Notifier.Notify(context.Background())
	return nil
}

func (s *notifyingStateMachine[Q, R, E]) Close() error {
	return errors.Wrap(s.underlying.Close(), "failed to close")
}

func (s *notifyingStateMachine[Q, R, E]) Save(writer io.Writer) error {
	return errors.Wrap(s.underlying.Save(writer), "failed to save")
}

func (s *notifyingStateMachine[Q, R, E]) Recover(reader io.Reader) error {
	return errors.Wrap(s.underlying.Recover(reader), "failed to recover")
}
