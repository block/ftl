package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/reflection"
)

type pubsubCmd struct {
	Subscription subscriptionCmd `cmd:"" help:"Manage subscriptions."`
	Topic        topicCmd        `cmd:"" help:"Manage topics."`
}

type topicCmd struct {
	Info topicInfoCmd `cmd:"" help:"Get info about a topic."`
}

type topicInfoCmd struct {
	Topics []reflection.Ref `arg:"" required:"" help:"Topics to get info about." predictor:"topics"`
}

func (t *topicInfoCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	resps := make([]*adminpb.GetTopicInfoResponse, 0, len(t.Topics))
	for _, top := range t.Topics {
		resp, err := client.GetTopicInfo(ctx, connect.NewRequest(&adminpb.GetTopicInfoRequest{
			Topic: top.ToProto(),
		}))
		if err != nil {
			return fmt.Errorf("failed to get topic info: %w", err)
		}
		resps = append(resps, resp.Msg)
	}
	indent := "  "
	for i, resp := range resps {
		top := t.Topics[i]
		fmt.Println(top.String())
		for _, p := range resp.GetPartitions() {
			fmt.Printf("%sPartition %d\n", indent, p.Partition)
			if p.Head != nil {
				fmt.Printf("%sHead:\n%s\n", indent+indent, eventMetadataString(p.Head, indent+indent+indent))
			} else {
				fmt.Printf("%sNo events published\n", indent+indent)
			}
		}
	}
	return nil
}

type subscriptionCmd struct {
	Info  subscriptionInfoCmd  `cmd:"" help:"Get info about a subscription. This allows you to see how far behind a subscription is from the latest event."`
	Reset resetSubscriptionCmd `cmd:"" help:"Reset the subscription to the head of its topic."`
}

type subscriptionInfoCmd struct {
	Subscriptions []reflection.Ref `arg:"" required:"" help:"Subscriptions to get info about." predictor:"subscriptions"`
}

func (s *subscriptionInfoCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	resps := make([]*adminpb.GetSubscriptionInfoResponse, 0, len(s.Subscriptions))
	for _, sub := range s.Subscriptions {
		resp, err := client.GetSubscriptionInfo(ctx, connect.NewRequest(&adminpb.GetSubscriptionInfoRequest{
			Subscription: sub.ToProto(),
		}))
		if err != nil {
			return fmt.Errorf("failed to get topic info: %w", err)
		}
		resps = append(resps, resp.Msg)
	}
	indent := "  "
	for i, resp := range resps {
		sub := s.Subscriptions[i]
		fmt.Println(sub.String())
		for _, p := range resp.GetPartitions() {
			if p.Head == nil {
				fmt.Printf("%sPartition %d: No events published\n", indent, p.Partition)
				continue
			} else if p.Consumed != nil {
				if p.Consumed.Offset == p.Head.Offset {
					fmt.Printf("%sPartition %d: Idle, waiting for next event to be published\n", indent, p.Partition)
				} else {
					fmt.Printf("%sPartition %d: Trailing head by %d events\n", indent, p.Partition, p.Head.Offset-p.Consumed.Offset)
				}
			} else {
				fmt.Printf("%sPartition %d:\n", indent, p.Partition)
			}
			if p.Consumed != nil {
				fmt.Printf("%sLast Event Consumed:\n%s\n", indent+indent, eventMetadataString(p.Consumed, indent+indent+indent))
			}
			if p.Next != nil {
				fmt.Printf("%sConsuming from:\n%s\n", indent+indent, eventMetadataString(p.Next, indent+indent+indent))
			}
		}
	}
	return nil
}

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
	fmt.Printf("Subscription %s reset\n", s.Subscription)
	return nil
}

func eventMetadataString(event *adminpb.PubSubEventMetadata, indent string) string {
	lines := []string{
		fmt.Sprintf("%sOffset:      %d", indent, event.Offset),
		fmt.Sprintf("%sRequest Key: %v", indent, event.RequestKey),
		fmt.Sprintf("%sPublished:   %v", indent, event.Timestamp.AsTime().Format(time.RFC3339)),
	}
	return strings.Join(lines, "\n")
}
