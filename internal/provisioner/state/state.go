package state

// State of a single resource in execution.
type State interface {
	DebugString() string
}
