package main

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
)

type changesetCmd struct {
	List changesetListCmd `cmd:"" help:"List active changesets."`
	Fail changesetFailCmd `cmd:"" help:"Fail an active changeset."`
}

type changesetListCmd struct {
}

func (c *changesetListCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	resp, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}
	if len(resp.Msg.Changesets) == 0 {
		fmt.Println("No active changesets.")
		return nil
	}
	fmt.Println("Active changesets:")
	for _, changeset := range resp.Msg.Changesets {
		fmt.Printf("%s\n\tCreated: %s\n\tState: %s\n\tModules: %s\n\tRemoving Modules: %s\n",
			changeset.Key,
			changeset.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
			stringForChangesetState(changeset.State),
			strings.Join(slices.Map(changeset.Modules, func(m *schemapb.Module) string { return m.Name }), ", "),
			strings.Join(slices.Map(changeset.RemovingModules, func(m *schemapb.Module) string { return m.Name }), ", "),
		)
	}
	return nil
}

func stringForChangesetState(state schemapb.ChangesetState) string {
	switch state {
	case schemapb.ChangesetState_CHANGESET_STATE_PREPARING:
		return "preparing"
	case schemapb.ChangesetState_CHANGESET_STATE_PREPARED:
		return "prepared"
	case schemapb.ChangesetState_CHANGESET_STATE_COMMITTED:
		return "committed"
	case schemapb.ChangesetState_CHANGESET_STATE_DRAINED:
		return "drained"
	case schemapb.ChangesetState_CHANGESET_STATE_FINALIZED:
		return "finalized"
	case schemapb.ChangesetState_CHANGESET_STATE_ROLLING_BACK:
		return "rollingback"
	case schemapb.ChangesetState_CHANGESET_STATE_FAILED:
		return "failed"
	default:
		panic(fmt.Sprintf("unexpected changeset state %v", state))
	}
}

type changesetFailCmd struct {
	Changeset key.Changeset `arg:"" help:"Changeset key to fail." required:""`
	Error     string        `arg:"" help:"Error message to fail the changeset with." required:""`
}

func (c *changesetFailCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	_, err := client.FailChangeset(ctx, connect.NewRequest(&ftlv1.FailChangesetRequest{
		Changeset: c.Changeset.String(),
		Error:     c.Error,
	}))
	if err != nil {
		return fmt.Errorf("could not fail changeset: %w", err)
	}
	fmt.Printf("Done")
	return nil
}
