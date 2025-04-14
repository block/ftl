package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"reflect"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/puzpuzpuz/xsync/v3"

	pubsubpb "github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/go-runtime/server/rpccontext"
	"github.com/block/ftl/internal/deploymentcontext"
)

type mapCacheEntry struct {
	checksum [32]byte
	output   any
}

// RealFTL is the real implementation of the [internal.FTL] interface using the Controller.
type RealFTL struct {
	dmctx *deploymentcontext.DynamicDeploymentContext
	// Cache for Map() calls
	mapped *xsync.MapOf[uintptr, mapCacheEntry]
}

// New creates a new [RealFTL]
func New(dmctx *deploymentcontext.DynamicDeploymentContext) *RealFTL {
	return &RealFTL{
		dmctx:  dmctx,
		mapped: xsync.NewMapOf[uintptr, mapCacheEntry](),
	}
}

var _ FTL = &RealFTL{}

func (r *RealFTL) GetConfig(_ context.Context, name string, dest any) error {
	return errors.WithStack(r.dmctx.CurrentContext().GetConfig(name, dest))
}

func (r *RealFTL) GetSecret(_ context.Context, name string, dest any) error {
	return errors.WithStack(r.dmctx.CurrentContext().GetSecret(name, dest))
}

func (r *RealFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any, key string) error {
	caller := reflection.CallingVerb()
	if topic.Module != caller.Module {
		return errors.Errorf("can not publish to another module's topic: %s", topic)
	}
	client := rpccontext.ClientFromContext[pubsubpbconnect.PublishServiceClient](ctx)
	body, err := encoding.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}
	_, err = client.PublishEvent(ctx, connect.NewRequest(&pubsubpb.PublishEventRequest{
		Topic:  topic.ToProto(),
		Caller: caller.Name,
		Body:   body,
		Key:    key,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to publish event")
	}
	return nil
}

func (r *RealFTL) CallMap(ctx context.Context, mapper any, value any, mapImpl func(context.Context) (any, error)) any {
	// Compute checksum of the input.
	inputData, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to marshal input data")
	}
	checksum := sha256.Sum256(inputData)

	// Check cache.
	key := reflect.ValueOf(mapper).Pointer()
	cached, ok := r.mapped.Load(key)
	if ok && checksum == cached.checksum {
		return cached.output
	}

	// Map the value
	t, err := mapImpl(ctx)
	if err != nil {
		panic(err)
	}

	// Write the cache back.
	r.mapped.Store(key, mapCacheEntry{
		checksum: checksum,
		output:   t,
	})
	return t
}
