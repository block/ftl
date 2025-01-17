package subscriber_test

import (
	"ftl/pubsub"
	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/go-runtime/ftl/ftltest"
	"testing"
)

func TestPublishToExternalModule(t *testing.T) {
	ctx := ftltest.Context()

	assert.NoError(t, pubsub.Topic1.Publish(ctx, pubsub.Event{Value: "external"}))

	ftltest.WaitForSubscriptionsToComplete(ctx)

	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, pubsub.Topic1)))

	// Make sure we correctly made the right ref for the external module.
	assert.Equal(t, "pubsub", pubsub.Topic1.Ref.Module)
}
