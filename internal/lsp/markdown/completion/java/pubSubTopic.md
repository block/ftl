Declare a topic for publishing events.

Topics are where events are sent in the PubSub system. A topic can be exported to allow other modules to subscribe to it. Each topic can have multiple subscriptions.

```java
import xyz.block.ftl.Export;
import xyz.block.ftl.SinglePartitionMapper;
import xyz.block.ftl.Topic;
import xyz.block.ftl.WriteableTopic;

// Define the event type for the topic
record Invoice(String invoiceNo) {
}

// Add @Export if you want other modules to be able to consume from this topic
@Topic(name = "invoices", partitions = 1)
interface InvoicesTopic extends WriteableTopic<Invoice, SinglePartitionMapper> {
}
```

See https://block.github.io/ftl/docs/reference/pubsub/
---

record ${1:Event}(${2:// Event fields}) {
}

@Export
@Topic(partitions = 1)
interface ${4:TopicName} extends WriteableTopic<${1:Event}, SinglePartitionMapper> {
} 
