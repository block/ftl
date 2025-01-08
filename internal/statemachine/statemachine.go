package statemachine

import (
	"context"
	"encoding"
	"io"
)

// StateMachine updates an underlying state based on events.
// It can be queried for the current state.
//
// Q is the query type.
// R is the query response type.
// E is the event type.
type StateMachine[Q any, R any, E any] interface {
	// Query the state of the state machine.
	Lookup(key Q) (R, error)
	// Update the state of the state machine.
	Update(msg E) error
}

// Unmarshallable is a type that can be unmarshalled from a binary representation.
type Unmarshallable[T any] interface {
	*T
	encoding.BinaryUnmarshaler
}

// Marshallable is a type that can be marshalled to a binary representation.
type Marshallable encoding.BinaryMarshaler

// SnapshottingStateMachine adds snapshotting capabilities to a StateMachine.
type SnapshottingStateMachine[Q any, R any, E any] interface {
	StateMachine[Q, R, E]

	// Save the state of the state machine to a snapshot.
	Save(writer io.Writer) error
	// Recover the state of the state machine from a snapshot.
	Recover(reader io.Reader) error
	// Close the state machine.
	Close() error
}

// ListenableStateMachine adds change notification capabilities to a StateMachine.
type ListenableStateMachine[Q any, R any, E any] interface {
	StateMachine[Q, R, E]

	// Subscribe returns a channel that emits an event when the state of the state machine changes.
	Subscribe(ctx context.Context) (<-chan struct{}, error)
}

// StateMachineHandle is a handle to update and query a state machine.
type StateMachineHandle[Q any, R any, E any] interface {
	Update(ctx context.Context, msg E) error
	Query(ctx context.Context, query Q) (R, error)
	Changes(ctx context.Context, query Q) (<-chan R, error)
}
