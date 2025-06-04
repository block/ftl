package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
)

type describeChangesetCmd struct {
	Changeset string `arg:"" help:"Changeset to describe."`
}

func (g *describeChangesetCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	// TODO: this needs a lot more thought how to make it look nice, for now it just dumps internals
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-changesets-describe"}))
	if err != nil {
		return errors.Wrap(err, "failed to pull schema")
	}
	for resp.Receive() {
		msg := resp.Msg()
		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			for _, cpb := range e.FullSchemaNotification.Changesets {
				if cpb.Key == g.Changeset {
					cs, err := schema.ChangesetFromProto(cpb)
					if err != nil {
						return errors.Wrap(err, "failed to parse changeset")
					}
					describeChangeset(cs)
					return nil
				}
			}
			return nil
		default:
			// Ignore for now
		}

	}
	if err := resp.Err(); err != nil {
		return errors.Wrap(err, "failed to pull schema")
	}
	return nil
}

func describeChangeset(cs *schema.Changeset) {
	fmt.Printf("Changeset: %s\n", cs.Key)
	fmt.Printf("  Created: %s\n", cs.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Status:  %s\n", cs.State)

	fmt.Printf("Modules: \n\n")
	for _, m := range cs.InternalRealm().Modules {
		fmt.Printf("Module: %s\n", m.Name)
		resources := schema.GetProvisionedResources(m)
		for _, res := range resources {
			fmt.Printf("\t%v: %v %v\n", res.Kind, res.Config, res.State)
		}
	}

	if cs.ModulesAreCanonical() {
		if len(cs.InternalRealm().ToRemove) > 0 {
			fmt.Printf("Module To Remove\n\n")
			for _, m := range cs.InternalRealm().ToRemove {
				fmt.Printf("Module: %s\n", m)
			}
		}
	} else if len(cs.InternalRealm().RemovingModules) > 0 {
		fmt.Printf("Removing Modules\n")
		for _, m := range cs.InternalRealm().ToRemove {
			fmt.Printf("Module: %s\n", m)
		}
	}
}
