package raft

import (
	"context"
	"encoding"
	"io"
	"iter"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/internal/eventstream"
	sm "github.com/block/ftl/internal/statemachine"
)

type UnitQuery struct{}

type RaftStreamEvent[View encoding.BinaryMarshaler, VPtr sm.Unmarshallable[View]] interface {
	encoding.BinaryMarshaler
	eventstream.Event[View]
}

type RaftEventView[V encoding.BinaryMarshaler, VPrt sm.Unmarshallable[V], E RaftStreamEvent[V, VPrt]] struct {
	shard sm.Handle[UnitQuery, V, E]
}

func (s *RaftEventView[V, VPrt, E]) Publish(ctx context.Context, event E) error {
	if err := s.shard.Publish(ctx, event); err != nil {
		return errors.Wrap(err, "failed to update shard")
	}
	return nil
}

func (s *RaftEventView[V, VPrt, E]) View(ctx context.Context) (V, error) {
	var zero V

	view, err := s.shard.Query(ctx, UnitQuery{})
	if err != nil {
		return zero, errors.WithStack(err)
	}

	return view, nil
}

func (s *RaftEventView[V, VPrt, E]) Changes(ctx context.Context) (iter.Seq[V], error) {
	res, err := s.shard.StateIter(ctx, UnitQuery{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get changes")
	}

	return res, nil
}

type eventStreamStateMachine[
	V encoding.BinaryMarshaler,
	VPrt sm.Unmarshallable[V],
	E RaftStreamEvent[V, VPrt],
	EPtr sm.Unmarshallable[E],
] struct {
	view V
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Close() error {
	return nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Lookup(key UnitQuery) (V, error) {
	return s.view, nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Publish(msg E) error {
	v, err := msg.Handle(s.view)
	if err != nil {
		return errors.Wrap(err, "failed to handle event")
	}
	s.view = v
	return nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Save(writer io.Writer) error {
	bytes, err := s.view.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "failed to marshal view")
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "failed to write view")
	}
	return nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Recover(reader io.Reader) error {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "failed to read view")
	}
	if err := (VPrt)(&s.view).UnmarshalBinary(bytes); err != nil {
		return errors.Wrap(err, "failed to unmarshal view")
	}
	return nil
}

// AddEventView to the Builder
func AddEventView[
	V encoding.BinaryMarshaler,
	VPtr sm.Unmarshallable[V],
	E RaftStreamEvent[V, VPtr],
	EPtr sm.Unmarshallable[E],
](ctx context.Context, builder *Builder, shardID uint64) eventstream.EventView[V, E] {
	sm := &eventStreamStateMachine[V, VPtr, E, EPtr]{}
	shard := AddShard[UnitQuery, V, E, EPtr](ctx, builder, shardID, sm)
	return &RaftEventView[V, VPtr, E]{shard: shard}
}
