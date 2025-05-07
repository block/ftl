package schema

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

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

func (s ChangesetState) String() string {
	switch s {
	case ChangesetStateUnspecified:
		return "unspecified"
	case ChangesetStatePreparing:
		return "preparing"
	case ChangesetStatePrepared:
		return "prepared"
	case ChangesetStateCommitted:
		return "committed"
	case ChangesetStateDrained:
		return "drained"
	case ChangesetStateFinalized:
		return "finalized"
	case ChangesetStateRollingBack:
		return "rolling back"
	case ChangesetStateFailed:
		return "failed"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

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
	// ToRemove is the list of deployment keys that will be removed in this changeset.
	ToRemove []string `protobuf:"4"`
	// RemovingModules is the list of modules that are being removed in this changeset.
	// Once the changeset is committed this becomes the canonical state of the removing modules
	RemovingModules []*Module `protobuf:"5"`
}

// ModulesAreCanonical returns true if the changeset is in a state when there the Modules field is the canonical state of the modules.
func (c *Changeset) ModulesAreCanonical() bool {
	return c.State == ChangesetStatePreparing || c.State == ChangesetStatePrepared || c.State == ChangesetStateRollingBack
}

// InternalRealm returns the internal realm for the given changeset.
//
// For now, all changesets have exactly one internal realm change.
func (c *Changeset) InternalRealm() *RealmChange {
	for _, realm := range c.RealmChanges {
		if !realm.External {
			return realm
		}
	}
	return nil
}

// Realm returns the realm change for the given name.
func (c *Changeset) Realm(name string) optional.Option[*RealmChange] {
	for _, realm := range c.RealmChanges {
		if realm.Name == name {
			return optional.Some(realm)
		}
	}
	return optional.None[*RealmChange]()
}

// OwnedModules returns the modules that are owned by this changeset for the given realm.
// Depending on the state this may be the added modules list or the removing modules list.
func (c *Changeset) OwnedModules(realm *RealmChange) []*Module {
	if c.ModulesAreCanonical() {
		return realm.Modules
	}
	return realm.RemovingModules
}
