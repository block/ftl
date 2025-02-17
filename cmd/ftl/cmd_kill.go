package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type killCmd struct {
	Deployment string `arg:"" help:"Deployment or module to kill." predictor:"deployments"`
}

func (k *killCmd) Run(ctx context.Context, client ftlv1connect.AdminServiceClient, source *schemaeventsource.EventSource) error {
	dep, err := key.ParseDeploymentKey(k.Deployment)
	if err != nil {
		// Assume a module name
		source.WaitForInitialSync(ctx)
		mod, ok := source.CanonicalView().Module(k.Deployment).Get()
		if !ok {
			return fmt.Errorf("deployment %s not found", k.Deployment)
		}
		dep = mod.Runtime.Deployment.DeploymentKey
	}
	_, err = client.ApplyChangeset(ctx, connect.NewRequest(&ftlv1.ApplyChangesetRequest{
		ToRemove: []string{dep.String()},
	}))
	if err != nil {
		return fmt.Errorf("failed to kill deployment: %w", err)
	}
	return nil
}
