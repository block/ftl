package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

type updateCmd struct {
	Replicas   int32          `short:"n" help:"Number of replicas to deploy." default:"1"`
	Deployment key.Deployment `arg:"" help:"Deployment to update." predictor:"deployments"`
}

func (u *updateCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	//TODO: implement this as a changeset

	update := schema.RuntimeElement{Deployment: u.Deployment, Element: &schema.ModuleRuntimeScaling{MinReplicas: u.Replicas}}
	_, err := client.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
		Update: update.ToProto(),
	}))
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}
	return nil
}
