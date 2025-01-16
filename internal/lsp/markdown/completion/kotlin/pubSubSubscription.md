Declare a subscription to a topic.

A subscription consumes events from a topic, starting either from the beginning of the topic or only new events.

```kotlin
import ftl.publisher.Invoice
import ftl.publisher.InvoicesTopic

import xyz.block.ftl.FromOffset
import xyz.block.ftl.Subscription

@Subscription(topic = InvoicesTopic::class, from = FromOffset.LATEST)
fun consumeInvoice(event: Invoice) {
	// Process the invoice event
}
```

See https://block.github.io/ftl/docs/reference/pubsub/
---

@Subscription(topic = ${1:TopicClass}::class, from = FromOffset.${2:LATEST})
fun ${3:onEvent}(event: ${4:EventType}) {
	${5:// TODO: Implement}
} 
