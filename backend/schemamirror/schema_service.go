package schemamirror

import (
	"context"
	"maps"
	"slices"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/key"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
)

type SchemaService struct {
	mirror *MirrorService
}

func NewSchemaService(mirror *MirrorService) *SchemaService {
	return &SchemaService{mirror: mirror}
}

var _ ftlv1connect.SchemaServiceHandler = (*SchemaService)(nil)

func (s *SchemaService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	if !s.mirror.receiving.Load() {
		reason := "Mirror is not receiving schema push updates"
		return connect.NewResponse(&ftlv1.PingResponse{
			NotReady: &reason,
		}), nil
	}
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// GetSchema gets the full schema.
func (s *SchemaService) GetSchema(context.Context, *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	if !s.mirror.receiving.Load() {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("mirror is not receiving schema updates"))
	}
	return connect.NewResponse(&ftlv1.GetSchemaResponse{
		Schema: s.mirror.eventSource.CanonicalView().ToProto(),
	}), nil
}

// PullSchema streams changes to the schema.
func (s *SchemaService) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	if !s.mirror.receiving.Load() {
		return connect.NewError(connect.CodeUnavailable, errors.New("mirror is not receiving schema updates"))
	}
	if err := stream.Send(&ftlv1.PullSchemaResponse{
		Event: &schemapb.Notification{
			Value: &schemapb.Notification_FullSchemaNotification{
				FullSchemaNotification: &schemapb.FullSchemaNotification{
					Schema: s.mirror.eventSource.CanonicalView().ToProto(),
					Changesets: islices.Map(slices.Collect(maps.Values(s.mirror.eventSource.ActiveChangesets())), func(cs *schema.Changeset) *schemapb.Changeset {
						return cs.ToProto()
					}),
				},
			},
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send initial schema")
	}
	for event := range channels.IterSubscribable[schema.Notification](ctx, s.mirror.eventSource) {
		if err := stream.Send(&ftlv1.PullSchemaResponse{
			Event: schema.NotificationToProto(event),
		}); err != nil {
			return errors.Wrap(err, "failed to send schema update")
		}
	}
	return errors.WithStack(ctx.Err())
}

// GetDeployments is used to get the schema for all deployments.
func (s *SchemaService) GetDeployments(context.Context, *connect.Request[ftlv1.GetDeploymentsRequest]) (*connect.Response[ftlv1.GetDeploymentsResponse], error) {
	if !s.mirror.receiving.Load() {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("mirror is not receiving schema updates"))
	}
	result := []*ftlv1.DeployedSchema{}

	activeDeployments := map[key.Deployment]*schema.Module{}
	for _, realm := range s.mirror.eventSource.CanonicalView().Realms {
		for _, m := range realm.Modules {
			d := m.GetRuntime().GetDeployment()
			if d == nil {
				continue
			}
			activeDeployments[d.DeploymentKey] = m
			result = append(result, &ftlv1.DeployedSchema{
				DeploymentKey: d.DeploymentKey.String(),
				Schema:        m.ToProto(),
				IsActive:      true,
			})
		}
	}
	for _, cs := range s.mirror.eventSource.ActiveChangesets() {
		for _, m := range cs.InternalRealm().Modules {
			d := m.GetRuntime().GetDeployment()
			if d == nil {
				continue
			}
			if activeDeployments[d.DeploymentKey] != nil {
				continue
			}
			activeDeployments[d.DeploymentKey] = m
			result = append(result, &ftlv1.DeployedSchema{
				DeploymentKey: d.DeploymentKey.String(),
				Schema:        m.ToProto(),
				IsActive:      false,
			})
		}
	}
	return connect.NewResponse(&ftlv1.GetDeploymentsResponse{
		Schema: result,
	}), nil
}

// GetDeployment gets a deployment by deployment key
func (s *SchemaService) GetDeployment(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	if !s.mirror.receiving.Load() {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("mirror is not receiving schema updates"))
	}
	deploymentKey, err := key.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key")))
	}
	for _, realm := range s.mirror.eventSource.CanonicalView().Realms {
		for _, m := range realm.Modules {
			d := m.GetRuntime().GetDeployment()
			if d == nil {
				continue
			}
			if deploymentKey == d.DeploymentKey {
				return connect.NewResponse(&ftlv1.GetDeploymentResponse{
					Schema: m.ToProto(),
				}), nil
			}
		}
	}
	for _, cs := range s.mirror.eventSource.ActiveChangesets() {
		for _, m := range cs.InternalRealm().Modules {
			d := m.GetRuntime().GetDeployment()
			if d == nil {
				continue
			}
			if deploymentKey == d.DeploymentKey {
				return connect.NewResponse(&ftlv1.GetDeploymentResponse{
					Schema: m.ToProto(),
				}), nil
			}
		}
	}
	return nil, errors.Wrapf(err, "failed to find deployment %s", deploymentKey)
}

func (s *SchemaService) UpdateDeploymentRuntime(context.Context, *connect.Request[ftlv1.UpdateDeploymentRuntimeRequest]) (*connect.Response[ftlv1.UpdateDeploymentRuntimeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) CreateChangeset(context.Context, *connect.Request[ftlv1.CreateChangesetRequest]) (*connect.Response[ftlv1.CreateChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) PrepareChangeset(context.Context, *connect.Request[ftlv1.PrepareChangesetRequest]) (*connect.Response[ftlv1.PrepareChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) CommitChangeset(context.Context, *connect.Request[ftlv1.CommitChangesetRequest]) (*connect.Response[ftlv1.CommitChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) DrainChangeset(context.Context, *connect.Request[ftlv1.DrainChangesetRequest]) (*connect.Response[ftlv1.DrainChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) FinalizeChangeset(context.Context, *connect.Request[ftlv1.FinalizeChangesetRequest]) (*connect.Response[ftlv1.FinalizeChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) RollbackChangeset(context.Context, *connect.Request[ftlv1.RollbackChangesetRequest]) (*connect.Response[ftlv1.RollbackChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}

func (s *SchemaService) FailChangeset(context.Context, *connect.Request[ftlv1.FailChangesetRequest]) (*connect.Response[ftlv1.FailChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("schema mirror is read only"))
}
