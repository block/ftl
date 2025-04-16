package schemaeventsource

import (
	"context"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

func TestSchemaEventSource(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.TODO())
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	t.Cleanup(cancel)

	server := &mockSchemaService{changes: make(chan *ftlv1.PullSchemaResponse, 8)}
	sv, err := rpc.NewServer(ctx, must.Get(url.Parse("http://127.0.0.1:0")), rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, server)) //nolint:errcheck
	assert.NoError(t, err)
	bindChan := sv.Bind.Subscribe(nil)
	defer sv.Bind.Unsubscribe(bindChan)
	go sv.Serve(ctx) //nolint:errcheck
	bind := <-bindChan

	changes := New(ctx, "test", rpc.Dial(ftlv1connect.NewSchemaServiceClient, bind.String(), log.Debug))

	send := func(t testing.TB, resp *ftlv1.PullSchemaResponse) {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())

		case server.changes <- resp:
		}
	}
	events := changes.Subscribe(ctx)

	recv := func(t testing.TB) schema.Notification {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())

		case change := <-events:
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
	time1.ModRuntime().ModDeployment().DeploymentKey = key.NewDeploymentKey("test", "time")
	echo1.ModRuntime().ModDeployment().DeploymentKey = key.NewDeploymentKey("test", "echo")
	time2.ModRuntime().ModDeployment().DeploymentKey = key.NewDeploymentKey("test", "time")
	time1.ModRuntime().ModDeployment().State = schema.DeploymentStateCanonical
	echo1.ModRuntime().ModDeployment().State = schema.DeploymentStateCanonical
	time2.ModRuntime().ModDeployment().State = schema.DeploymentStateCanonical

	t.Run("InitialSend", func(t *testing.T) {
		send(t, &ftlv1.PullSchemaResponse{
			Event: &schemapb.Notification{Value: &schemapb.Notification_FullSchemaNotification{FullSchemaNotification: &schemapb.FullSchemaNotification{
				Schema: &schemapb.Schema{
					Realms: []*schemapb.Realm{{
						Modules: []*schemapb.Module{time1.ToProto()},
					}},
				},
			}}},
		})

		waitCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		assert.True(t, changes.WaitForInitialSync(waitCtx))

		key := key.NewChangesetKey()
		send(t, &ftlv1.PullSchemaResponse{
			Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetCommittedNotification{ChangesetCommittedNotification: &schemapb.ChangesetCommittedNotification{
				Changeset: &schemapb.Changeset{
					Key: key.String(),
					RealmChanges: []*schemapb.RealmChange{{
						Modules: []*schemapb.Module{echo1.ToProto()},
					}},
				},
			}}},
		})

		waitCtx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
		assert.True(t, changes.WaitForInitialSync(waitCtx))

		var expected schema.Notification = &schema.FullSchemaNotification{
			Schema: &schema.Schema{Realms: []*schema.Realm{{Modules: []*schema.Module{time1}}}},
		}
		assertEqual(t, expected, recv(t))

		expected = &schema.ChangesetCommittedNotification{
			Changeset: &schema.Changeset{
				RealmChanges: []*schema.RealmChange{{
					Modules: []*schema.Module{echo1},
				}},
				Key: key,
			},
		}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Realms: []*schema.Realm{{Modules: []*schema.Module{time1, echo1}}}}, changes.CanonicalView())
	})

	t.Run("Mutation", func(t *testing.T) {

		key := key.NewChangesetKey()
		send(t, &ftlv1.PullSchemaResponse{
			Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetCommittedNotification{ChangesetCommittedNotification: &schemapb.ChangesetCommittedNotification{
				Changeset: &schemapb.Changeset{
					Key: key.String(),
					RealmChanges: []*schemapb.RealmChange{{
						Modules: []*schemapb.Module{time2.ToProto()},
					}},
				},
			}}},
		})

		var expected schema.Notification = &schema.ChangesetCommittedNotification{
			Changeset: &schema.Changeset{
				RealmChanges: []*schema.RealmChange{{
					Modules: []*schema.Module{time2},
				}},
				Key: key,
			},
		}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Realms: []*schema.Realm{{Modules: []*schema.Module{time2, echo1}}}}, changes.CanonicalView())
	})

	t.Run("Delete", func(t *testing.T) {

		key := key.NewChangesetKey()
		send(t, &ftlv1.PullSchemaResponse{
			Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetCommittedNotification{ChangesetCommittedNotification: &schemapb.ChangesetCommittedNotification{
				Changeset: &schemapb.Changeset{
					Key: key.String(),
					RealmChanges: []*schemapb.RealmChange{{
						ToRemove:        []string{"echo"},
						RemovingModules: []*schemapb.Module{echo1.ToProto()},
					}},
				},
			}}},
		})

		var expected schema.Notification = &schema.ChangesetCommittedNotification{
			Changeset: &schema.Changeset{
				Key: key,
				RealmChanges: []*schema.RealmChange{{
					ToRemove:        []string{"echo"},
					RemovingModules: []*schema.Module{echo1},
				}},
			},
		}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Realms: []*schema.Realm{{Modules: []*schema.Module{time2}}}}, changes.CanonicalView())
	})
}

type mockSchemaService struct {
	ftlv1connect.UnimplementedSchemaServiceHandler
	changes chan *ftlv1.PullSchemaResponse
}

var _ ftlv1connect.SchemaServiceHandler = &mockSchemaService{}

func (m *mockSchemaService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (m *mockSchemaService) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], resp *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	for change := range channels.IterContext(ctx, m.changes) {
		if err := resp.Send(change); err != nil {
			return errors.Wrap(err, "send change")
		}
	}
	return nil
}

func assertEqual[T comparable](t testing.TB, expected, actual T) {
	t.Helper()
	assert.Equal(t, expected, actual, assert.Exclude[optional.Option[key.Deployment]](), assert.Exclude[*schema.Schema]())
}
