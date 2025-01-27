package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/log"
)

type resetSubscriptionCmd struct {
	Subscription reflection.Ref `arg:"" required:"" help:"Full path of subscription to reset."`
	Latest       bool           `flag:"latest" help:"Reset subscription to latest offset." default:"true" negatable:"beginning"`
}

func (s *resetSubscriptionCmd) Run(ctx context.Context, client ftlv1connect.AdminServiceClient) error {
	var offset ftlv1.SubscriptionOffset
	if s.Latest {
		offset = ftlv1.SubscriptionOffset_SUBSCRIPTION_OFFSET_LATEST
	} else {
		offset = ftlv1.SubscriptionOffset_SUBSCRIPTION_OFFSET_EARLIEST
	}
	_, err := client.ResetSubscription(ctx, connect.NewRequest(&ftlv1.ResetSubscriptionRequest{
		Subscription: s.Subscription.ToProto(),
		Offset:       offset,
	}))
	if err != nil {
		return fmt.Errorf("failed to reset subscription: %w", err)
	}
	log.FromContext(ctx).Infof("Subscription %s reset", s.Subscription)
	return nil
}
