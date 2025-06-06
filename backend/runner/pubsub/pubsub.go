package pubsub

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	pubsubpb "github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/schema"
	sl "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/timelineclient"
)

type Service struct {
	moduleName string
	publishers map[string]*publisher
	consumers  map[string]*consumer
}

type VerbClient interface {
	Call(ctx context.Context, c *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

var _ pubsubpbconnect.PublishServiceHandler = (*Service)(nil)
var _ pubsubpbconnect.PubSubAdminServiceHandler = (*Service)(nil)

func New(module *schema.Module, deployment key.Deployment, verbClient VerbClient, timelineClient timelineclient.Publisher) (*Service, error) {
	publishers := map[string]*publisher{}
	for t := range sl.FilterVariants[*schema.Topic](module.Decls) {
		publisher, err := newPublisher(module.Name, t, deployment, timelineClient)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		publishers[t.Name] = publisher
	}

	consumers := map[string]*consumer{}
	for v := range sl.FilterVariants[*schema.Verb](module.Decls) {
		subscriber, ok := sl.FindVariant[*schema.MetadataSubscriber](v.Metadata)
		if !ok {
			continue
		}
		var deadLetterPublisher optional.Option[*publisher]
		if subscriber.DeadLetter {
			p, ok := publishers[schema.DeadLetterNameForSubscriber(v.Name)]
			if !ok {
				return nil, errors.Errorf("dead letter publisher not found for subscription %s", v.Name)
			}
			deadLetterPublisher = optional.Some(p)
		}
		consumer, err := newConsumer(module.Name, v, subscriber, deployment, deadLetterPublisher, verbClient, timelineClient)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		consumers[v.Name] = consumer
	}

	return &Service{
		moduleName: module.Name,
		publishers: publishers,
		consumers:  consumers,
	}, nil
}

func (s *Service) Consume(ctx context.Context) error {
	for _, c := range s.consumers {
		err := c.Begin(ctx)
		if err != nil {
			return errors.Wrap(err, "could not begin consumer")
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) PublishEvent(ctx context.Context, req *connect.Request[pubsubpb.PublishEventRequest]) (*connect.Response[pubsubpb.PublishEventResponse], error) {
	publisher, ok := s.publishers[req.Msg.Topic.Name]
	if !ok {
		return nil, errors.Errorf("topic %s not found", req.Msg.Topic.Name)
	}
	caller := schema.Ref{
		Module: s.moduleName,
		Name:   req.Msg.Caller,
	}
	err := publisher.publish(ctx, req.Msg.Body, req.Msg.Key, caller)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&pubsubpb.PublishEventResponse{}), nil
}

func (s *Service) ResetOffsetsOfSubscription(ctx context.Context, req *connect.Request[pubsubpb.ResetOffsetsOfSubscriptionRequest]) (*connect.Response[pubsubpb.ResetOffsetsOfSubscriptionResponse], error) {
	consumer, ok := s.consumers[req.Msg.Subscription.Name]
	if !ok {
		return connect.NewResponse(&pubsubpb.ResetOffsetsOfSubscriptionResponse{}), nil
	}
	partitions, err := consumer.ResetOffsetsForClaimedPartitions(ctx, req.Msg.Offset != adminpb.SubscriptionOffset_SUBSCRIPTION_OFFSET_EARLIEST,
		sl.Map(req.Msg.Partitions, func(i int32) int { return int(i) }))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&pubsubpb.ResetOffsetsOfSubscriptionResponse{
		Partitions: sl.Map(partitions, func(p int) int32 {
			return int32(p)
		}),
	}), nil
}
