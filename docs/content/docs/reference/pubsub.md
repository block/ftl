+++
title = "PubSub"
description = "Asynchronous publishing of events to topics"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 80
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

FTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent) and subscribers (a verb which consumes events). Subscribers are, as you would expect, sinks. Each subscriber is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each published event has an at least once delivery guarantee for each subscription.

A topic can be exported to allow other modules to subscribe to it. Subscriptions are always private to their module.

When a subscription is first created in an environment, it can start consuming from the beginning of the topic or only consume events published afterwards.

Topics allow configuring the number of partitions and how each event should be mapped to a partition, allowing for greater throughput. Subscriptions will consume in order within each partition. There are cases where a small amount of progress on a subscription will be lost, so subscriptions should be able to handle receiving some events that have already been consumed.

{% code_selector() %}
<!-- go -->

First, declare a new topic:

```go
package payments

import (
  "github.com/block/ftl/go-runtime/ftl"
)

// Define an event type
type Invoice struct {
  InvoiceNo string
}

// ftl.TopicPartitionMap is an interface for mapping each event to a partition in the topic.
// 
// If creating a topic with multiple partitions, you'll need to define a partition mapper for your event type.
// Otherwise you can use ftl.SinglePartitionMap[Event]
type PartitionMapper struct{}

var _ ftl.TopicPartitionMap[PubSubEvent] = PartitionMapper{}

func (PartitionMapper) PartitionKey(event PubSubEvent) string {
	return event.Time.String()
}

//ftl:topic export partitions=10
type Invoices = ftl.TopicHandle[Invoice, PartitionMapper]
```

Note that the name of the topic as represented in the FTL schema is the lower camel case version of the type name.

The `Invoices` type is a handle to the topic. It is a generic type that takes two arguments: the event type and the partition map type. The partition map type is used to map events to partitions.

Then define a Sink to consume from the topic:

```go
// Configure initial event consumption with either from=beginning or from=latest
//
//ftl:subscribe payments.invoices from=beginning
func SendInvoiceEmail(ctx context.Context, in Invoice) error {
  // ...
}
```

Events can be published to a topic by injecting the topic type into a verb:

```go
//ftl:verb
func PublishInvoice(ctx context.Context, topic Invoices) error {
   topic.Publish(ctx, Invoice{...})
   // ...
}
```

<!-- kotlin -->

First, declare a new topic:

```kotlin

import com.block.ftl.WriteableTopic

// Define the event type for the topic
data class Invoice(val invoiceNo: String)

// PartitionMapper maps each to a partition in the topic
class PartitionMapper : TopicPartitionMapper<Invoice> {
    override fun getPartitionKey(invoice: Invoice): String {
        return invoice.getInvoiceNo()
    }
}

@Export
@Topic(name = "invoices", partitions = 8)
internal interface InvoiceTopic : WriteableTopic<Invoice>
```

Events can be published to a topic by injecting it into an `@Verb` method:

```kotlin
@Verb
fun publishInvoice(request: InvoiceRequest, topic: InvoiceTopic) {
    topic.publish(Invoice(request.getInvoiceNo()))
}
```

To subscribe to a topic use the `@Subscription` annotation, referencing the topic class and providing a method to consume the event:

```kotlin
@Subscription(topic = InvoiceTopic::class, from = FromOffset.LATEST)
fun consumeInvoice(event: Invoice) {
    // ...
}
```

If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated
topic cannot be published to, only subscribed to:

```kotlin
@Topic(name="invoices", module="publisher")
internal interface InvoiceTopic : ConsumableTopic<Invoice>
```

<!-- java -->

First, declare a new topic:

```java
import com.block.ftl.WriteableTopic;

// Define the event type for the topic
record Invoice(String invoiceNo) {}

// PartitionMapper maps each to a partition in the topic
class PartitionMapper implements TopicPartitionMapper<Invoice> {
    public String getPartitionKey(Invoice invoice) {
        return invoice.getInvoiceNo();
    }
}

@Export
@Topic(name = "invoices", partitions = 8)
public interface InvoiceTopic extends WriteableTopic<Invoice> {}
```

Events can be published to a topic by injecting it into an `@Verb` method:

```java
@Verb
void publishInvoice(InvoiceRequest request, InvoiceTopic topic) throws Exception {
    topic.publish(new Invoice(request.getInvoiceNo()));
}
```

To subscribe to a topic use the `@Subscription` annotation, referencing the topic class and providing a method to consume the event:

```java
@Subscription(topic = InvoiceTopic.class, from = FromOffset.LATEST)
public void consumeInvoice(Invoice event) {
    // ...
}
```

If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated
topic cannot be published to, only subscribed to:

```java
@Topic(name="invoices", module="publisher")
 interface InvoiceTopic extends ConsumableTopic<Invoice> {}
```

{% end %}
> **NOTE!**
> PubSub topics cannot be published to from outside the module that declared them, they can only be subscribed to. That is, if a topic is declared in module `A`, module `B` cannot publish to it.
