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

type killCmd struct {
	Deployment key.Deployment `arg:"" help:"Deployment to kill."`
}

func (k *killCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	//TODO: implement this as a changeset
	update := schema.RuntimeElement{Deployment: k.Deployment, Element: &schema.ModuleRuntimeScaling{MinReplicas: 0}}
	_, err := client.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
		Update: update.ToProto(),
	}))
	if err != nil {
		return fmt.Errorf("failed to kill deployment: %w", err)
	}
	return nil
}
