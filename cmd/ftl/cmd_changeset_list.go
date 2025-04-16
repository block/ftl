package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
)

type listChangesetCmd struct {
}

func (g *listChangesetCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-changesets-list"}))
	if err != nil {
		return errors.Wrap(err, "failed to pull schema")
	}
	for resp.Receive() {
		msg := resp.Msg()
		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			for _, cpb := range e.FullSchemaNotification.Changesets {
				cs, err := schema.ChangesetFromProto(cpb)
				if err != nil {
					return errors.Wrap(err, "failed to parse changeset")
				}
				mods := make([]string, 0, len(cs.InternalRealm().Modules))
				for _, m := range cs.InternalRealm().Modules {
					mods = append(mods, m.Name)
				}
				fmt.Printf("%s\tState: %s\tModules: %v\tRemoving %v\n", cs.Key, cs.State, mods, cs.InternalRealm().ToRemove)
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
