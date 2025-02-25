Declare a topic for publishing events.

Topics are where events are sent in the PubSub system. A topic can be exported to allow other modules to subscribe to it. Each topic can have multiple subscriptions.

```kotlin
import xyz.block.ftl.Export
import xyz.block.ftl.SinglePartitionMapper
import xyz.block.ftl.Topic
import xyz.block.ftl.WriteableTopic

// Define the event type for the topic
data class Invoice(
	val invoiceNo: String
)

// Add @Export if you want other modules to be able to consume from this topic
@Topic(name = "invoices", partitions = 1)
internal interface InvoicesTopic : WriteableTopic<Invoice, SinglePartitionMapper>
```

See https://block.github.io/ftl/docs/reference/pubsub/
---
data class ${1:Event}(
	${2:// Event fields}
)

@Export
@Topic(name = "${3:topicName}", partitions = 1)
interface ${4:TopicName} : WriteableTopic<${1:Event}, SinglePartitionMapper> 
