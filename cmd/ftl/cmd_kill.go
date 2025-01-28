package main

import (
	"context"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/key"
)

type killCmd struct {
	Deployment key.Deployment `arg:"" help:"Deployment to kill."`
}

func (k *killCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	//TODO: implement this as a changeset
	//_, err := client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{DeploymentKey: k.Deployment.String()}))
	//if err != nil {
	//	return err
	//}
	return nil
}
