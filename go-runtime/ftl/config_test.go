package ftl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/go-runtime/ftl/reflection"
	"github.com/block/ftl/go-runtime/internal"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/modulecontext"
	. "github.com/block/ftl/internal/testutils/modulecontext"
)

func TestConfig(t *testing.T) {
	type C struct {
		One string
		Two string
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	data, err := json.Marshal(C{"one", "two"})
	assert.NoError(t, err)

	moduleCtx := modulecontext.NewBuilder("test").AddConfigs(map[string][]byte{"test": data}).Build()
	ctx = internal.WithContext(ctx, internal.New(MakeDynamic(ctx, moduleCtx)))

	config := Config[C]{Ref: reflection.Ref{Module: "test", Name: "test"}}
	assert.Equal(t, C{"one", "two"}, config.Get(ctx))
}
