---
sidebar_position: 11
title: PubSub
description: Asynchronous publishing of events to topics
---

# PubSub

FTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent) and subscribers (a verb which consumes events). Subscribers are, as you would expect, sinks. Each subscriber is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each published event has an at least once delivery guarantee for each subscription.

A topic can be exported to allow other modules to subscribe to it. Subscriptions are always private to their module.

When a subscription is first created in an environment, it can start consuming from the beginning of the topic or only consume events published afterwards.

Topics allow configuring the number of partitions and how each event should be mapped to a partition, allowing for greater throughput. Subscriptions will consume in order within each partition. There are cases where a small amount of progress on a subscription will be lost, so subscriptions should be able to handle receiving some events that have already been consumed.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
<TabItem value="go" label="Go" default>

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

//ftl:topic partitions=1
type Invoices = ftl.TopicHandle[Invoice, ftl.SinglePartitionMap[Invoice]]
```

If you want multiple partitions in the topic, you'll also need to write a partition mapper:

```go
package payments

import (
  "github.com/block/ftl/go-runtime/ftl"
)

// Define an event type
type Invoice struct {
  InvoiceNo string
}

type PartitionMapper struct{}

var _ ftl.TopicPartitionMap[PubSubEvent] = PartitionMapper{}

func (PartitionMapper) PartitionKey(event PubSubEvent) string {
	return event.Time.String()
}

//ftl:topic partitions=10
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

</TabItem>
<TabItem value="kotlin" label="Kotlin">

First, declare a new topic:

```kotlin
import xyz.block.ftl.Export;
import xyz.block.ftl.SinglePartitionMapper
import xyz.block.ftl.Topic
import xyz.block.ftl.WriteableTopic

// Define the event type for the topic
data class Invoice(val invoiceNo: String)

// Add @Export if you want other modules to be able to consum from this topic
@Topic(name = "invoices", partitions = 1)
internal interface InvoicesTopic : WriteableTopic<Invoice, SinglePartitionMapper>
```

If you want multiple partitions in the topic, you'll also need to write a partition mapper:

```kotlin
import xyz.block.ftl.Export;
import xyz.block.ftl.SinglePartitionMapper
import xyz.block.ftl.Topic
import xyz.block.ftl.TopicPartitionMapper
import xyz.block.ftl.WriteableTopic

// Define the event type for the topic
data class Invoice(val invoiceNo: String)

// PartitionMapper maps each to a partition in the topic
class PartitionMapper : TopicPartitionMapper<Invoice> {
    override fun getPartitionKey(invoice: Invoice): String {
        return invoice.invoiceNo
    }
}

// Add @Export if you want other modules to be able to consum from this topic
@Topic(name = "invoices", partitions = 8)
internal interface InvoicesTopic : WriteableTopic<Invoice, PartitionMapper>
```

Events can be published to a topic by injecting it into an `@Verb` method:

```kotlin
@Verb
fun publishInvoice(request: InvoiceRequest, topic: InvoicesTopic) {
    topic.publish(Invoice(request.invoiceNo))
}
```

To subscribe to a topic use the `@Subscription` annotation, referencing the topic class and providing a method to consume the event:

```kotlin
// if subscribing from another module, import the event and topic
import ftl.publisher.Invoice
import ftl.publisher.InvoicesTopic

import xyz.block.ftl.FromOffset
import xyz.block.ftl.Subscription

@Subscription(topic = InvoicesTopic::class, from = FromOffset.LATEST)
fun consumeInvoice(event: Invoice) {
    // ...
}
```

If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated
topic cannot be published to, only subscribed to:

```kotlin
@Topic(name="invoices", module="publisher")
internal interface InvoicesTopic : ConsumableTopic<Invoice>
```

</TabItem>
<TabItem value="java" label="Java">

First, declare a new topic:

```java
import xyz.block.ftl.Export;
import xyz.block.ftl.SinglePartitionMapper;
import xyz.block.ftl.Topic;
import xyz.block.ftl.WriteableTopic;

// Define the event type for the topic
record Invoice(String invoiceNo) {
}

// Add @Export if you want other modules to be able to consum from this topic
@Topic(name = "invoices", partitions = 1)
interface InvoicesTopic extends WriteableTopic<Invoice, SinglePartitionMapper> {
}
```

If you want multiple partitions in the topic, you'll also need to write a partition mapper:

```java
import xyz.block.ftl.Export;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicPartitionMapper;
import xyz.block.ftl.WriteableTopic;

// Define the event type for the topic
record Invoice(String invoiceNo) {
}

// PartitionMapper maps each to a partition in the topic
class PartitionMapper implements TopicPartitionMapper<Invoice> {
    public String getPartitionKey(Invoice invoice) {
        return invoice.invoiceNo();
    }
}

// Add @Export if you want other modules to be able to consum from this topic
@Topic(name = "invoices", partitions = 8)
interface InvoicesTopic extends WriteableTopic<Invoice, PartitionMapper> {
}
```

Events can be published to a topic by injecting it into an `@Verb` method:

```java
@Verb
void publishInvoice(InvoiceRequest request, InvoicesTopic topic) throws Exception {
    topic.publish(new Invoice(request.invoiceNo()));
}
```

To subscribe to a topic use the `@Subscription` annotation, referencing the topic class and providing a method to consume the event:

```java
// if subscribing from another module, import the event and topic
import ftl.othermodule.Invoice;
import ftl.othermodule.InvoicesTopic;

import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Subscription;

class Subscriber {
    @Subscription(topic = InvoicesTopic.class, from = FromOffset.LATEST)
    public void consumeInvoice(Invoice event) {
        // ...
    }
}
```

If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated
topic cannot be published to, only subscribed to:

```java
@Topic(name="invoices", module="publisher")
interface InvoicesTopic extends ConsumableTopic<Invoice> {}
```

</TabItem>
</Tabs> 
