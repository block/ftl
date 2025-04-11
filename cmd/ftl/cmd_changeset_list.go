package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

type listChangesetCmd struct {
}

func (g *listChangesetCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-changesets-list"}))
	if err != nil {
		return fmt.Errorf("failed to pull schema: %w", err)
	}
	for resp.Receive() {
		msg := resp.Msg()
		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			for _, cs := range e.FullSchemaNotification.Changesets {
				var mods []string
				var removing []string
				for _, rc := range cs.RealmChanges {
					for _, m := range rc.Modules {
						mods = append(mods, moduleWithRealm(m.Name, rc.Name, rc.External))
					}
					for _, m := range rc.ToRemove {
						removing = append(removing, moduleWithRealm(m, rc.Name, false))
					}
				}
				fmt.Printf("%s\tState: %s\tModules: %v\tRemoving %v\n", cs.Key, cs.State, mods, removing)
			}
			return nil
		default:
			// Ignore for now
		}

	}
	if err := resp.Err(); err != nil {
		return fmt.Errorf("failed to pull schema %w", resp.Err())
	}
	return nil
}

func moduleWithRealm(moduleName, realmName string, external bool) string {
	if external {
		return fmt.Sprintf("%s@%s", moduleName, realmName)
	}
	return moduleName
}
