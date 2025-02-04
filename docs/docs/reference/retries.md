---
sidebar_position: 12
title: Retries
description: Retrying asynchronous verbs
---

# Retries

Some FTL features allow specifying a retry policy via a language-specific directive. Retries back off exponentially until the maximum is reached.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
<TabItem value="go" label="Go" default>

The directive has the following syntax:

```go
//ftl:retry [<attempts=10>] <min-backoff> [<max-backoff=1hr>] [catch <catchVerb>]
```

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```go
//ftl:retry 10 5s 1m
func Process(ctx context.Context, in Invoice) error {
  // ...
}
```

### PubSub Subscribers

Subscribers can have a retry policy. For example:

```go
//ftl:retry 5 1s catch recoverPaymentProcessing
func ProcessPayment(ctx context.Context, payment Payment) error {
...
}
```

### Catching

After all retries have failed, a catch verb can be used to safely recover.

These catch verbs have a request type of `builtin.CatchRequest<Req>` and no response type. If a catch verb returns an error, it will be retried until it succeeds so it is important to handle errors carefully.

```go
//ftl:retry 5 1s catch recoverPaymentProcessing
func ProcessPayment(ctx context.Context, payment Payment) error {
...
}

//ftl:verb
func RecoverPaymentProcessing(ctx context.Context, request builtin.CatchRequest[Payment]) error {
// safely handle final failure of the payment
}
```

</TabItem>
<TabItem value="kotlin" label="Kotlin">

The directive has the following syntax:

```kotlin
@Retry(attempts = 10, minBackoff = "5s", maxBackoff = "1h", catchVerb = "<catchVerb>", catchModule = "<catchModule>")
```

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```kotlin
@Retry(count = 10, minBackoff = "5s", maxBackoff = "1m")
fun process(inv: Invoice) {
    // ... 
}
```

### PubSub Subscribers

Subscribers can have a retry policy. For example:

```kotlin
@Subscription(topic = "example", name = "exampleSubscription")
@SubscriptionOptions(from = FromOffset.LATEST)
@Retry(count = 5, minBackoff = "1s", catchVerb = "recoverPaymentProcessing")
fun processPayment(payment: Payment) {
    // ... 
}
```

### Catching

After all retries have failed, a catch verb can be used to safely recover.

These catch verbs have a request type of `CatchRequest<Req>` and no response type. If a catch verb returns an error, it will be retried until it succeeds so it is important to handle errors carefully.

```kotlin
@Retry(count = 5, minBackoff = "1s", catchVerb = "recoverPaymentProcessing")
fun processPayment(payment: Payment) {
    // ... 
}

@Verb
fun recoverPaymentProcessing(req: CatchRequest<Payment>) {
    // safely handle final failure of the payment
}
```

</TabItem>
<TabItem value="java" label="Java">

The directive has the following syntax:

```java
@Retry(attempts = 10, minBackoff = "5s", maxBackoff = "1h", catchVerb = "<catchVerb>", catchModule = "<catchModule>")
```

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```java
@Retry(count = 10, minBackoff = "5s", maxBackoff = "1m")
public void process(Invoice in) {
    // ... 
}
```

### PubSub Subscribers

Subscribers can have a retry policy. For example:

```java
@Subscription(topic = "example", name = "exampleSubscription")
@SubscriptionOptions(from = FromOffset.LATEST)
@Retry(count = 5, minBackoff = "1s", catchVerb = "recoverPaymentProcessing")
public void processPayment(Payment payment) {
    // ... 
}
```

### Catching

After all retries have failed, a catch verb can be used to safely recover.

These catch verbs have a request type of `CatchRequest<Req>` and no response type. If a catch verb returns an error, it will be retried until it succeeds so it is important to handle errors carefully.

```java
@Retry(count = 5, minBackoff = "1s", catchVerb = "recoverPaymentProcessing")
public void processPayment(Payment payment) {
    // ... 
}

@Verb
public void recoverPaymentProcessing(CatchRequest<Payment> req) {
    // safely handle final failure of the payment
}
```

</TabItem>
</Tabs> 
