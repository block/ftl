package cron

import (
	"context"
	"net/url"
	"os"
	"sort"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/raft"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type verbClient struct {
	requests chan *ftlv1.CallRequest
}

var _ routing.CallClient = (*verbClient)(nil)

func (v *verbClient) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	v.requests <- req.Msg
	return connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Body{Body: []byte("{}")}}), nil
}

func TestCron(t *testing.T) {
	eventSource := schemaeventsource.NewUnattached()
	module := &schema.Module{
		Name: "echo",
		Runtime: &schema.ModuleRuntime{
			Deployment: &schema.ModuleRuntimeDeployment{
				DeploymentKey: key.NewDeploymentKey("test", "echo"),
			},
		},
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.Unit{},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataCronJob{Cron: "0/2 * * * * *"},
				},
			},
			&schema.Verb{
				Name:     "time",
				Request:  &schema.Unit{},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataCronJob{Cron: "1/2 * * * * *"},
				},
			},
		},
	}
	assert.NoError(t, eventSource.PublishModuleForTest(module))

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{Level: log.Trace}))
	timelineClient := timelineclient.NewClient(ctx, timelineclient.NullConfig)
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	t.Cleanup(cancel)

	schemaEndpoint, err := url.Parse("http://localhost:8892")
	assert.NoError(t, err)

	cfg := Config{
		SchemaServiceEndpoint: schemaEndpoint,
		TimelineConfig:        timelineclient.NullConfig,
		Raft:                  raft.RaftConfig{},
	}

	wg, ctx := errgroup.WithContext(ctx)

	requestsch := make(chan *ftlv1.CallRequest, 8)
	client := &verbClient{
		requests: requestsch,
	}

	wg.Go(func() error { return errors.WithStack(Start(ctx, cfg, eventSource, client, timelineClient)) })

	requests := make([]*ftlv1.CallRequest, 0, 2)

done:
	for range 2 {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out: %s", ctx.Err())

		case request := <-requestsch:
			requests = append(requests, request)
			if len(requests) == 2 {
				break done
			}
		}
	}

	cancel()

	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].Verb.Name < requests[j].Verb.Name
	})
	assert.Equal(t, []*ftlv1.CallRequest{
		{
			Metadata: &ftlv1.Metadata{},
			Verb:     &schemapb.Ref{Module: "echo", Name: "echo"},
			Body:     []byte("{}"),
		},
		{
			Metadata: &ftlv1.Metadata{},
			Verb:     &schemapb.Ref{Module: "echo", Name: "time"},
			Body:     []byte("{}"),
		},
	}, requests, assert.Exclude[*schemapb.Position]())

	err = wg.Wait()
	assert.IsError(t, err, context.Canceled)
}
