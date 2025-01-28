package main

import (
	"context"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/key"
)

type updateCmd struct {
	Replicas   int32          `short:"n" help:"Number of replicas to deploy." default:"1"`
	Deployment key.Deployment `arg:"" help:"Deployment to update."`
}

func (u *updateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	//TODO: implement this as a changeset
	//_, err := client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{
	//	DeploymentKey: u.Deployment.String(),
	//	MinReplicas:   &u.Replicas,
	//}))
	//if err != nil {
	//	return err
	//}
	return nil
}
