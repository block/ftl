package schemamirror_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/result"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/schemamirror"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/rpc"
	"github.com/jpillora/backoff"
)

func TestMirror(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(t.Context())

	// Find a free port.
	hostAddr, err := plugin.AllocatePort()
	assert.NoError(t, err, "failed to allocate port for mirror service")
	hostURL, err := url.Parse("http://" + hostAddr.String())
	assert.NoError(t, err, "failed to parse address for mirror service")

	mirrorSvc := schemamirror.NewMirrorService()
	schemaSvc := schemamirror.NewSchemaService(mirrorSvc)

	go func() {
		opts := []rpc.Option{
			rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, schemaSvc),
			rpc.GRPC(ftlv1connect.NewSchemaMirrorServiceHandler, mirrorSvc),
		}
		assert.NoError(t, rpc.Serve(ctx, hostURL, opts...), "services failed")
	}()

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, hostURL.String(), log.Debug)
	mirrorClient := rpc.Dial(ftlv1connect.NewSchemaMirrorServiceClient, hostURL.String(), log.Debug)

	// Mirror service should be available, but the schema service should not yet
	checkDisconnectedReadiness(ctx, t, mirrorClient, schemaClient)

	// Begin pushing schema
	stream := mirrorClient.PushSchema(ctx)
	sendInitialSchema(t, stream)

	// Schema service should now be available and we can begin pulling schema
	assert.NoError(t, rpc.Wait(ctx, backoff.Backoff{}, time.Second*10, schemaClient), "failed to connect to schema service")
	pullSchemaStream, err := schemaClient.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	assert.NoError(t, err, "failed to create PullSchema stream")

	// We receive an initial empty schema when mirror service doesn't have any schema yet
	receiveSchemaUpdate[*schemapb.Notification_FullSchemaNotification](ctx, t, pullSchemaStream)

	// Send schema and an update. Make sure mirror passes it along
	receiveSchemaUpdate[*schemapb.Notification_FullSchemaNotification](ctx, t, pullSchemaStream)
	sendChangesetCreatedNotification(t, stream)
	receiveSchemaUpdate[*schemapb.Notification_ChangesetCreatedNotification](ctx, t, pullSchemaStream)

	// Should only allow one stream pushing schema at a time
	badStream := mirrorClient.PushSchema(ctx)
	_, err = badStream.CloseAndReceive()
	assert.True(t, connect.CodeOf(err) == connect.CodeFailedPrecondition, "stream should fail because one is still active")

	// Disconnect push stream and connect another one
	_, err = stream.CloseAndReceive()
	assert.NoError(t, err, "failed to close PushSchema stream")

	time.Sleep(time.Second) // Give some time for the stream to close
	checkDisconnectedReadiness(ctx, t, mirrorClient, schemaClient)

	stream = mirrorClient.PushSchema(ctx)
	sendInitialSchema(t, stream)
	receiveSchemaUpdate[*schemapb.Notification_FullSchemaNotification](ctx, t, pullSchemaStream)
	sendChangesetCreatedNotification(t, stream)
	receiveSchemaUpdate[*schemapb.Notification_ChangesetCreatedNotification](ctx, t, pullSchemaStream)
}

// checkDisconnectedReadiness makes sure that the mirror service is available, but the schema service is not ready yet.
func checkDisconnectedReadiness(ctx context.Context, t *testing.T, mirrorClient ftlv1connect.SchemaMirrorServiceClient, schemaClient ftlv1connect.SchemaServiceClient) {
	assert.NoError(t, rpc.Wait(ctx, backoff.Backoff{}, time.Second*10, mirrorClient), "failed to connect to mirror service")
	assert.Contains(t, rpc.Wait(ctx, backoff.Backoff{}, time.Second*1, schemaClient).Error(), "service is not ready: Mirror is not receiving schema push updates")
	_, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	assert.Equal(t, connect.CodeOf(err), connect.CodeUnavailable, "GetSchema should fail because schema service is not ready")
}

func sendInitialSchema(t *testing.T, stream *connect.ClientStreamForClient[ftlv1.PushSchemaRequest, ftlv1.PushSchemaResponse]) {
	sch := &schema.Schema{
		Realms: []*schema.Realm{
			{
				External: false,
				Name:     "test",
				Modules: []*schema.Module{
					{
						Name:  "testmodule1",
						Decls: []schema.Decl{},
					},
				},
			},
		},
	}
	assert.NoError(t, stream.Send(&ftlv1.PushSchemaRequest{
		Event: &schemapb.Notification{
			Value: &schemapb.Notification_FullSchemaNotification{
				FullSchemaNotification: &schemapb.FullSchemaNotification{
					Schema: sch.ToProto(),
				},
			},
		},
	}), "initial schema push failed")
}

func sendChangesetCreatedNotification(t *testing.T, stream *connect.ClientStreamForClient[ftlv1.PushSchemaRequest, ftlv1.PushSchemaResponse]) {
	changeset := &schema.Changeset{
		Key: key.NewChangesetKey(),
		RealmChanges: []*schema.RealmChange{
			{
				External: false,
				Name:     "test",
				Modules: []*schema.Module{
					{
						Name: "testmodule2",
					},
				},
			},
		},
	}
	assert.NoError(t, stream.Send(&ftlv1.PushSchemaRequest{
		Event: &schemapb.Notification{
			Value: &schemapb.Notification_ChangesetCreatedNotification{
				ChangesetCreatedNotification: &schemapb.ChangesetCreatedNotification{
					Changeset: changeset.ToProto(),
				},
			},
		},
	}), "initial schema push failed")
}

func receiveSchemaUpdate[E any](ctx context.Context, t *testing.T, stream *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) {
	resultChan := make(chan result.Result[*ftlv1.PullSchemaResponse])
	go func() {
		if stream.Receive() {
			resultChan <- result.Ok(stream.Msg())
		} else {
			resultChan <- result.Err[*ftlv1.PullSchemaResponse](stream.Err())
		}
	}()
	select {
	case r := <-resultChan:
		resp, err := r.Result()
		assert.NoError(t, err, "failed to receive PullSchema response")
		_, ok := resp.Event.Value.(E)
		var expected E
		assert.True(t, ok, "expected event to be of type %T but it was %T", expected, resp.Event.Value)

	case <-time.After(time.Second * 5):
		assert.True(t, false, "timeout waiting for PullSchema response")
	case <-ctx.Done():
		assert.NoError(t, ctx.Err(), "context cancelled while waiting for PullSchema response")
	}
}
