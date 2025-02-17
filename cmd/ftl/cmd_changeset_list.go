package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

type listChangesetCmd struct {
}

func (g *listChangesetCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-changesets-list"}))
	if err != nil {
		return fmt.Errorf("failed to pull schema: %w", err)
	}
	for resp.Receive() {
		msg := resp.Msg()
		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			for _, cs := range e.FullSchemaNotification.Changesets {
				mods := make([]string, 0, len(cs.Modules))
				for _, m := range cs.Modules {
					mods = append(mods, m.Name)
				}
				fmt.Printf("%s\tState: %s\tModules: %v\tRemoving %v\n", cs.Key, cs.State, mods, cs.ToRemove)
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
