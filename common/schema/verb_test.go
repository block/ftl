package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestUpsert(t *testing.T) {
	var metadata []Metadata
	metadata = upsert[MetadataCalls](metadata, &Ref{Name: "foo"})
	metadata = upsert[MetadataCalls](metadata, &Ref{Name: "bar"})
	assert.Equal(t, []Metadata{&MetadataCalls{Calls: []*Ref{{Name: "foo"}, {Name: "bar"}}}}, metadata)

}
