package subscriber

import (
	"context"
	"ftl/pubsub"
)

//ftl:verb
//ftl:subscribe pubsub.publicBroadcast from=beginning
func ConsumesSubscriptionFromExternalTopic(ctx context.Context, req pubsub.PayinEvent) error {
	return nil
}
