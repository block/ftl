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
	"sync"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/controller/state"
	pubsubpb "github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
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
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type Config struct {
	Bind              *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8896" env:"FTL_BIND"`
	ArtefactChunkSize int      `help:"Size of each chunk streamed to the client." default:"1048576"`
}

type Service struct {
	env          *EnvironmentManager
	schemaClient ftlv1connect.SchemaServiceClient
	source       *schemaeventsource.EventSource
	storage      *artefacts.OCIArtefactService
	config       Config
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
func NewAdminService(config Config, env *EnvironmentManager, schr ftlv1connect.SchemaServiceClient, source *schemaeventsource.EventSource, storage *artefacts.OCIArtefactService) *Service {
	return &Service{
		config:       config,
		env:          env,
		schemaClient: schr,
		source:       source,
		storage:      storage,
	}
}

func Start(
	ctx context.Context,
	config Config,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	schr ftlv1connect.SchemaServiceClient, source *schemaeventsource.EventSource,
	storage *artefacts.OCIArtefactService) error {

	logger := log.FromContext(ctx).Scope("admin")

	svc := NewAdminService(config, &EnvironmentManager{schr: NewSchemaRetriever(source), cm: cm, sm: sm}, schr, source, storage)

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

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	byteDigests, err := islices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, err
	}
	_, need, err := s.storage.GetDigestsKeys(ctx, byteDigests)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: islices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, stream *connect.ClientStream[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx).Scope("uploadArtefact")
	firstMsg := NewOnceValue[*ftlv1.UploadArtefactRequest]()
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
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
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
			if proto.Equal(ftlv1.ArtefactToProto(deploymentArtefact), clientArtefact) {
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
				if err := resp.Send(&ftlv1.GetDeploymentArtefactsResponse{
					Artefact: ftlv1.ArtefactToProto(deploymentArtefact),
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

func (s *Service) ClusterInfo(ctx context.Context, req *connect.Request[ftlv1.ClusterInfoRequest]) (*connect.Response[ftlv1.ClusterInfoResponse], error) {
	return connect.NewResponse(&ftlv1.ClusterInfoResponse{Os: runtime.GOOS, Arch: runtime.GOARCH}), nil
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
