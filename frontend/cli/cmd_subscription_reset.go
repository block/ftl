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
}

func (s *resetSubscriptionCmd) Run(ctx context.Context, client ftlv1connect.AdminServiceClient) error {
	_, err := client.ResetSubscription(ctx, connect.NewRequest(&ftlv1.ResetSubscriptionRequest{
		Subscription: s.Subscription.ToProto(),
	}))
	if err != nil {
		return fmt.Errorf("failed to reset subscription: %w", err)
	}
	log.FromContext(ctx).Infof("Subscription %s reset", s.Subscription)
	return nil
}
