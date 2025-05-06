package app

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/schema"
)

type psCmd struct {
}

func (s *psCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	status, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}

	sch, err := schema.SchemaFromProto(status.Msg.Schema)
	if err != nil {
		return errors.Wrap(err, "failed to parse schema")
	}

	// Format: deployment, N/M replicas running
	format := "%-30s %d\n"
	fmt.Printf("%-30s %s\n", "DEPLOYMENT", "REPLICAS")
	for _, realm := range sch.InternalRealms() {
		for _, module := range realm.Modules {
			if module.Builtin {
				continue
			}
			fmt.Printf(format, module.GetRuntime().GetDeployment().DeploymentKey, module.GetRuntime().GetScaling().MinReplicas)
		}
	}
	for _, pcs := range status.Msg.Changesets {
		cs, err := schema.ChangesetFromProto(pcs)
		if err != nil {
			return errors.Wrap(err, "failed to parse changeset")
		}
		for _, rc := range cs.RealmChanges {
			for _, module := range cs.OwnedModules(rc) {
				fmt.Printf(format, module.GetRuntime().GetDeployment().DeploymentKey, module.GetRuntime().GetScaling().MinReplicas)
			}
		}

	}
	return nil
}
