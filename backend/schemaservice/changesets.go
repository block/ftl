package schemaservice

import (
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

func (r *SchemaState) ActiveChangeset() optional.Option[*schema.Changeset] {
	for _, changeset := range r.changesets {
		if changeset.State == schema.ChangesetStatePreparing {
			return optional.Some(changeset)
		}
	}
	return optional.None[*schema.Changeset]()
}

func (r *SchemaState) GetChangeset(changeset key.Changeset) (*schema.Changeset, error) {
	c, ok := r.changesets[changeset]
	if !ok {
		return nil, errors.Errorf("changeset %s not found", changeset)
	}
	return c, nil
}

func (r *SchemaState) GetChangesets() map[key.Changeset]*schema.Changeset {
	return r.changesets
}
