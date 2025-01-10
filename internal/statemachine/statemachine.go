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
	// Publish an event to the state machine.
	Publish(msg E) error
}

// Unmarshallable is a type that can be unmarshalled from a binary representation.
type Unmarshallable[T any] interface {
	*T
	encoding.BinaryUnmarshaler
}

// Marshallable is a type that can be marshalled to a binary representation.
type Marshallable encoding.BinaryMarshaler

// Snapshotting adds snapshotting capabilities to a StateMachine.
type Snapshotting[Q any, R any, E any] interface {
	StateMachine[Q, R, E]

	// Save the state of the state machine to a snapshot.
	Save(writer io.Writer) error
	// Recover the state of the state machine from a snapshot.
	Recover(reader io.Reader) error
	// Close the state machine.
	Close() error
}

// Listenable adds change notification capabilities to a StateMachine.
type Listenable[Q any, R any, E any] interface {
	StateMachine[Q, R, E]

	// Subscribe returns a channel that emits an event when the state of the state machine changes.
	Subscribe(ctx context.Context) (<-chan struct{}, error)
}
