package schemaeventsource

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/rpc"
)

func TestSchemaEventSource(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.TODO())
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	t.Cleanup(cancel)

	server := &mockSchemaService{changes: make(chan *ftlv1.WatchResponse, 8)}
	sv, err := rpc.NewServer(ctx, must.Get(url.Parse("http://127.0.0.1:0")), rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, server)) //nolint:errcheck
	assert.NoError(t, err)
	bindChan := sv.Bind.Subscribe(nil)
	defer sv.Bind.Unsubscribe(bindChan)
	go sv.Serve(ctx) //nolint:errcheck
	bind := <-bindChan

	changes := New(ctx, rpc.Dial(ftlv1connect.NewSchemaServiceClient, bind.String(), log.Debug))

	send := func(t testing.TB, resp *ftlv1.WatchResponse) {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())

		case server.changes <- resp:
		}
	}

	recv := func(t testing.TB) Event {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())

		case change := <-changes.Events():
			return change

		}
		panic("unreachable")
	}

	time1 := &schema.Module{
		Name: "time",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "time",
				Request:  &schema.Unit{},
				Response: &schema.Time{},
			},
		},
	}
	echo1 := &schema.Module{
		Name: "echo",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.String{},
				Response: &schema.String{},
			},
		},
	}
	echo2 := &schema.Module{
		Name: "echo",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "echo2",
				Request:  &schema.String{},
				Response: &schema.String{},
			},
		},
	}
	time2 := &schema.Module{
		Name: "time",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "time",
				Request:  &schema.Unit{},
				Response: &schema.Time{},
			},
			&schema.Verb{
				Name:     "timezone",
				Request:  &schema.Unit{},
				Response: &schema.String{},
			},
		},
	}

	t.Run("InitialSend", func(t *testing.T) {
		waitCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		assert.False(t, changes.WaitForInitialSync(waitCtx))

		// initial schema does not produce events
		send(t, &ftlv1.WatchResponse{Schema: &schemapb.Schema{Modules: []*schemapb.Module{}}})

		send(t, &ftlv1.WatchResponse{Schema: &schemapb.Schema{Modules: []*schemapb.Module{time1.ToProto(), echo1.ToProto()}}})

		waitCtx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
		assert.True(t, changes.WaitForInitialSync(waitCtx))

		var expected Event = EventUpsert{Module: time1}
		assertEqual(t, expected, recv(t))

		expected = EventUpsert{Module: echo1}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time1, echo1}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})

	t.Run("Mutation", func(t *testing.T) {
		send(t, &ftlv1.WatchResponse{
			Schema: &schemapb.Schema{Modules: []*schemapb.Module{time2.ToProto(), echo1.ToProto()}},
		})

		var expected Event = EventUpsert{Module: time2}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time2, echo1}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})

	// Verify that schemasync doesn't propagate "initial" again.
	t.Run("SimulatedReconnect", func(t *testing.T) {
		send(t, &ftlv1.WatchResponse{
			Schema: &schemapb.Schema{Modules: []*schemapb.Module{time2.ToProto(), echo1.ToProto()}},
		})
		send(t, &ftlv1.WatchResponse{
			Schema: &schemapb.Schema{Modules: []*schemapb.Module{time2.ToProto(), echo2.ToProto()}},
		})

		var expected Event = EventUpsert{Module: echo2}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time2, echo2}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})

	t.Run("Delete", func(t *testing.T) {
		send(t, &ftlv1.WatchResponse{
			Schema: &schemapb.Schema{Modules: []*schemapb.Module{time2.ToProto()}},
		})
		var expected Event = EventRemove{Module: echo2}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time2}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})
}

type mockSchemaService struct {
	ftlv1connect.UnimplementedSchemaServiceHandler
	changes chan *ftlv1.WatchResponse
}

var _ ftlv1connect.SchemaServiceHandler = &mockSchemaService{}

func (m *mockSchemaService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (m *mockSchemaService) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], resp *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	return nil
}

func (m *mockSchemaService) Watch(ctx context.Context, req *connect.Request[ftlv1.WatchRequest], resp *connect.ServerStream[ftlv1.WatchResponse]) error {
	for change := range channels.IterContext(ctx, m.changes) {
		if err := resp.Send(change); err != nil {
			return fmt.Errorf("send change: %w", err)
		}
	}
	return nil
}

func assertEqual[T comparable](t testing.TB, expected, actual T) {
	t.Helper()
	assert.Equal(t, expected, actual, assert.Exclude[optional.Option[model.DeploymentKey]](), assert.Exclude[*schema.Schema]())
}
