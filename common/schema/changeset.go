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
	// RealmChanges is the list of realm changes in this changeset.
	RealmChanges []*RealmChange `protobuf:"3"`
	// State the changeset state
	State ChangesetState `protobuf:"4"`
	// Error is present if state is failed.
	Error string `protobuf:"5,optional"`
}

type RealmChange struct {
	Name     string `protobuf:"1"`
	External bool   `protobuf:"2"`
	// Modules is the list of modules that will be added or modified in this changeset. Initial this is the canonical state of the modules.
	// Once this changeset is committed this is no longer canonical and it kept for archival purposes.
	Modules []*Module `protobuf:"3"`
	// ModulesToRemove is the list of module names that will be removed in this changeset.
	ToRemove []string `protobuf:"4"`
	// RemovingModules is the list of modules that are being removed in this changeset.
	// Once the changeset is committed this becomes the canonical state of the removing modules
	RemovingModules []*Module `protobuf:"5"`
}

// ModulesAreCanonical returns true if the changeset is in a state when there the Modules field is the canonical state of the modules.
func (c *Changeset) ModulesAreCanonical() bool {
	return c.State == ChangesetStatePreparing || c.State == ChangesetStatePrepared || c.State == ChangesetStateRollingBack
}

// OwnedModules returns the modules that are owned by this changeset for the given realm.
// Depending on the state this may be the added modules list or the removing modules list.
func (c *Changeset) OwnedModules(realm *RealmChange) []*Module {
	if c.ModulesAreCanonical() {
		return realm.Modules
	}
	return realm.RemovingModules
}

func (c *Changeset) InternalModules() []*Module {
	var res []*Module
	for _, realm := range c.RealmChanges {
		if realm.External {
			continue
		}
		res = append(res, realm.Modules...)
	}
	return res
}

func (c *Changeset) InternalToRemove() []string {
	var res []string
	for _, realm := range c.RealmChanges {
		if realm.External {
			continue
		}
		res = append(res, realm.ToRemove...)
	}
	return res
}

func (c *Changeset) InternalRemovingModules() []*Module {
	var res []*Module
	for _, realm := range c.RealmChanges {
		if realm.External {
			continue
		}
		res = append(res, realm.RemovingModules...)
	}
	return res
}
