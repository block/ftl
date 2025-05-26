package pubsub

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/jpillora/backoff"

	cobservability "github.com/block/ftl/backend/controller/observability"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/runner/pubsub/observability"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/timelineclient"
)

// resetOffsetCommand is sent to each partition claim session to reset the offset
// err chan will be closed when the command is completed.
type resetOffsetCommand struct {
	latest bool
	err    chan error
}

//sumtype:decl
type partitionEvent interface {
	partitionEvent()
}

// claimedPartitionEvent is sent when a claimed partition session has started
// The resetOffset chan is used to send a resetOffsetCommand to the session
type claimedPartitionEvent struct {
	partition   int
	resetOffset chan resetOffsetCommand
}

func (claimedPartitionEvent) partitionEvent() {}

type lostPartitionEvent struct {
	partition int
}

func (lostPartitionEvent) partitionEvent() {}

type resetOffsetsEvent struct {
	latest     bool
	partitions []int
	result     chan result.Result[[]int]
}

func (resetOffsetsEvent) partitionEvent() {}

type consumer struct {
	moduleName          string
	deployment          key.Deployment
	verb                *schema.Verb
	subscriber          *schema.MetadataSubscriber
	retryParams         schema.RetryParams
	group               sarama.ConsumerGroup
	deadLetterPublisher optional.Option[*publisher]

	verbClient     VerbClient
	timelineClient *timelineclient.Client

	claimedPartitionsChan chan partitionEvent
}

func newConsumer(moduleName string, verb *schema.Verb, subscriber *schema.MetadataSubscriber, deployment key.Deployment,
	deadLetterPublisher optional.Option[*publisher], verbClient VerbClient, timelineClient *timelineclient.Client) (*consumer, error) {
	if verb.Runtime == nil {
		return nil, errors.Errorf("subscription %s has no runtime", verb.Name)
	}
	if verb.Runtime.SubscriptionConnector == nil {
		return nil, errors.Errorf("subscription %s has no subscription connector", verb.Name)
	}

	connection, ok := verb.Runtime.SubscriptionConnector.(*schema.PlaintextKafkaSubscriptionConnector)
	if !ok {
		return nil, errors.Errorf("only plaintext kafka subscription connector is supported, got %T", verb.Runtime.SubscriptionConnector)
	}

	if len(connection.KafkaBrokers) == 0 {
		return nil, errors.Errorf("subscription %s has no Kafka brokers", verb.Name)
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	switch subscriber.FromOffset {
	case schema.FromOffsetBeginning, schema.FromOffsetUnspecified:
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case schema.FromOffsetLatest:
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	groupID := kafkaConsumerGroupID(moduleName, verb)
	group, err := sarama.NewConsumerGroup(connection.KafkaBrokers, groupID, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create consumer group for subscription %s", verb.Name)
	}

	c := &consumer{
		moduleName:          moduleName,
		deployment:          deployment,
		verb:                verb,
		subscriber:          subscriber,
		group:               group,
		deadLetterPublisher: deadLetterPublisher,

		verbClient:     verbClient,
		timelineClient: timelineClient,

		claimedPartitionsChan: make(chan partitionEvent, 16),
	}
	retryMetada, ok := slices.FindVariant[*schema.MetadataRetry](verb.Metadata)
	if ok {
		retryParams, err := retryMetada.RetryParams()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse retry params for subscription %s", verb.Name)
		}
		c.retryParams = retryParams
	} else {
		c.retryParams = schema.RetryParams{}
	}

	return c, nil
}

func kafkaConsumerGroupID(moduleName string, verb *schema.Verb) string {
	return schema.RefKey{Module: moduleName, Name: verb.Name}.String()
}

func (c *consumer) kafkaTopicID() string {
	return c.subscriber.Topic.String()
}

func (c *consumer) Begin(ctx context.Context) error {
	// set up config
	log.FromContext(ctx).Debugf("Starting subscription for %v", c.verb.Name)

	go c.watchPartitions(ctx)
	go c.watchErrors(ctx)
	go c.subscribe(ctx)
	return nil
}

func (c *consumer) watchErrors(ctx context.Context) {
	logger := log.FromContext(ctx)
	for err := range channels.IterContext(ctx, c.group.Errors()) {
		logger.Errorf(err, "Consumer group error for %v", c.verb.Name)
	}
}

func (c *consumer) subscribe(ctx context.Context) {
	logger := log.FromContext(ctx)
	// Iterate over consumer sessions.
	//
	// `Consume` should be called inside an infinite loop, when a server-side rebalance happens,
	// the consumer session will need to be recreated to get the new claims.
	for {
		select {
		case <-ctx.Done():
			c.group.Close()
			return
		default:
		}

		err := c.group.Consume(ctx, []string{c.kafkaTopicID()}, c)
		if err != nil {
			logger.Errorf(err, "Consumer group session failed for %s", c.verb.Name)
		} else {
			logger.Debugf("Ending consumer group session for %s", c.verb.Name)
		}
	}
}

// watchPartitions keeps an up to date list of claimed partitions and watches for reset commands to execute on them.
func (c *consumer) watchPartitions(ctx context.Context) {
	activePartitions := map[int]claimedPartitionEvent{}
	for event := range channels.IterContext(ctx, c.claimedPartitionsChan) {
		switch event := event.(type) {
		case claimedPartitionEvent:
			activePartitions[event.partition] = event

		case lostPartitionEvent:
			delete(activePartitions, event.partition)

		case resetOffsetsEvent:
			results := make(chan result.Result[int], len(activePartitions))
			wg := &sync.WaitGroup{}
			for _, p := range activePartitions {
				if !slices.Contains(event.partitions, p.partition) {
					continue
				}
				wg.Add(1)
				go func() {
					resultChan := make(chan error)
					p.resetOffset <- resetOffsetCommand{
						latest: event.latest,
						err:    resultChan,
					}
					err := <-resultChan
					if err != nil {
						results <- result.Err[int](errors.Wrapf(err, "could not reset offset for %v partition %v", c.verb.Name, p.partition))
					} else {
						results <- result.Ok(p.partition)
					}
					wg.Done()
				}()
			}
			wg.Wait()
			close(results)
			errs := []error{}

			partitions := []int{}
			for r := range results {
				p, err := r.Result()
				if err != nil {
					errs = append(errs, err)
				} else {
					partitions = append(partitions, p)
				}
			}
			if len(errs) > 0 {
				event.result <- result.Err[[]int](errors.Join(errs...))
			} else {
				event.result <- result.Ok(partitions)
			}
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	logger := log.FromContext(session.Context())

	partitions := session.Claims()[c.kafkaTopicID()]
	logger.Debugf("Starting session for %v with partitions [%v]", c.verb.Name, strings.Join(slices.Map(partitions, func(partition int32) string { return strconv.Itoa(int(partition)) }), ","))

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited but before the
// offsets are committed for the very last time.
func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages(). Once the Messages() channel
// is closed, the Handler must finish its processing loop and exit.
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	logger := log.FromContext(ctx)

	closed := make(chan struct{})
	defer func() {
		c.claimedPartitionsChan <- lostPartitionEvent{partition: int(claim.Partition())}
		close(closed)
	}()

	resetOffset := make(chan resetOffsetCommand)
	c.claimedPartitionsChan <- claimedPartitionEvent{
		partition:   int(claim.Partition()),
		resetOffset: resetOffset,
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case cmd := <-resetOffset:
			if err := c.resetPartitionOffset(session, int(claim.Partition()), cmd.latest); err != nil {
				cmd.err <- err
			}
			close(cmd.err)
			// We have to exit because current events in claim.Messages() will not be relevant anymore
			// return errors.Errorf("cancelled due to offset reset")
			return nil
		case msg, ok := <-claim.Messages():
			if !ok || msg == nil {
				// Channel closed, rebalance or shutdown needed
				return nil
			}

			logger.Debugf("Consuming message from %v[%v:%v]", c.verb.Name, msg.Partition, msg.Offset)
			publishedAt, publisherRequestKey := parseHeaders(logger, msg.Headers)
			remainingRetries := c.retryParams.Count
			backoff := c.retryParams.MinBackoff
			for {
				startTime := time.Now()
				callCtx, callCancel := context.WithCancelCause(ctx)
				callChan := make(chan error)
				go func() {
					callChan <- c.call(callCtx, msg.Value, int(msg.Partition), int(msg.Offset), publisherRequestKey)
				}()
				var err error
				select {
				case cmd := <-resetOffset:
					logger.Debugf("Cancelled call for subscription %s due to offset reset", c.verb.Name)
					observability.PubSub.Consumed(ctx, c.subscriber.Topic.ToRefKey(), schema.RefKey{Module: c.moduleName, Name: c.verb.Name}, publishedAt.Default(startTime), errors.New("cancelled due to offset reset"))

					// Don't wait for call to end before resetting offsets as it may take a while.
					callCancel(errors.Wrap(context.Canceled, "cancelled due to offset reset"))
					if err := c.resetPartitionOffset(session, int(claim.Partition()), cmd.latest); err != nil {
						cmd.err <- err
					}
					close(cmd.err)
					// We have to exit because current events in claim.Messages() will not be relevant anymore
					// return errors.Errorf("cancelled due to offset reset")
					return nil

				case err = <-callChan:
					// close call context now that the call is finished
					if err != nil {
						callCancel(errors.Wrap(errors.Join(err, context.Canceled), "call failed"))
					} else {
						callCancel(errors.Wrap(context.Canceled, "call succeeded"))
					}
				}
				observability.PubSub.Consumed(ctx, c.subscriber.Topic.ToRefKey(), schema.RefKey{Module: c.moduleName, Name: c.verb.Name}, publishedAt.Default(startTime), err)
				if err == nil {
					break
				}
				select {
				case <-ctx.Done():
					// Do not commit the message if we did not succeed and the context is done.
					// No need to retry message either.
					logger.Errorf(err, "Failed to consume message from %v[%v,%v]", c.verb.Name, msg.Partition, msg.Offset)
					return nil
				default:
				}
				var connectErr *connect.Error
				if errors.As(err, &connectErr) {
					// Connection error, do not count as an attempt
					// This can happen when a runner is shutting down. This should never mark the message as consumed.
					time.Sleep(time.Millisecond * 500)
					continue
				}
				if remainingRetries == 0 {
					logger.Errorf(err, "Failed to consume message from %v[%v,%v]", c.verb.Name, msg.Partition, msg.Offset)
					if !c.publishToDeadLetterTopic(ctx, msg, err) {
						return nil
					}
					break
				}
				logger.Errorf(err, "Failed to consume message from %v[%v,%v] and will retry in %vs", c.verb.Name, msg.Partition, msg.Offset, int(backoff.Seconds()))
				time.Sleep(backoff)
				remainingRetries--
				backoff *= 2
				if backoff > c.retryParams.MaxBackoff {
					backoff = c.retryParams.MaxBackoff
				}
			}
			session.MarkMessage(msg, "")
		}
	}
}

// parseHeaders extracts FTL header values for a kafka message
func parseHeaders(logger *log.Logger, headers []*sarama.RecordHeader) (publishedAt optional.Option[time.Time], requestKey optional.Option[key.Request]) {
	for _, h := range headers {
		switch string(h.Key) {
		case createdAtHeader:
			t, err := time.Parse(time.RFC3339Nano, string(h.Value))
			if err != nil {
				logger.Warnf("failed to parse %s header: %v", createdAtHeader, err)
			} else {
				publishedAt = optional.Some(t)
			}

		case requestKeyHeader:
			k, err := key.ParseRequestKey(string(h.Value))
			if err != nil {
				logger.Warnf("failed to parse %s header: %v", requestKeyHeader, err)
			} else {
				requestKey = optional.Some(k)
			}
		}
	}
	return
}

func (c *consumer) call(ctx context.Context, body []byte, partition, offset int, publisherRequestKey optional.Option[key.Request]) error {
	start := time.Now()
	requestKey := publisherRequestKey.Default(key.NewRequestKey(key.OriginPubsub, schema.RefKey{Module: c.moduleName, Name: c.verb.Name}.String()))
	destRef := &schema.Ref{
		Module: c.moduleName,
		Name:   c.verb.Name,
	}
	req := &ftlv1.CallRequest{
		Verb: schema.RefKey{Module: c.moduleName, Name: c.verb.Name}.ToProto(),
		Body: body,
	}
	consumeEvent := timelineclient.PubSubConsume{
		DeploymentKey: c.deployment,
		RequestKey:    optional.Some(requestKey.String()),
		Time:          time.Now(),
		DestVerb:      optional.Some(destRef.ToRefKey()),
		Topic:         c.subscriber.Topic.String(),
		Partition:     partition,
		Offset:        offset,
	}
	defer c.timelineClient.Publish(ctx, consumeEvent)

	callEvent := &timelineclient.Call{
		DeploymentKey: c.deployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      destRef,
		Callers:       []*schema.Ref{},
		Request:       req,
	}
	defer c.timelineClient.Publish(ctx, callEvent)

	request := connect.NewRequest(req)
	headers.SetRequestKey(request.Header(), requestKey)
	resp, callErr := c.verbClient.Call(ctx, request)
	if callErr == nil {
		if errResp, ok := resp.Msg.Response.(*ftlv1.CallResponse_Error_); ok {
			callErr = errors.Errorf("verb call failed: %s", errResp.Error.Message)
		}
	}
	if callErr != nil {
		consumeEvent.Error = optional.Some(callErr.Error())
		callEvent.Response = result.Err[*ftlv1.CallResponse](callErr)
		cobservability.Calls.Request(ctx, req.Verb, start, optional.Some("verb call failed"))
		return errors.WithStack(callErr)
	}
	callEvent.Response = result.Ok(resp.Msg)
	cobservability.Calls.Request(ctx, req.Verb, start, optional.None[string]())
	return nil
}

// publishToDeadLetterTopic tries to publish the message to the dead letter topic.
//
// If it does not succeed it will retry until it succeeds or the context is done.
// Returns true if the message was published or if there is no dead letter queue.
// Returns false if the context is done.
func (c *consumer) publishToDeadLetterTopic(ctx context.Context, msg *sarama.ConsumerMessage, callErr error) bool {
	p, ok := c.deadLetterPublisher.Get()
	if !ok {
		return true
	}

	deadLetterEvent, err := encoding.Marshal(map[string]any{
		"event": json.RawMessage(msg.Value),
		"error": callErr.Error(),
	})
	if err != nil {
		panic(errors.Wrapf(err, "failed to marshal dead letter event for %v on partition %v and offset %v", c.kafkaTopicID(), msg.Partition, msg.Offset))
	}

	bo := &backoff.Backoff{Min: time.Second, Max: 10 * time.Second}
	first := true
	for {
		var waitDuration time.Duration
		if first {
			first = false
		} else {
			waitDuration = bo.Duration()
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(waitDuration):
		}
		err := p.publish(ctx, deadLetterEvent, string(msg.Key), schema.Ref{Module: c.moduleName, Name: c.verb.Name})
		if err == nil {
			return true
		}
	}
}

func (c *consumer) ResetOffsetsForClaimedPartitions(ctx context.Context, latest bool, requestedPartitions []int) (partitions []int, err error) {
	resultChan := make(chan result.Result[[]int])
	c.claimedPartitionsChan <- resetOffsetsEvent{latest: latest, result: resultChan, partitions: requestedPartitions}
	result, err := (<-resultChan).Result()
	if err != nil {
		return nil, errors.WithStack(err) //nolint:wrapcheck
	}
	return result, nil
}

func (c *consumer) resetPartitionOffset(session sarama.ConsumerGroupSession, partition int, latest bool) error {
	connection, ok := c.verb.Runtime.SubscriptionConnector.(*schema.PlaintextKafkaSubscriptionConnector)
	if !ok {
		return errors.Errorf("only plaintext kafka subscription connector is supported, got %T", c.verb.Runtime.SubscriptionConnector)
	}

	config := sarama.NewConfig()
	client, err := sarama.NewClient(connection.KafkaBrokers, config)
	if err != nil {
		return errors.Wrapf(err, "failed to create client for subscription %s", c.verb.Name)
	}

	var offsetTime int64
	if latest {
		offsetTime = sarama.OffsetNewest
	} else {
		offsetTime = sarama.OffsetOldest
	}
	newOffset, err := client.GetOffset(c.kafkaTopicID(), int32(partition), offsetTime)
	if err != nil {
		return errors.Wrapf(err, "failed to get offset for %v partition %v", c.verb.Name, partition)
	}

	if latest {
		session.MarkOffset(c.kafkaTopicID(), int32(partition), newOffset, "")
	} else {
		session.ResetOffset(c.kafkaTopicID(), int32(partition), newOffset, "")
	}
	// no simple way to get an error corresponding to committing this offset...
	session.Commit()
	return nil
}
