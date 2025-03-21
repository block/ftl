package raft

import (
	"errors"
	"fmt"
	"io"

	"github.com/lni/dragonboat/v4/statemachine"

	sm "github.com/block/ftl/internal/statemachine"
)

// ErrInvalidEvent is returned if we are attempting to publish an invalid event.
var ErrInvalidEvent = errors.New("invalid event")

const InvalidEventValue = 0x1001

// stateMachineShim is a shim to convert a typed StateMachine to a dragonboat statemachine.IStateMachine.
type stateMachineShim[Q any, R any, E sm.Marshallable, EPtr sm.Unmarshallable[E]] struct {
	sm sm.Snapshotting[Q, R, E]
}

func newStateMachineShim[Q any, R any, E sm.Marshallable, EPtr sm.Unmarshallable[E]](
	sm sm.Snapshotting[Q, R, E],
) statemachine.CreateStateMachineFunc {
	return func(clusterID uint64, nodeID uint64) statemachine.IStateMachine {
		return &stateMachineShim[Q, R, E, EPtr]{sm: sm}
	}
}

func (s *stateMachineShim[Q, R, E, EPtr]) Lookup(key any) (any, error) {
	typed, ok := key.(Q)
	if !ok {
		panic(fmt.Errorf("invalid key type: %T", key))
	}

	res, err := s.sm.Lookup(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup: %w", err)
	}

	return res, nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) Update(entry statemachine.Entry) (statemachine.Result, error) {
	var to E
	toptr := (EPtr)(&to)

	if err := toptr.UnmarshalBinary(entry.Cmd); err != nil {
		return statemachine.Result{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	if err := s.sm.Publish(to); err != nil {
		if errors.Is(err, ErrInvalidEvent) {
			return statemachine.Result{
				Value: InvalidEventValue,
			}, nil
		}
		return statemachine.Result{}, fmt.Errorf("failed to update state machine: %w", err)
	}

	return statemachine.Result{}, nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) Close() error {
	if err := s.sm.Close(); err != nil {
		return fmt.Errorf("failed to close state machine: %w", err)
	}
	return nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) RecoverFromSnapshot(
	reader io.Reader,
	_ []statemachine.SnapshotFile, // do not support extra immutable files for now
	_ <-chan struct{}, // do not support snapshot recovery cancellation for now
) error {
	if err := s.sm.Recover(reader); err != nil {
		return fmt.Errorf("failed to recover from snapshot: %w", err)
	}
	return nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) SaveSnapshot(
	writer io.Writer,
	_ statemachine.ISnapshotFileCollection, // do not support extra immutable files for now
	_ <-chan struct{}, // do not support snapshot save cancellation for now
) error {
	if err := s.sm.Save(writer); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}
	return nil
}
