package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/internal/key"
)

type updateCmd struct {
	Replicas   int32          `short:"n" help:"Number of replicas to deploy." default:"1"`
	Deployment key.Deployment `arg:"" help:"Deployment to update."`
}

func (u *updateCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	//TODO: implement this as a changeset
	_, err := client.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
		Deployment: u.Deployment.String(),
		Event: &schemapb.ModuleRuntimeEvent{
			DeploymentKey: u.Deployment.String(),
			Scaling:       &schemapb.ModuleRuntimeScaling{MinReplicas: u.Replicas},
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}
	return nil
}
