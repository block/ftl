package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type updateCmd struct {
	Replicas   int32  `short:"n" help:"Number of replicas to deploy." default:"1"`
	Deployment string `arg:"" help:"Deployment to update." predictor:"deployments"`
}

func (u *updateCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient, source *schemaeventsource.EventSource, projConfig projectconfig.Config) error {
	dep, err := key.ParseDeploymentKey(u.Deployment)
	if err != nil {
		// Assume a module name
		source.WaitForInitialSync(ctx)
		mod, ok := source.CanonicalView().Module(projConfig.Name, u.Deployment).Get()
		if !ok {
			return errors.Errorf("deployment %s not found", u.Deployment)
		}
		dep = mod.Runtime.Deployment.DeploymentKey
	}
	update := schema.RuntimeElement{Deployment: dep, Element: &schema.ModuleRuntimeScaling{MinReplicas: u.Replicas}}

	_, err = client.UpdateDeploymentRuntime(ctx, connect.NewRequest(&adminpb.UpdateDeploymentRuntimeRequest{
		Element: update.ToProto(),
	}))
	if err != nil {
		return errors.Wrap(err, "failed to update deployment")
	}
	return nil
}
