package admin

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"

	pubsubpb "github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type Config struct {
	Bind *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8896" env:"FTL_BIND"`
}

type Service struct {
	env          *EnvironmentManager
	schemaClient ftlv1connect.SchemaServiceClient
	source       *schemaeventsource.EventSource
}

var _ ftlv1connect.AdminServiceHandler = (*Service)(nil)

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
	return &schema.Schema{Modules: view.Modules}, nil
}

// NewAdminService creates a new Service.
// bindAllocator is optional and should be set if a local client is to be used that accesses schema from disk using language plugins.
func NewAdminService(env *EnvironmentManager, schr ftlv1connect.SchemaServiceClient, source *schemaeventsource.EventSource) *Service {
	return &Service{
		env:          env,
		schemaClient: schr,
		source:       source,
	}
}

func Start(
	ctx context.Context,
	config Config,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	schr ftlv1connect.SchemaServiceClient, source *schemaeventsource.EventSource,
) error {

	logger := log.FromContext(ctx).Scope("admin")

	svc := NewAdminService(&EnvironmentManager{schr: NewSchemaRetriever(source), cm: cm, sm: sm}, schr, source)

	logger.Debugf("Admin service listening on: %s", config.Bind)
	err := rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewAdminServiceHandler, svc),
	)
	if err != nil {
		return fmt.Errorf("admin service stopped serving: %w", err)
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *Service) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ConfigListRequest]) (*connect.Response[ftlv1.ConfigListResponse], error) {
	return s.env.ConfigList(ctx, req)
}

// ConfigGet returns the configuration value for a given ref string.
func (s *Service) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.ConfigGetRequest]) (*connect.Response[ftlv1.ConfigGetResponse], error) {
	return s.env.ConfigGet(ctx, req)
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *Service) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.ConfigSetRequest]) (*connect.Response[ftlv1.ConfigSetResponse], error) {
	return s.env.ConfigSet(ctx, req)
}

// ConfigUnset unsets the config value at the given ref.
func (s *Service) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.ConfigUnsetRequest]) (*connect.Response[ftlv1.ConfigUnsetResponse], error) {
	return s.env.ConfigUnset(ctx, req)
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *Service) SecretsList(ctx context.Context, req *connect.Request[ftlv1.SecretsListRequest]) (*connect.Response[ftlv1.SecretsListResponse], error) {
	return s.env.SecretsList(ctx, req)
}

// SecretGet returns the secret value for a given ref string.
func (s *Service) SecretGet(ctx context.Context, req *connect.Request[ftlv1.SecretGetRequest]) (*connect.Response[ftlv1.SecretGetResponse], error) {
	return s.env.SecretGet(ctx, req)
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *Service) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SecretSetRequest]) (*connect.Response[ftlv1.SecretSetResponse], error) {
	return s.env.SecretSet(ctx, req)
}

// SecretUnset unsets the secret value at the given ref.
func (s *Service) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.SecretUnsetRequest]) (*connect.Response[ftlv1.SecretUnsetResponse], error) {
	return s.env.SecretUnset(ctx, req)
}

// MapConfigsForModule combines all configuration values visible to the module.
func (s *Service) MapConfigsForModule(ctx context.Context, req *connect.Request[ftlv1.MapConfigsForModuleRequest]) (*connect.Response[ftlv1.MapConfigsForModuleResponse], error) {
	return s.env.MapConfigsForModule(ctx, req)
}

// MapSecretsForModule combines all secrets visible to the module.
func (s *Service) MapSecretsForModule(ctx context.Context, req *connect.Request[ftlv1.MapSecretsForModuleRequest]) (*connect.Response[ftlv1.MapSecretsForModuleResponse], error) {
	return s.env.MapSecretsForModule(ctx, req)
}

func (s *Service) ResetSubscription(ctx context.Context, req *connect.Request[ftlv1.ResetSubscriptionRequest]) (*connect.Response[ftlv1.ResetSubscriptionResponse], error) {
	// Find nodes in schema
	// TODO: we really want all deployments for a module... not just latest... Use canonical and check ActiveChangeset?
	sch := s.source.CanonicalView()
	module, ok := islices.Find(sch.Modules, func(m *schema.Module) bool {
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
	if verb.Runtime == nil || verb.Runtime.Subscription == nil || len(verb.Runtime.Subscription.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("no Kafka brokers for subscription %q", req.Msg.Subscription)
	}
	if module.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
		return nil, fmt.Errorf("no deployment for module %s", req.Msg.Subscription.Module)
	}
	topicID := subscriber.Topic.String()
	totalPartitions, err := kafkaPartitionCount(ctx, verb.Runtime.Subscription.KafkaBrokers, topicID)
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
	var failedPartitions = []int{}
	for p := range totalPartitions {
		if p >= len(successfulPartitions) || int(successfulPartitions[p]) != p {
			failedPartitions = append(failedPartitions, p)
		}
	}
	if len(failedPartitions) > 0 {
		return nil, fmt.Errorf("failed to reset partitions %v: no runner had partition claim", failedPartitions)
	}
	return connect.NewResponse(&ftlv1.ResetSubscriptionResponse{}), nil
}

// kafkaPartitionCount returns the number of partitions for a given topic in kafka. This may differ from the number
// of partitions in the schema if the topic was originally provisioned with a different number of partitions.
func kafkaPartitionCount(ctx context.Context, brokers []string, topicID string) (int, error) {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return 0, fmt.Errorf("failed to create kafka admin client: %w", err)
	}
	defer admin.Close()

	topicMetas, err := admin.DescribeTopics([]string{topicID})
	if err != nil {
		return 0, fmt.Errorf("failed to describe topic %s: %w", topicID, err)
	}
	log.FromContext(ctx).Infof("topic metadata for %s: %v", topicID, topicMetas)
	if len(topicMetas) != 1 {
		return 0, fmt.Errorf("expected topic metadata for %s from kafka but received none", topicID)
	}
	if topicMetas[0].Err == sarama.ErrUnknownTopicOrPartition {
		return 0, fmt.Errorf("can not reset subscription for topic %s that does not exist yet: %w", topicID, err)
	} else if topicMetas[0].Err != sarama.ErrNoError {
		return 0, fmt.Errorf("failed to describe topic %s: %w", topicID, topicMetas[0].Err)
	}
	return len(topicMetas[0].Partitions), nil
}

func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	sch, err := s.schemaClient.GetSchema(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to get latest schema: %w", err)
	}
	return connect.NewResponse(sch.Msg), nil
}

func (s *Service) ApplyChangeset(ctx context.Context, req *connect.Request[ftlv1.ApplyChangesetRequest]) (*connect.Response[ftlv1.ApplyChangesetResponse], error) {
	events := s.source.Subscribe(ctx)
	cs, err := s.schemaClient.CreateChangeset(ctx, connect.NewRequest(&ftlv1.CreateChangesetRequest{
		Modules:  req.Msg.Modules,
		ToRemove: req.Msg.ToRemove,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset: %w", err)
	}
	changeset := &schemapb.Changeset{
		Key:      cs.Msg.Changeset,
		Modules:  req.Msg.Modules,
		ToRemove: req.Msg.ToRemove,
	}
	for e := range channels.IterContext(ctx, events) {
		switch event := e.(type) {
		case *schema.ChangesetFinalizedNotification:
			//
			return connect.NewResponse(&ftlv1.ApplyChangesetResponse{
				Changeset: changeset,
			}), nil
		case *schema.ChangesetFailedNotification:
			return nil, fmt.Errorf("failed to apply changeset: %s", event.Error)
		case *schema.ChangesetCommittedNotification:
			changeset = event.Changeset.ToProto()
			// We don't wait for cleanup, just return immediately
			return connect.NewResponse(&ftlv1.ApplyChangesetResponse{
				Changeset: changeset,
			}), nil
		case *schema.ChangesetRollingBackNotification:
			changeset = event.Changeset.ToProto()
		default:

		}
	}
	return nil, fmt.Errorf("failed to apply changeset: context cancelled")
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
