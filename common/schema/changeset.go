package schema

import (
	"time"

	"github.com/block/ftl/internal/key"
)

type ChangesetState int

const (
	ChangesetStateUnspecified ChangesetState = iota
	ChangesetStatePreparing
	ChangesetStatePrepared
	ChangesetStateCleaningUp
	ChangesetStateCommitted
	ChangesetStateRollingBack
	ChangesetStateFailed
)

//protobuf:export
type Changeset struct {
	Key       key.Changeset `protobuf:"1"`
	CreatedAt time.Time     `protobuf:"2"`
	Modules   []*Module     `protobuf:"3"`

	State ChangesetState `protobuf:"4"`
	// Error is present if state is failed.
	Error string `protobuf:"5,optional"`
}
