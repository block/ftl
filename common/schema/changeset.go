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
	ChangesetStateCommitted
	ChangesetStateDrained
	ChangesetStateFinalized
	ChangesetStateRollingBack
	ChangesetStateFailed
)

//protobuf:export
type Changeset struct {
	Key       key.Changeset `protobuf:"1"`
	CreatedAt time.Time     `protobuf:"2"`
	// Modules is the list of modules that will be added or modified in this changeset. Initial this is the canonical state of the modules.
	// Once this changeset is committed this is no longer canonical and it kept for archival purposes.
	Modules []*Module `protobuf:"3"`
	// ModulesToRemove is the list of module names that will be removed in this changeset.
	ToRemove []string `protobuf:"4"`
	// RemovingModules is the list of modules that are being removed in this changeset.
	// Once the changeset is committed this becomes the canonical state of the removing modules
	RemovingModules []*Module `protobuf:"5"`

	// State the changeset state
	State ChangesetState `protobuf:"6"`
	// Error is present if state is failed.
	Error string `protobuf:"7,optional"`
}

// ModulesAreCanonical returns true if the changeset is in a state when there the Modules field is the canonical state of the modules.
func (c *Changeset) ModulesAreCanonical() bool {
	return c.State == ChangesetStatePreparing || c.State == ChangesetStatePrepared || c.State == ChangesetStateRollingBack
}

// OwnedModules returns the modules that are owned by this changeset. Depending on the state this may be the added modules list or the removing modules list.
func (c *Changeset) OwnedModules() []*Module {
	if c.ModulesAreCanonical() {
		return c.Modules
	}
	return c.RemovingModules
}
