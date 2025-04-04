package admin

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/controller/state"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	pubsubpb "github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type Config struct {
	Bind              *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	ArtefactChunkSize int      `help:"Size of each chunk streamed to the client." default:"1048576"`
}

type Service struct {
	env            *EnvironmentManager
	timelineClient *timelineclient.Client
	schemaClient   ftlv1connect.SchemaServiceClient
	source         *schemaeventsource.EventSource
	storage        *artefacts.OCIArtefactService
	config         Config
	routeTable     *routing.VerbCallRouter
	waitFor        []string
}

var _ adminpbconnect.AdminServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

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
	env *EnvironmentManager,
	schr ftlv1connect.SchemaServiceClient,
	source *schemaeventsource.EventSource,
	storage *artefacts.OCIArtefactService,
	routes *routing.VerbCallRouter,
	timelineClient *timelineclient.Client,
	waitFor []string,
) *Service {
	return &Service{
		config:         config,
		env:            env,
		schemaClient:   schr,
		source:         source,
		storage:        storage,
		timelineClient: timelineClient,
		routeTable:     routes,
		waitFor:        waitFor,
	}
}

func Start(
	ctx context.Context,
	config Config,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	schr ftlv1connect.SchemaServiceClient,
	source *schemaeventsource.EventSource,
	timelineClient *timelineclient.Client,
	storage *artefacts.OCIArtefactService,
	waitFor []string) error {

	logger := log.FromContext(ctx).Scope("admin")

	svc := NewAdminService(
		config,
		&EnvironmentManager{schr: NewSchemaRetriever(source), cm: cm, sm: sm},
		schr,
		source,
		storage,
		routing.NewVerbRouter(ctx, source, timelineClient),
		timelineClient,
		waitFor,
	)

	logger.Debugf("Admin service listening on: %s", config.Bind)
	err := rpc.Serve(ctx, config.Bind,
		rpc.GRPC(adminpbconnect.NewAdminServiceHandler, svc),
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
	)
	if err != nil {
		return fmt.Errorf("admin service stopped serving: %w", err)
	}
	return nil
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
	return s.env.ConfigList(ctx, req)
}

// ConfigGet returns the configuration value for a given ref string.
func (s *Service) ConfigGet(ctx context.Context, req *connect.Request[adminpb.ConfigGetRequest]) (*connect.Response[adminpb.ConfigGetResponse], error) {
	return s.env.ConfigGet(ctx, req)
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *Service) ConfigSet(ctx context.Context, req *connect.Request[adminpb.ConfigSetRequest]) (*connect.Response[adminpb.ConfigSetResponse], error) {
	return s.env.ConfigSet(ctx, req)
}

// ConfigUnset unsets the config value at the given ref.
func (s *Service) ConfigUnset(ctx context.Context, req *connect.Request[adminpb.ConfigUnsetRequest]) (*connect.Response[adminpb.ConfigUnsetResponse], error) {
	return s.env.ConfigUnset(ctx, req)
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *Service) SecretsList(ctx context.Context, req *connect.Request[adminpb.SecretsListRequest]) (*connect.Response[adminpb.SecretsListResponse], error) {
	return s.env.SecretsList(ctx, req)
}

// SecretGet returns the secret value for a given ref string.
func (s *Service) SecretGet(ctx context.Context, req *connect.Request[adminpb.SecretGetRequest]) (*connect.Response[adminpb.SecretGetResponse], error) {
	return s.env.SecretGet(ctx, req)
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *Service) SecretSet(ctx context.Context, req *connect.Request[adminpb.SecretSetRequest]) (*connect.Response[adminpb.SecretSetResponse], error) {
	return s.env.SecretSet(ctx, req)
}

// SecretUnset unsets the secret value at the given ref.
func (s *Service) SecretUnset(ctx context.Context, req *connect.Request[adminpb.SecretUnsetRequest]) (*connect.Response[adminpb.SecretUnsetResponse], error) {
	return s.env.SecretUnset(ctx, req)
}

// MapConfigsForModule combines all configuration values visible to the module.
func (s *Service) MapConfigsForModule(ctx context.Context, req *connect.Request[adminpb.MapConfigsForModuleRequest]) (*connect.Response[adminpb.MapConfigsForModuleResponse], error) {
	return s.env.MapConfigsForModule(ctx, req)
}

// MapSecretsForModule combines all secrets visible to the module.
func (s *Service) MapSecretsForModule(ctx context.Context, req *connect.Request[adminpb.MapSecretsForModuleRequest]) (*connect.Response[adminpb.MapSecretsForModuleResponse], error) {
	return s.env.MapSecretsForModule(ctx, req)
}

func (s *Service) ResetSubscription(ctx context.Context, req *connect.Request[adminpb.ResetSubscriptionRequest]) (*connect.Response[adminpb.ResetSubscriptionResponse], error) {
	// Find nodes in schema
	// TODO: we really want all deployments for a module... not just latest... Use canonical and check ActiveChangeset?
	sch := s.source.CanonicalView()
	module, ok := islices.Find(sch.InternalModules(), func(m *schema.Module) bool {
		return m.Name == req.Msg.Subscription.Module
	})
	if !ok {
		return nil, fmt.Errorf("module %q not found", req.Msg.Subscription.Module)
	}
	verb, ok := islices.Find(slices.Collect(islices.FilterVariants[*schema.Verb](module.Decls)), func(v *schema.Verb) bool {
		return v.Name == req.Msg.Subscription.Name
	})
	if !ok {
		return nil, fmt.Errorf("verb %q not found in module %q", req.Msg.Subscription.Name, req.Msg.Subscription.Module)
	}
	subscriber, ok := islices.FindVariant[*schema.MetadataSubscriber](verb.Metadata)
	if !ok {
		return nil, fmt.Errorf("%q is not a subscriber", req.Msg.Subscription)
	}
	connection, ok := verb.Runtime.SubscriptionConnector.(*schema.PlaintextKafkaSubscriptionConnector)
	if !ok {
		return nil, fmt.Errorf("only plaintext kafka subscription connector is supported, got %T", verb.Runtime.SubscriptionConnector)
	}
	if len(connection.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("no Kafka brokers for subscription %q", req.Msg.Subscription)
	}
	if module.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
		return nil, fmt.Errorf("no deployment for module %s", req.Msg.Subscription.Module)
	}
	topicID := subscriber.Topic.String()
	totalPartitions, err := kafkaPartitionCount(optional.None[sarama.ClusterAdmin](), connection.KafkaBrokers, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partition count for topic %s: %w", topicID, err)
	}

	client := rpc.Dial(pubsubpbconnect.NewPubSubAdminServiceClient, module.Runtime.Runner.Endpoint, log.Error)
	resp, err := client.ResetOffsetsOfSubscription(ctx, connect.NewRequest(&pubsubpb.ResetOffsetsOfSubscriptionRequest{
		Subscription: req.Msg.Subscription,
		Offset:       req.Msg.Offset,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to reset subscription: %w", err)
	}

	var successfulPartitions = resp.Msg.Partitions
	slices.Sort(successfulPartitions)
	var failedPartitions = []int32{}
	for p := range totalPartitions {
		if int(p) >= len(successfulPartitions) || successfulPartitions[p] != p {
			failedPartitions = append(failedPartitions, p)
		}
	}
	if len(failedPartitions) > 0 {
		return nil, fmt.Errorf("failed to reset partitions %v: no runner had partition claim", failedPartitions)
	}
	return connect.NewResponse(&adminpb.ResetSubscriptionResponse{}), nil
}

func (s *Service) GetTopicInfo(ctx context.Context, req *connect.Request[adminpb.GetTopicInfoRequest]) (*connect.Response[adminpb.GetTopicInfoResponse], error) {
	ref, err := schema.RefFromProto(req.Msg.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse topic: %w", err)
	}
	sch := s.source.ViewOnly().GetCanonical()
	t, ok := sch.Resolve(ref).Get()
	if !ok {
		return nil, fmt.Errorf("failed to resolve topic %s", ref)
	}
	topic, ok := t.(*schema.Topic)
	if !ok {
		return nil, fmt.Errorf("expected topic instead of %T", t)
	}
	if topic.Runtime == nil {
		return nil, fmt.Errorf("topic %s has no runtime", ref)
	}
	config := sarama.NewConfig()
	client, err := sarama.NewClient(topic.Runtime.KafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kakfa client: %w", err)
	}
	defer client.Close()
	partitionCount, err := kafkaPartitionCount(optional.None[sarama.ClusterAdmin](), topic.Runtime.KafkaBrokers, topic.Runtime.TopicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partition count: %w", err)
	}
	partitions := make(chan *adminpb.GetTopicInfoResponse_PartitionInfo, partitionCount)
	wg := &errgroup.Group{}
	for i := range partitionCount {
		wg.Go(func() error {
			oldestOffset, err := client.GetOffset(topic.Runtime.TopicID, i, sarama.OffsetOldest)
			if err != nil {
				return fmt.Errorf("failed to get offset for partition %d: %w", i, err)
			}
			offset, err := client.GetOffset(topic.Runtime.TopicID, i, sarama.OffsetNewest)
			if err != nil {
				return fmt.Errorf("failed to get offset for partition %d: %w", i, err)
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
				return fmt.Errorf("failed to get event info for partition %d and offset %d: %w", i, headOffset, err)
			}
			partitions <- info
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get partition info: %w", err)
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
		return nil, fmt.Errorf("failed to parse topic: %w", err)
	}
	sch := s.source.ViewOnly().GetCanonical()
	v, ok := sch.Resolve(ref).Get()
	if !ok {
		return nil, fmt.Errorf("failed to resolve topic %s", ref)
	}
	verb, ok := v.(*schema.Verb)
	if !ok {
		return nil, fmt.Errorf("expected subscription instead of %T", v)
	}
	subscription, ok := islices.FindVariant[*schema.MetadataSubscriber](verb.Metadata)
	if !ok {
		return nil, fmt.Errorf("verb %s is not a subscriber", ref)
	}
	if verb.Runtime == nil || verb.Runtime.SubscriptionConnector == nil {
		return nil, fmt.Errorf("verb %s has no runtime info for subscription", ref)
	}
	connector, ok := verb.Runtime.SubscriptionConnector.(*schema.PlaintextKafkaSubscriptionConnector)
	if !ok {
		return nil, fmt.Errorf("unsupported subscription connector %T", verb.Runtime.SubscriptionConnector)
	}

	config := sarama.NewConfig()
	client, err := sarama.NewClient(connector.KafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kakfa client: %w", err)
	}
	defer client.Close()

	admin, err := sarama.NewClusterAdmin(connector.KafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka admin client: %w", err)
	}
	partitionCount, err := kafkaPartitionCount(optional.Some(admin), connector.KafkaBrokers, subscription.Topic.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get partition count: %w", err)
	}
	partitionIDs := make([]int32, partitionCount)
	for i := range partitionCount {
		partitionIDs[i] = i
	}
	offsetResp, err := admin.ListConsumerGroupOffsets(ref.String(), map[string][]int32{subscription.Topic.String(): partitionIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer group offsets: %w", err)
	}
	consumerOffsets, ok := offsetResp.Blocks[subscription.Topic.String()]
	if !ok {
		return nil, fmt.Errorf("no consumer offsets found for topic %s", subscription.Topic.String())
	}

	partitions := make(chan *adminpb.GetSubscriptionInfoResponse_PartitionInfo, partitionCount)
	wg := &errgroup.Group{}
	for i := range partitionCount {
		wg.Go(func() error {
			oldestOffset, err := client.GetOffset(subscription.Topic.String(), i, sarama.OffsetOldest)
			if err != nil {
				return fmt.Errorf("failed to get offset for partition %d: %w", i, err)
			}
			highwaterMark, err := client.GetOffset(subscription.Topic.String(), i, sarama.OffsetNewest)
			if err != nil {
				return fmt.Errorf("failed to get offset for partition %d: %w", i, err)
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
				return fmt.Errorf("failed to get event info for partition %d and offset %d: %w", i, headOffset, err)
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
					return fmt.Errorf("failed to get event info for partition %d and offset %d: %w", i, offsetBlock.Offset-1, err)
				}
			}
			if offsetBlock.Offset <= headOffset && offsetBlock.Offset >= oldestOffset {
				info.Next, err = getEventMetadata(ctx, connector.KafkaBrokers, subscription.Topic.String(), i, offsetBlock.Offset)
				if err != nil {
					return fmt.Errorf("failed to get event info for partition %d and offset %d: %w", i, offsetBlock.Offset, err)
				}
			}

			partitions <- info
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get partition info: %w", err)
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
			return 0, fmt.Errorf("failed to create kafka admin client: %w", err)
		}
		defer admin.Close()
	}

	topicMetas, err := admin.DescribeTopics([]string{topicID})
	if err != nil {
		return 0, fmt.Errorf("failed to describe topic %s: %w", topicID, err)
	}
	if len(topicMetas) != 1 {
		return 0, fmt.Errorf("expected topic metadata for %s from kafka but received none", topicID)
	}
	if topicMetas[0].Err == sarama.ErrUnknownTopicOrPartition {
		return 0, fmt.Errorf("can not reset subscription for topic %s that does not exist yet: %w", topicID, err)
	} else if topicMetas[0].Err != sarama.ErrNoError {
		return 0, fmt.Errorf("failed to describe topic %s: %w", topicID, topicMetas[0].Err)
	}
	return int32(len(topicMetas[0].Partitions)), nil //nolint:gosec
}

func getEventMetadata(ctx context.Context, brokers []string, topicID string, partition int32, offset int64) (*adminpb.PubSubEventMetadata, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topicID, partition, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to create partition consumer: %w", err)
	}
	defer partitionConsumer.Close()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled")
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for message")
	case err := <-partitionConsumer.Errors():
		return nil, fmt.Errorf("failed to consume partition: %w", err)
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
				return nil, fmt.Errorf("failed to parse request key: %w", err)
			}
		}
		if !found {
			return nil, errors.New("request key not found")
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
		return nil, fmt.Errorf("failed to get latest schema: %w", err)
	}
	return connect.NewResponse(sch.Msg), nil
}

func (s *Service) ApplyChangeset(ctx context.Context, req *connect.Request[adminpb.ApplyChangesetRequest], stream *connect.ServerStream[adminpb.ApplyChangesetResponse]) error {
	events := s.source.Subscribe(ctx)
	cs, err := s.schemaClient.CreateChangeset(ctx, connect.NewRequest(&ftlv1.CreateChangesetRequest{
		RealmChanges: islices.Map(req.Msg.RealmChanges, func(r *schemapb.RealmChange) *ftlv1.RealmChange {
			return &ftlv1.RealmChange{
				Modules:  r.Modules,
				ToRemove: r.ToRemove,
			}
		}),
	}))
	if err != nil {
		return fmt.Errorf("failed to create changeset: %w", err)
	}
	key, err := key.ParseChangesetKey(cs.Msg.Changeset)
	if err != nil {
		return fmt.Errorf("failed to parse changeset key: %w", err)
	}
	changeset := &schemapb.Changeset{
		Key:          cs.Msg.Changeset,
		RealmChanges: req.Msg.RealmChanges,
	}
	if err := stream.Send(&adminpb.ApplyChangesetResponse{
		Changeset: changeset,
	}); err != nil {
		return fmt.Errorf("failed to send changeset: %w", err)
	}
	for e := range channels.IterContext(ctx, events) {
		switch event := e.(type) {
		case *schema.ChangesetFinalizedNotification:
			if event.Key != key {
				continue
			}
			if err := stream.Send(&adminpb.ApplyChangesetResponse{
				Changeset: changeset,
			}); err != nil {
				return fmt.Errorf("failed to send changeset: %w", err)
			}
			return nil
		case *schema.ChangesetFailedNotification:
			if event.Key != key {
				continue
			}
			return fmt.Errorf("failed to apply changeset: %s", event.Error)
		case *schema.ChangesetCommittedNotification:
			if event.Changeset.Key != key {
				continue
			}
			changeset = event.Changeset.ToProto()
			// We don't wait for cleanup, just return immediately
			if err := stream.Send(&adminpb.ApplyChangesetResponse{
				Changeset: changeset,
			}); err != nil {
				return fmt.Errorf("failed to send changeset: %w", err)
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
	return fmt.Errorf("failed to apply changeset: context cancelled")
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
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.DeploymentRuntimeNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_DeploymentRuntimeNotification{DeploymentRuntimeNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetCreatedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetCreatedNotification{ChangesetCreatedNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetPreparedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetPreparedNotification{ChangesetPreparedNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetCommittedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetCommittedNotification{ChangesetCommittedNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetDrainedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetDrainedNotification{ChangesetDrainedNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetFinalizedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetFinalizedNotification{ChangesetFinalizedNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetRollingBackNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetRollingBackNotification{ChangesetRollingBackNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		case *schema.ChangesetFailedNotification:
			proto := e.ToProto()
			err := resp.Send(&ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetFailedNotification{ChangesetFailedNotification: proto},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
		}

	}
	return fmt.Errorf("context cancelled %w", ctx.Err())
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[adminpb.GetArtefactDiffsRequest]) (*connect.Response[adminpb.GetArtefactDiffsResponse], error) {
	byteDigests, err := islices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to parse digests: %w", err)
	}
	_, need, err := s.storage.GetDigestsKeys(ctx, byteDigests)
	if err != nil {
		return nil, fmt.Errorf("failed to get digests: %w", err)
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
			return fmt.Errorf("artefact size must be specified")
		}
		digest, err := sha256.ParseSHA256(hex.EncodeToString(msg.Digest))
		if err != nil {
			return fmt.Errorf("failed to parse digest: %w", err)
		}
		logger = logger.Scope("uploadArtefact:" + digest.String())
		logger.Debugf("Starting upload to OCI")
		err = s.storage.Upload(ctx, artefacts.ArtefactUpload{
			Digest:  digest,
			Size:    msg.Size,
			Content: r,
		})
		if err != nil {
			return fmt.Errorf("failed to upload artefact: %w", err)
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
				return fmt.Errorf("zero length chunk received")
			}
			firstMsg.Set(msg)
			if _, err := w.Write(msg.Chunk); err != nil {
				return fmt.Errorf("failed to write chunk: %w", err)
			}
		}
		if err := stream.Err(); err != nil {
			return fmt.Errorf("failed to upload artefact: %w", err)
		}
		return nil
	})
	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to upload artefact: %w", err)
	}
	return connect.NewResponse(&adminpb.UploadArtefactResponse{}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[adminpb.GetDeploymentArtefactsRequest], resp *connect.ServerStream[adminpb.GetDeploymentArtefactsResponse]) error {
	dkey, err := key.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	deploymentOpt := s.source.CanonicalView().Deployment(dkey)
	deployment, ok := deploymentOpt.Get()
	if !ok {
		return fmt.Errorf("could not get deployment: %s", req.Msg.DeploymentKey)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Get deployment artefacts for: %s", dkey.String())

	chunk := make([]byte, s.config.ArtefactChunkSize)
nextArtefact:
	for artefact := range islices.FilterVariants[schema.MetadataArtefact](deployment.Metadata) {
		deploymentArtefact := &state.DeploymentArtefact{
			Digest:     artefact.Digest,
			Path:       artefact.Path,
			Executable: artefact.Executable,
		}
		for _, clientArtefact := range req.Msg.HaveArtefacts {
			if proto.Equal(adminpb.ArtefactToProto(deploymentArtefact), clientArtefact) {
				continue nextArtefact
			}
		}
		reader, err := s.storage.Download(ctx, artefact.Digest)
		if err != nil {
			return fmt.Errorf("could not download artefact: %w", err)
		}
		defer reader.Close()
		for {

			n, err := reader.Read(chunk)
			if n != 0 {
				if err := resp.Send(&adminpb.GetDeploymentArtefactsResponse{
					Artefact: adminpb.ArtefactToProto(deploymentArtefact),
					Chunk:    chunk[:n],
				}); err != nil {
					return fmt.Errorf("could not send artefact chunk: %w", err)
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return fmt.Errorf("could not read artefact chunk: %w", err)
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
		return nil, fmt.Errorf("failed to parse verb: %w", err)
	}

	if err := schema.ValidateJSONCall(req.Msg.Body, ref, s.source.CanonicalView()); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	call, err := s.routeTable.Call(ctx, headers.CopyRequestForForwarding(req))
	if err != nil {
		return nil, fmt.Errorf("failed to call verb: %w", err)
	}
	return call, nil
}

func (s *Service) RollbackChangeset(ctx context.Context, c *connect.Request[ftlv1.RollbackChangesetRequest]) (*connect.Response[ftlv1.RollbackChangesetResponse], error) {
	res, err := s.schemaClient.RollbackChangeset(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to rollback changeset: %w", err)
	}
	return connect.NewResponse(res.Msg), nil
}

func (s *Service) FailChangeset(ctx context.Context, c *connect.Request[ftlv1.FailChangesetRequest]) (*connect.Response[ftlv1.FailChangesetResponse], error) {
	res, err := s.schemaClient.FailChangeset(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to fail changeset: %w", err)
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
		return fmt.Errorf("failed to get timeline: %w", err)
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
			return fmt.Errorf("failed to send logs: %w", err)
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
	_, err := s.schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{Update: c.Msg.Element, Realm: c.Msg.Realm}))
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment runtime: %w", err)
	}
	return connect.NewResponse(&adminpb.UpdateDeploymentRuntimeResponse{}), nil
}
