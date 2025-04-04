package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/schema"
)

type psCmd struct {
}

func (s *psCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	status, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return err
	}

	sch, err := schema.SchemaFromProto(status.Msg.Schema)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Format: deployment, N/M replicas running
	format := "%-30s %d\n"
	fmt.Printf("%-30s %s\n", "DEPLOYMENT", "REPLICAS")
	for _, module := range sch.InternalModules() {
		if module.Builtin {
			continue
		}
		fmt.Printf(format, module.GetRuntime().GetDeployment().DeploymentKey, module.GetRuntime().GetScaling().MinReplicas)
	}
	return nil
}
