package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/log"
)

type resetSubscriptionCmd struct {
	Subscription reflection.Ref `arg:"" required:"" help:"Full path of subscription to reset." predictor:"subscriptions"`
	Latest       bool           `flag:"latest" help:"Reset subscription to latest offset." default:"true" negatable:"beginning"`
}

func (s *resetSubscriptionCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	var offset adminpb.SubscriptionOffset
	if s.Latest {
		offset = adminpb.SubscriptionOffset_SUBSCRIPTION_OFFSET_LATEST
	} else {
		offset = adminpb.SubscriptionOffset_SUBSCRIPTION_OFFSET_EARLIEST
	}
	_, err := client.ResetSubscription(ctx, connect.NewRequest(&adminpb.ResetSubscriptionRequest{
		Subscription: s.Subscription.ToProto(),
		Offset:       offset,
	}))
	if err != nil {
		return fmt.Errorf("failed to reset subscription: %w", err)
	}
	log.FromContext(ctx).Infof("Subscription %s reset", s.Subscription)
	return nil
}
