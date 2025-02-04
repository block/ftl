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

type killCmd struct {
	Deployment key.Deployment `arg:"" help:"Deployment to kill."`
}

func (k *killCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	//TODO: implement this as a changeset
	_, err := client.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
		Event: &schemapb.ModuleRuntimeEvent{
			Key:     k.Deployment.String(),
			Scaling: &schemapb.ModuleRuntimeScaling{MinReplicas: 0},
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to kill deployment: %w", err)
	}
	return nil
}
