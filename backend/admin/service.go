package admin

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	pubsubpb "github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	configuration "github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/oci"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type Config struct {
	ArtefactChunkSize int `help:"Size of each chunk streamed to the client." default:"1048576"`
}

type Service struct {
	env            *EnvironmentManager
	timelineClient *timelineclient.RealClient
	schemaClient   ftlv1connect.SchemaServiceClient
	source         *schemaeventsource.EventSource
	storage        *oci.ArtefactService
	config         Config
	routeTable     *routing.VerbCallRouter
	waitFor        []string
}

var _ adminpbconnect.AdminServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)
var _ rpc.Service = (*Service)(nil)

func NewSchemaRetriever(source *schemaeventsource.EventSource) SchemaClient {
	return &streamSchemaRetriever{
		source: source,
	}
}

type streamSchemaRetriever struct {
	source *schemaeventsource.EventSource
}

func (c *streamSchemaRetriever) GetSchema(ctx context.Context) (*schema.Schema, error) {
	view := c.source.CanonicalView()
	return &schema.Schema{Realms: view.Realms}, nil
}

// NewAdminService creates a new Service.
// bindAllocator is optional and should be set if a local client is to be used that accesses schema from disk using language plugins.
func NewAdminService(
	config Config,
	cm configuration.Provider[configuration.Configuration],
	sm configuration.Provider[configuration.Secrets],
	schr ftlv1connect.SchemaServiceClient,
	source *schemaeventsource.EventSource,
	storage *oci.ArtefactService,
	routes *routing.VerbCallRouter,
	timelineClient *timelineclient.RealClient,
	waitFor []string,
) *Service {
	return &Service{
		config:         config,
		env:            &EnvironmentManager{schr: NewSchemaRetriever(source), cm: cm, sm: sm},
		schemaClient:   schr,
		source:         source,
		storage:        storage,
		timelineClient: timelineClient,
		routeTable:     routes,
		waitFor:        waitFor,
	}
}

func (s *Service) StartServices(context.Context) ([]rpc.Option, error) {
	return []rpc.Option{rpc.GRPC(adminpbconnect.NewAdminServiceHandler, s),
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s)}, nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	if len(s.waitFor) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	// It's not actually ready until it is in the routes table
	var missing []string
	for _, module := range s.waitFor {
		if _, _, ok := s.routeTable.LookupClient(ctx, module); !ok {
			missing = append(missing, module)
		}
	}
	if len(missing) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	msg := fmt.Sprintf("waiting for deployments: %s", strings.Join(missing, ", "))
	return connect.NewResponse(&ftlv1.PingResponse{NotReady: &msg}), nil
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *Service) ConfigList(ctx context.Context, req *connect.Request[adminpb.ConfigListRequest]) (*connect.Response[adminpb.ConfigListResponse], error) {
	return errors.WithStack2(s.env.ConfigList(ctx, req))
}

// ConfigGet returns the configuration value for a given ref string.
func (s *Service) ConfigGet(ctx context.Context, req *connect.Request[adminpb.ConfigGetRequest]) (*connect.Response[adminpb.ConfigGetResponse], error) {
	return errors.WithStack2(s.env.ConfigGet(ctx, req))
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *Service) ConfigSet(ctx context.Context, req *connect.Request[adminpb.ConfigSetRequest]) (*connect.Response[adminpb.ConfigSetResponse], error) {
	return errors.WithStack2(s.env.ConfigSet(ctx, req))
}

// ConfigUnset unsets the config value at the given ref.
func (s *Service) ConfigUnset(ctx context.Context, req *connect.Request[adminpb.ConfigUnsetRequest]) (*connect.Response[adminpb.ConfigUnsetResponse], error) {
	return errors.WithStack2(s.env.ConfigUnset(ctx, req))
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *Service) SecretsList(ctx context.Context, req *connect.Request[adminpb.SecretsListRequest]) (*connect.Response[adminpb.SecretsListResponse], error) {
	return errors.WithStack2(s.env.SecretsList(ctx, req))
}

// SecretGet returns the secret value for a given ref string.
func (s *Service) SecretGet(ctx context.Context, req *connect.Request[adminpb.SecretGetRequest]) (*connect.Response[adminpb.SecretGetResponse], error) {
	return errors.WithStack2(s.env.SecretGet(ctx, req))
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *Service) SecretSet(ctx context.Context, req *connect.Request[adminpb.SecretSetRequest]) (*connect.Response[adminpb.SecretSetResponse], error) {
	return errors.WithStack2(s.env.SecretSet(ctx, req))
}

// SecretUnset unsets the secret value at the given ref.
func (s *Service) SecretUnset(ctx context.Context, req *connect.Request[adminpb.SecretUnsetRequest]) (*connect.Response[adminpb.SecretUnsetResponse], error) {
	return errors.WithStack2(s.env.SecretUnset(ctx, req))
}

// MapConfigsForModule combines all configuration values visible to the module.
func (s *Service) MapConfigsForModule(ctx context.Context, req *connect.Request[adminpb.MapConfigsForModuleRequest]) (*connect.Response[adminpb.MapConfigsForModuleResponse], error) {
	return errors.WithStack2(s.env.MapConfigsForModule(ctx, req))
}

// MapSecretsForModule combines all secrets visible to the module.
func (s *Service) MapSecretsForModule(ctx context.Context, req *connect.Request[adminpb.MapSecretsForModuleRequest]) (*connect.Response[adminpb.MapSecretsForModuleResponse], error) {
	return errors.WithStack2(s.env.MapSecretsForModule(ctx, req))
}

func (s *Service) ResetSubscription(ctx context.Context, req *connect.Request[adminpb.ResetSubscriptionRequest]) (*connect.Response[adminpb.ResetSubscriptionResponse], error) {
	// Find nodes in schema
	// TODO: we really want all deployments for a module... not just latest... Use canonical and check ActiveChangeset?
	sch := s.source.CanonicalView()
	module, ok := islices.Find(sch.InternalModules(), func(m *schema.Module) bool {
		return m.Name == req.Msg.Subscription.Module
	})
	if !ok {
		return nil, errors.Errorf("module %q not found", req.Msg.Subscription.Module)
	}
	verb, ok := islices.Find(slices.Collect(islices.FilterVariants[*schema.Verb](module.Decls)), func(v *schema.Verb) bool {
		return v.Name == req.Msg.Subscription.Name
	})
	if !ok {
		return nil, errors.Errorf("verb %q not found in module %q", req.Msg.Subscription.Name, req.Msg.Subscription.Module)
	}
	subscriber, ok := islices.FindVariant[*schema.MetadataSubscriber](verb.Metadata)
	if !ok {
		return nil, errors.Errorf("%q is not a subscriber", req.Msg.Subscription)
	}
	connection, ok := verb.Runtime.SubscriptionConnector.(*schema.PlaintextKafkaSubscriptionConnector)
	if !ok {
		return nil, errors.Errorf("only plaintext kafka subscription connector is supported, got %T", verb.Runtime.SubscriptionConnector)
	}
	if len(connection.KafkaBrokers) == 0 {
		return nil, errors.Errorf("no Kafka brokers for subscription %q", req.Msg.Subscription)
	}
	if module.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
		return nil, errors.Errorf("no deployment for module %s", req.Msg.Subscription.Module)
	}
	topicID := subscriber.Topic.String()
	totalPartitions, err := kafkaPartitionCount(optional.None[sarama.ClusterAdmin](), connection.KafkaBrokers, topicID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get partition count for topic %s", topicID)
	}

	client := rpc.Dial(pubsubpbconnect.NewPubSubAdminServiceClient, module.Runtime.Runner.Endpoint, log.Error)
	remainingPartitions := make([]int32, totalPartitions)
	for i := range totalPartitions {
		remainingPartitions[i] = i
	}
	// Allowing multiple attempts to reset offsets in case runners are currently rebalancing partition claims
	for attempt := 0; attempt < 10 && len(remainingPartitions) > 0; attempt++ {
		if attempt != 0 {
			time.Sleep(time.Millisecond * 500)
		}
		resp, err := client.ResetOffsetsOfSubscription(ctx, connect.NewRequest(&pubsubpb.ResetOffsetsOfSubscriptionRequest{
			Subscription: req.Msg.Subscription,
			Offset:       req.Msg.Offset,
			Partitions:   remainingPartitions,
		}))
		if err != nil {
			return nil, errors.Wrap(err, "failed to reset subscription")
		}
		remainingPartitions = islices.Filter(remainingPartitions, func(p int32) bool {
			return !slices.Contains(resp.Msg.Partitions, p)
		})
	}
	if len(remainingPartitions) > 0 {
		return nil, errors.Errorf("failed to reset partitions %v: no runner had partition claim", remainingPartitions)
	}
	return connect.NewResponse(&adminpb.ResetSubscriptionResponse{}), nil
}

func (s *Service) GetTopicInfo(ctx context.Context, req *connect.Request[adminpb.GetTopicInfoRequest]) (*connect.Response[adminpb.GetTopicInfoResponse], error) {
	ref, err := schema.RefFromProto(req.Msg.Topic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse topic")
	}
	sch := s.source.ViewOnly().GetCanonical()
	t, ok := sch.Resolve(ref).Get()
	if !ok {
		return nil, errors.Errorf("failed to resolve topic %s", ref)
	}
	topic, ok := t.(*schema.Topic)
	if !ok {
		return nil, errors.Errorf("expected topic instead of %T", t)
	}
	if topic.Runtime == nil {
		return nil, errors.Errorf("topic %s has no runtime", ref)
	}
	config := sarama.NewConfig()
	client, err := sarama.NewClient(topic.Runtime.KafkaBrokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kakfa client")
	}
	defer client.Close()
	partitionCount, err := kafkaPartitionCount(optional.None[sarama.ClusterAdmin](), topic.Runtime.KafkaBrokers, topic.Runtime.TopicID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get partition count")
	}
	partitions := make(chan *adminpb.GetTopicInfoResponse_PartitionInfo, partitionCount)
	wg := &errgroup.Group{}
	for i := range partitionCount {
		wg.Go(func() error {
			oldestOffset, err := client.GetOffset(topic.Runtime.TopicID, i, sarama.OffsetOldest)
			if err != nil {
				return errors.Wrapf(err, "failed to get offset for partition %d", i)
			}
			offset, err := client.GetOffset(topic.Runtime.TopicID, i, sarama.OffsetNewest)
			if err != nil {
				return errors.Wrapf(err, "failed to get offset for partition %d", i)
			}
			info := &adminpb.GetTopicInfoResponse_PartitionInfo{
				Partition: i,
			}
			if oldestOffset == offset {
				// no messages
				partitions <- info
				return nil
			}

			headOffset := offset - 1
			info.Head, err = getEventMetadata(ctx, topic.Runtime.KafkaBrokers, topic.Runtime.TopicID, i, headOffset)
			if err != nil {
				return errors.Wrapf(err, "failed to get event info for partition %d and offset %d", i, headOffset)
			}
			partitions <- info
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, errors.Wrap(err, "failed to get partition info")
	}
	close(partitions)
	outPartitions := make([]*adminpb.GetTopicInfoResponse_PartitionInfo, partitionCount)
	for p := range channels.IterContext(ctx, partitions) {
		outPartitions[p.Partition] = p
	}
	return connect.NewResponse(&adminpb.GetTopicInfoResponse{
		Partitions: outPartitions,
	}), nil
}

func (s *Service) GetSubscriptionInfo(ctx context.Context, req *connect.Request[adminpb.GetSubscriptionInfoRequest]) (*connect.Response[adminpb.GetSubscriptionInfoResponse], error) {
	ref, err := schema.RefFromProto(req.Msg.Subscription)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse topic")
	}
	sch := s.source.ViewOnly().GetCanonical()
	v, ok := sch.Resolve(ref).Get()
	if !ok {
		return nil, errors.Errorf("failed to resolve topic %s", ref)
	}
	verb, ok := v.(*schema.Verb)
	if !ok {
		return nil, errors.Errorf("expected subscription instead of %T", v)
	}
	subscription, ok := islices.FindVariant[*schema.MetadataSubscriber](verb.Metadata)
	if !ok {
		return nil, errors.Errorf("verb %s is not a subscriber", ref)
	}
	if verb.Runtime == nil || verb.Runtime.SubscriptionConnector == nil {
		return nil, errors.Errorf("verb %s has no runtime info for subscription", ref)
	}
	connector, ok := verb.Runtime.SubscriptionConnector.(*schema.PlaintextKafkaSubscriptionConnector)
	if !ok {
		return nil, errors.Errorf("unsupported subscription connector %T", verb.Runtime.SubscriptionConnector)
	}

	config := sarama.NewConfig()
	client, err := sarama.NewClient(connector.KafkaBrokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kakfa client")
	}
	defer client.Close()

	admin, err := sarama.NewClusterAdmin(connector.KafkaBrokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kafka admin client")
	}
	partitionCount, err := kafkaPartitionCount(optional.Some(admin), connector.KafkaBrokers, subscription.Topic.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get partition count")
	}
	partitionIDs := make([]int32, partitionCount)
	for i := range partitionCount {
		partitionIDs[i] = i
	}
	offsetResp, err := admin.ListConsumerGroupOffsets(ref.String(), map[string][]int32{subscription.Topic.String(): partitionIDs})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get consumer group offsets")
	}
	consumerOffsets, ok := offsetResp.Blocks[subscription.Topic.String()]
	if !ok {
		return nil, errors.Errorf("no consumer offsets found for topic %s", subscription.Topic.String())
	}

	partitions := make(chan *adminpb.GetSubscriptionInfoResponse_PartitionInfo, partitionCount)
	wg := &errgroup.Group{}
	for i := range partitionCount {
		wg.Go(func() error {
			oldestOffset, err := client.GetOffset(subscription.Topic.String(), i, sarama.OffsetOldest)
			if err != nil {
				return errors.Wrapf(err, "failed to get offset for partition %d", i)
			}
			highwaterMark, err := client.GetOffset(subscription.Topic.String(), i, sarama.OffsetNewest)
			if err != nil {
				return errors.Wrapf(err, "failed to get offset for partition %d", i)
			}
			info := &adminpb.GetSubscriptionInfoResponse_PartitionInfo{
				Partition: i,
			}
			if oldestOffset == highwaterMark {
				// no messages
				partitions <- info
				return nil
			}

			headOffset := highwaterMark - 1
			info.Head, err = getEventMetadata(ctx, connector.KafkaBrokers, subscription.Topic.String(), i, headOffset)
			if err != nil {
				return errors.Wrapf(err, "failed to get event info for partition %d and offset %d", i, headOffset)
			}

			offsetBlock, ok := consumerOffsets[i]
			if !ok {
				// no messages consumed
				partitions <- info
				return nil
			}
			if offsetBlock.Offset-1 >= oldestOffset {
				info.Consumed, err = getEventMetadata(ctx, connector.KafkaBrokers, subscription.Topic.String(), i, offsetBlock.Offset-1)
				if err != nil {
					return errors.Wrapf(err, "failed to get event info for partition %d and offset %d", i, offsetBlock.Offset-1)
				}
			}
			if offsetBlock.Offset <= headOffset && offsetBlock.Offset >= oldestOffset {
				info.Next, err = getEventMetadata(ctx, connector.KafkaBrokers, subscription.Topic.String(), i, offsetBlock.Offset)
				if err != nil {
					return errors.Wrapf(err, "failed to get event info for partition %d and offset %d", i, offsetBlock.Offset)
				}
			}

			partitions <- info
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, errors.Wrap(err, "failed to get partition info")
	}
	close(partitions)
	outPartitions := make([]*adminpb.GetSubscriptionInfoResponse_PartitionInfo, partitionCount)
	for p := range channels.IterContext(ctx, partitions) {
		outPartitions[p.Partition] = p
	}
	return connect.NewResponse(&adminpb.GetSubscriptionInfoResponse{
		Partitions: outPartitions,
	}), nil
}

// kafkaPartitionCount returns the number of partitions for a given topic in kafka. This may differ from the number
// of partitions in the schema if the topic was originally provisioned with a different number of partitions.
func kafkaPartitionCount(adminClient optional.Option[sarama.ClusterAdmin], brokers []string, topicID string) (int32, error) {
	config := sarama.NewConfig()
	admin, ok := adminClient.Get()
	if !ok {
		var err error
		admin, err = sarama.NewClusterAdmin(brokers, config)
		if err != nil {
			return 0, errors.Wrap(err, "failed to create kafka admin client")
		}
		defer admin.Close()
	}

	topicMetas, err := admin.DescribeTopics([]string{topicID})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to describe topic %s", topicID)
	}
	if len(topicMetas) != 1 {
		return 0, errors.Errorf("expected topic metadata for %s from kafka but received none", topicID)
	}
	if topicMetas[0].Err == sarama.ErrUnknownTopicOrPartition {
		return 0, errors.Wrapf(err, "can not reset subscription for topic %s that does not exist yet", topicID)
	} else if topicMetas[0].Err != sarama.ErrNoError {
		return 0, errors.Wrapf(topicMetas[0].Err, "failed to describe topic %s", topicID)
	}
	return int32(len(topicMetas[0].Partitions)), nil //nolint:gosec
}

func getEventMetadata(ctx context.Context, brokers []string, topicID string, partition int32, offset int64) (*adminpb.PubSubEventMetadata, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kafka consumer")
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topicID, partition, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create partition consumer")
	}
	defer partitionConsumer.Close()

	select {
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	case <-time.After(5 * time.Second):
		return nil, errors.New("timeout waiting for message")
	case err := <-partitionConsumer.Errors():
		return nil, errors.Wrap(err, "failed to consume partition")
	case message := <-partitionConsumer.Messages():
		var requestKey key.Request
		found := false
		for _, header := range message.Headers {
			if string(header.Key) != "ftl.request_key" {
				continue
			}
			found = true
			requestKey, err = key.ParseRequestKey(string(header.Value))
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse request key")
			}
		}
		if !found {
			return nil, errors.WithStack(errors.New("request key not found"))
		}

		return &adminpb.PubSubEventMetadata{
			Offset:     offset,
			RequestKey: requestKey.String(),
			Timestamp:  timestamppb.New(message.Timestamp),
		}, nil

	}
}
func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	sch, err := s.schemaClient.GetSchema(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest schema")
	}
	return connect.NewResponse(sch.Msg), nil
}

func (s *Service) ApplyChangeset(ctx context.Context, req *connect.Request[adminpb.ApplyChangesetRequest], stream *connect.ServerStream[adminpb.ApplyChangesetResponse]) error {
	var changes []*ftlv1.RealmChange
	var pbchanges []*schemapb.RealmChange
	for _, realmChange := range req.Msg.RealmChanges {
		changes = append(changes, &ftlv1.RealmChange{
			Name:     realmChange.Name,
			Modules:  realmChange.Modules,
			ToRemove: realmChange.ToRemove,
			External: realmChange.External,
		})
		pbchanges = append(pbchanges, &schemapb.RealmChange{
			Name:     realmChange.Name,
			Modules:  realmChange.Modules,
			ToRemove: realmChange.ToRemove,
			External: realmChange.External,
		})
	}
	cs, err := s.schemaClient.CreateChangeset(ctx, connect.NewRequest(&ftlv1.CreateChangesetRequest{
		RealmChanges: changes,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to create changeset")
	}
	key, err := key.ParseChangesetKey(cs.Msg.Changeset)
	if err != nil {
		return errors.Wrap(err, "failed to parse changeset key")
	}
	changeset := &schemapb.Changeset{
		Key:          cs.Msg.Changeset,
		RealmChanges: pbchanges,
	}
	if err := stream.Send(&adminpb.ApplyChangesetResponse{
		Changeset: changeset,
	}); err != nil {
		return errors.Wrap(err, "failed to send changeset")
	}
	for e := range channels.IterContext(ctx, s.source.Subscribe(ctx)) {
		switch event := e.(type) {
		case *schema.ChangesetFinalizedNotification:
			if event.Key != key {
				continue
			}
			if err := stream.Send(&adminpb.ApplyChangesetResponse{
				Changeset: changeset,
			}); err != nil {
				return errors.Wrap(err, "failed to send changeset")
			}
			return nil
		case *schema.ChangesetFailedNotification:
			if event.Key != key {
				continue
			}
			return errors.Errorf("failed to apply changeset: %s", event.Error)
		case *schema.ChangesetCommittedNotification:
			if event.Changeset.Key != key {
				continue
			}
			changeset = event.Changeset.ToProto()
			// We don't wait for cleanup, just return immediately
			if err := stream.Send(&adminpb.ApplyChangesetResponse{
				Changeset: changeset,
			}); err != nil {
				return errors.Wrap(err, "failed to send changeset")
			}
			return nil
		case *schema.ChangesetRollingBackNotification:
			if event.Changeset.Key != key {
				continue
			}
			changeset = event.Changeset.ToProto()
		default:

		}
	}
	return errors.Errorf("failed to apply changeset: context cancelled")
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], resp *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	events := s.source.Subscribe(ctx)
	for event := range channels.IterContext(ctx, events) {
		switch e := event.(type) {
		case *schema.FullSchemaNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_FullSchemaNotification{FullSchemaNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.DeploymentRuntimeNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_DeploymentRuntimeNotification{DeploymentRuntimeNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetCreatedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetCreatedNotification{ChangesetCreatedNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetPreparedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetPreparedNotification{ChangesetPreparedNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetCommittedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetCommittedNotification{ChangesetCommittedNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetDrainedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetDrainedNotification{ChangesetDrainedNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetFinalizedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetFinalizedNotification{ChangesetFinalizedNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetRollingBackNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetRollingBackNotification{ChangesetRollingBackNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		case *schema.ChangesetFailedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetFailedNotification{ChangesetFailedNotification: proto},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send")
			}
		}

	}
	return errors.Wrap(ctx.Err(), "context cancelled")
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[adminpb.GetArtefactDiffsRequest]) (*connect.Response[adminpb.GetArtefactDiffsResponse], error) {
	byteDigests, err := islices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse digests")
	}
	_, need, err := s.storage.GetDigestsKeys(ctx, byteDigests)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get digests")
	}
	return connect.NewResponse(&adminpb.GetArtefactDiffsResponse{
		MissingDigests: islices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, stream *connect.ClientStream[adminpb.UploadArtefactRequest]) (*connect.Response[adminpb.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx).Scope("uploadArtefact")
	firstMsg := NewOnceValue[*adminpb.UploadArtefactRequest]()
	wg, ctx := errgroup.WithContext(ctx)
	logger.Debugf("Uploading artefact")
	r, w := io.Pipe()
	// Read bytes from client and upload to OCI
	wg.Go(func() error {
		defer r.Close()
		logger.Tracef("Waiting for first message")
		msg, ok := firstMsg.Get(ctx)
		if !ok {
			return nil
		}
		if msg.Size == 0 {
			return errors.Errorf("artefact size must be specified")
		}
		digest, err := sha256.ParseSHA256(hex.EncodeToString(msg.Digest))
		if err != nil {
			return errors.Wrap(err, "failed to parse digest")
		}
		logger = logger.Scope("uploadArtefact:" + digest.String())
		logger.Debugf("Starting upload to OCI")
		err = s.storage.Upload(ctx, oci.ArtefactUpload{
			Digest:  digest,
			Size:    msg.Size,
			Content: r,
		})
		if err != nil {
			return errors.Wrap(err, "failed to upload artefact")
		}
		logger.Debugf("Created new artefact %s", digest)
		return nil
	})
	// Stream bytes from client into the pipe
	wg.Go(func() error {
		defer w.Close()
		logger.Debugf("Starting forwarder from client to OCI")
		for stream.Receive() {
			msg := stream.Msg()
			if len(msg.Chunk) == 0 {
				return errors.Errorf("zero length chunk received")
			}
			firstMsg.Set(msg)
			if _, err := w.Write(msg.Chunk); err != nil {
				return errors.Wrap(err, "failed to write chunk")
			}
		}
		if err := stream.Err(); err != nil {
			return errors.Wrap(err, "failed to upload artefact")
		}
		return nil
	})
	err := wg.Wait()
	if err != nil {
		return nil, errors.Wrap(err, "failed to upload artefact")
	}
	return connect.NewResponse(&adminpb.UploadArtefactResponse{}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[adminpb.GetDeploymentArtefactsRequest], resp *connect.ServerStream[adminpb.GetDeploymentArtefactsResponse]) error {
	dkey, err := key.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key")))
	}

	deploymentOpt := s.source.CanonicalView().Deployment(dkey)
	deployment, ok := deploymentOpt.Get()
	if !ok {
		return errors.Errorf("could not get deployment: %s", req.Msg.DeploymentKey)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Get deployment artefacts for: %s", dkey.String())

	chunk := make([]byte, s.config.ArtefactChunkSize)
nextArtefact:
	for artefact := range islices.FilterVariants[schema.MetadataArtefact](deployment.Metadata) {
		deploymentArtefact := &adminpb.DeploymentArtefact{
			Digest:     artefact.Digest[:],
			Path:       artefact.Path,
			Executable: artefact.Executable,
		}
		for _, clientArtefact := range req.Msg.HaveArtefacts {
			if proto.Equal(deploymentArtefact, clientArtefact) {
				continue nextArtefact
			}
		}
		reader, err := s.storage.Download(ctx, artefact.Digest)
		if err != nil {
			return errors.Wrap(err, "could not download artefact")
		}
		defer reader.Close()
		for {

			n, err := reader.Read(chunk)
			if n != 0 {
				if err := resp.Send(&adminpb.GetDeploymentArtefactsResponse{
					Artefact: deploymentArtefact,
					Chunk:    chunk[:n],
				}); err != nil {
					return errors.Wrap(err, "could not send artefact chunk")
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return errors.Wrap(err, "could not read artefact chunk")
			}
		}
	}
	return nil
}

func (s *Service) ClusterInfo(ctx context.Context, req *connect.Request[adminpb.ClusterInfoRequest]) (*connect.Response[adminpb.ClusterInfoResponse], error) {
	return connect.NewResponse(&adminpb.ClusterInfoResponse{Os: runtime.GOOS, Arch: runtime.GOARCH}), nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	ref, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse verb")
	}

	if err := schema.ValidateJSONCall(req.Msg.Body, ref, s.source.CanonicalView()); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	call, err := s.routeTable.Call(ctx, headers.CopyRequestForForwarding(req))
	if err != nil {
		return nil, errors.Wrap(err, "failed to call verb")
	}
	return call, nil
}

func (s *Service) RollbackChangeset(ctx context.Context, c *connect.Request[ftlv1.RollbackChangesetRequest]) (*connect.Response[ftlv1.RollbackChangesetResponse], error) {
	res, err := s.schemaClient.RollbackChangeset(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, errors.Wrap(err, "failed to rollback changeset")
	}
	return connect.NewResponse(res.Msg), nil
}

func (s *Service) FailChangeset(ctx context.Context, c *connect.Request[ftlv1.FailChangesetRequest]) (*connect.Response[ftlv1.FailChangesetResponse], error) {
	res, err := s.schemaClient.FailChangeset(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, errors.Wrap(err, "failed to fail changeset")
	}
	return connect.NewResponse(res.Msg), nil
}

func (s *Service) StreamLogs(ctx context.Context, req *connect.Request[adminpb.StreamLogsRequest], resp *connect.ServerStream[adminpb.StreamLogsResponse]) error {
	query := req.Msg.Query
	query.Filters = append(query.Filters, &timelinepb.TimelineQuery_Filter{
		Filter: &timelinepb.TimelineQuery_Filter_EventTypes{
			EventTypes: &timelinepb.TimelineQuery_EventTypeFilter{
				EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_LOG},
			},
		},
	})

	timeline, err := s.timelineClient.StreamTimeline(ctx, connect.NewRequest(&timelinepb.StreamTimelineRequest{
		Query: query,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to get timeline")
	}

	for timeline.Receive() {
		msg := timeline.Msg()
		logs := []*timelinepb.LogEvent{}
		for _, event := range msg.Events {
			if log := event.GetLog(); log != nil {
				logs = append(logs, log)
			}
		}
		if err := resp.Send(&adminpb.StreamLogsResponse{
			Logs: logs,
		}); err != nil {
			return errors.Wrap(err, "failed to send logs")
		}
	}
	return nil
}

type OnceValue[T any] struct {
	value T
	ready chan struct{}
	once  sync.Once
}

func NewOnceValue[T any]() *OnceValue[T] {
	return &OnceValue[T]{
		ready: make(chan struct{}),
	}
}

func (o *OnceValue[T]) Set(value T) {
	o.once.Do(func() {
		o.value = value
		close(o.ready)
	})
}

func (o *OnceValue[T]) Get(ctx context.Context) (T, bool) {
	select {
	case <-o.ready:
		return o.value, true
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}

func (s *Service) UpdateDeploymentRuntime(ctx context.Context, c *connect.Request[adminpb.UpdateDeploymentRuntimeRequest]) (*connect.Response[adminpb.UpdateDeploymentRuntimeResponse], error) {
	_, err := s.schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{Update: c.Msg.Element}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to update deployment runtime")
	}
	return connect.NewResponse(&adminpb.UpdateDeploymentRuntimeResponse{}), nil
}
