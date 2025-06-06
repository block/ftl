package buildengine_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/profiles"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

func TestGraph(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx, cancel := context.WithCancelCause(log.ContextWithNewDefaultLogger(context.Background()))
	t.Cleanup(func() {
		cancel(errors.Wrap(context.Canceled, "test complete"))
	})
	root := t.TempDir()
	project := profiles.InitForTesting(t, root)

	engine, err := buildengine.New(ctx, nil, schemaeventsource.NewUnattached(), project.Config(), []string{"testdata/alpha", "testdata/other", "testdata/another"}, true)
	assert.NoError(t, err)

	defer engine.Close()

	// Import the schema from the third module, simulating a remote schema.
	otherSchema := &schema.Module{
		Name: "other",
		Decls: []schema.Decl{
			&schema.Data{
				Name: "EchoRequest",
				Fields: []*schema.Field{
					{Name: "name", Type: &schema.Optional{Type: &schema.String{}}, Metadata: []schema.Metadata{&schema.MetadataAlias{Alias: "name"}}},
				},
			},
			&schema.Data{
				Name: "EchoResponse",
				Fields: []*schema.Field{
					{Name: "message", Type: &schema.String{}, Metadata: []schema.Metadata{&schema.MetadataAlias{Alias: "message"}}},
				},
			},
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.Ref{Module: "other", Name: "EchoRequest"},
				Response: &schema.Ref{Module: "other", Name: "EchoResponse"},
			},
		},
	}
	engine.Import(ctx, "test", otherSchema)

	expected := map[string][]string{
		"alpha":   {"another", "other", "builtin"},
		"another": {"builtin"},
		"other":   {"another", "builtin"},
		"builtin": {},
	}
	graph, err := engine.Graph()
	assert.NoError(t, err)
	assert.Equal(t, expected, graph)
	err = engine.Build(ctx)
	assert.NoError(t, err)
	for _, module := range []string{
		"alpha",
		"other",
		"another",
	} {
		_, success := engine.GetModuleSchema(module)
		assert.True(t, success, "expected schema for %s to be found", module)
	}
}
