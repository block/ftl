Declare a subscription to a topic.

A subscription consumes events from a topic, starting either from the beginning of the topic or only new events.

```java
import ftl.othermodule.Invoice;
import ftl.othermodule.InvoicesTopic;

import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Subscription;

class PaymentProcessor {
	// Subscribe to payment events from the latest offset
	@Subscription(topic = InvoicesTopic.class, from = FromOffset.LATEST)
	public void onInvoice(Invoice event) {
		// Process the invoice event
	}
}
```

See https://block.github.io/ftl/docs/reference/pubsub/
---

class ${1:Name} {
	@Subscription(topic = ${2:TopicClass}.class, from = FromOffset.${3:LATEST})
	public void ${4:onEvent}(${5:EventType} event) {
		${6:// TODO: Implement}
	}
} 
