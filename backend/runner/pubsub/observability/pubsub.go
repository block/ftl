package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/observability"
)

// To learn more about how sinks and subscriptions work together, check out the
// https://block.github.io/ftl/docs/reference/pubsub/
const (
	meterName              = "ftl.pubsub"
	topicRefAttr           = "ftl.pubsub.topic.ref"
	topicModuleAttr        = "ftl.pubsub.topic.module.name"
	callerVerbRefAttr      = "ftl.pubsub.publish.caller.verb.ref"
	subscriptionRefAttr    = "ftl.pubsub.subscription.ref"
	subscriptionModuleAttr = "ftl.pubsub.subscription.module.name"
	// We do not know publication date anymore
	// timeSinceScheduledAtBucketAttr = "ftl.pubsub.time_since_scheduled_at_ms.bucket"
)

type PubSubMetrics struct {
	published   metric.Int64Counter
	consumed    metric.Int64Counter
	msToConsume metric.Int64Histogram
}

func initPubSubMetrics() (*PubSubMetrics, error) {
	result := &PubSubMetrics{
		published:   noop.Int64Counter{},
		consumed:    noop.Int64Counter{},
		msToConsume: noop.Int64Histogram{},
	}

	var err error
	meter := otel.Meter(meterName)

	counterName := fmt.Sprintf("%s.published", meterName)
	if result.published, err = meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that an event is published to a topic")); err != nil {
		return nil, wrapErr(counterName, err)
	}

	signalName := fmt.Sprintf("%s.consumed", meterName)
	if result.consumed, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries completing an async call")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.ms_to_consume", meterName)
	if result.msToConsume, err = meter.Int64Histogram(signalName, metric.WithUnit("ms"),
		metric.WithDescription("duration in ms to complete an async call, from the earliest time it was scheduled to execute")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *PubSubMetrics) Published(ctx context.Context, module, topic, caller string, maybeErr error) {
	attrs := []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(topicRefAttr, schema.RefKey{Module: module, Name: topic}.String()),
		attribute.String(callerVerbRefAttr, schema.RefKey{Module: module, Name: caller}.String()),
		observability.SuccessOrFailureStatusAttr(maybeErr == nil),
	}

	m.published.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *PubSubMetrics) Consumed(ctx context.Context, topic, subscription schema.RefKey, startTime time.Time, maybeErr error) {
	// This used to be time since publication time, not consumption start time.
	// We should consider changing this back to time since publication time.
	msToComplete := time.Since(startTime).Milliseconds()

	attrs := []attribute.KeyValue{
		attribute.String(topicRefAttr, schema.RefKey{Module: topic.Module, Name: topic.Name}.String()),
		attribute.String(topicModuleAttr, topic.Module),
		attribute.String(subscriptionRefAttr, subscription.String()),
		attribute.String(subscriptionModuleAttr, subscription.Module),
		observability.SuccessOrFailureStatusAttr(maybeErr == nil),
	}

	m.msToConsume.Record(ctx, msToComplete, metric.WithAttributes(attrs...))
	m.consumed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func wrapErr(signalName string, err error) error {
	return fmt.Errorf("failed to create %q signal: %w", signalName, err)
}
