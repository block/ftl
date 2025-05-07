package schemaservice

import (
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

func (r *SchemaState) ActiveChangeset() optional.Option[*schema.Changeset] {
	for _, changeset := range r.state.Changesets {
		if changeset.State == schema.ChangesetStatePreparing {
			return optional.Some(changeset)
		}
	}
	return optional.None[*schema.Changeset]()
}

func (r *SchemaState) GetChangeset(changeset key.Changeset) optional.Option[*schema.Changeset] {
	for _, cs := range r.state.Changesets {
		if cs.Key == changeset {
			return optional.Some(cs)
		}
	}
	return optional.None[*schema.Changeset]()
}

func (r *SchemaState) GetChangesets() map[key.Changeset]*schema.Changeset {
	ret := map[key.Changeset]*schema.Changeset{}
	for _, cs := range r.state.Changesets {
		ret[cs.Key] = cs
	}
	return ret
}
