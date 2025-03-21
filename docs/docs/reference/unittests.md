---
sidebar_position: 10
title: Unit Tests
description: Build unit tests for your modules
---

# Unit Tests

## Create a context

When writing a unit test, first create a context:
```go
func ExampleTest(t *testing.Test) {
    ctx := ftltest.Context(
        // options go here
    )
}
```

FTL will help isolate what you want to test by restricting access to FTL features by default. You can expand what is available to test by adding options to `ftltest.Context(...)`.

In this default set up, FTL does the following:
- prevents access to `ftl.ConfigValue` and `ftl.SecretValue` ([See options](#project-files-configs-and-secrets))
- prevents access to `ftl.Database` ([See options](#databases))
- prevents access to `ftl.MapHandle` ([See options](#maps))
- disables all subscribers ([See options](#pubsub))

## Customization

### Project files, configs and secrets

To enable configs and secrets from the default project file:
```go
ctx := ftltest.Context(
    ftltest.WithDefaultProjectFile(),
)
```

Or you can specify a specific project file:
```go
ctx := ftltest.Context(
    ftltest.WithProjectFile(path),
)
```

You can also override specific config and secret values:
```go
ctx := ftltest.Context(
    ftltest.WithDefaultProjectFile(),
    
    ftltest.WithConfig(endpoint, "test"),
    ftltest.WithSecret(secret, "..."),
)
```

### Databases

To enable database access in a test, you must first [provide a DSN via a project file](#project-files-configs-and-secrets). You can then set up a test database:
```go
ctx := ftltest.Context(
    ftltest.WithDefaultProjectFile(),
    ftltest.WithDatabase[MyDBConfig](),
)
```
This will:
- Take the provided DSN and appends `_test` to the database name. Eg: `accounts` becomes `accounts_test`
- Wipe all tables in the database so each test run happens on a clean database

You can access the database in your test using its handle:
```go
db, err := ftltest.GetDatabaseHandle[MyDBConfig]()
db.Get(ctx).Exec(...)
```

### Maps

By default, calling `Get(ctx)` on a map handle will panic.

You can inject a fake via a map:
```go
ctx := ftltest.Context(
    ftltest.WhenMap(exampleMap, func(ctx context.Context) (string, error) {
       return "Test Value"
    }),
)
```

You can also allow the use of all maps:
```go
ctx := ftltest.Context(
    ftltest.WithMapsAllowed(),
)
```

### Calls

Use `ftltest.Call[Client](...)` (or `ftltest.CallSource[Client](...)`, `ftltest.CallSink[Client](...)`, `ftltest.CallEmpty[Client](...)`) to invoke your 
verb. At runtime, FTL automatically provides these resources to your verb, and using `ftltest.Call(...)` rather than direct invocations simulates this behavior.

```go
// Call a verb
resp, err := ftltest.Call[ExampleVerbClient, Request, Response](ctx, Request{Param: "Test"})
```

You can inject fakes for verbs:
```go
ctx := ftltest.Context(
    ftltest.WhenVerb[ExampleVerbClient](func(ctx context.Context, req Request) (Response, error) {
       return Response{Result: "Lorem Ipsum"}, nil
    }),
)
```

If there is no request or response parameters, you can use `WhenSource(...)`, `WhenSink(...)`, or `WhenEmpty(...)`.

To enable all calls within a module:
```go
ctx := ftltest.Context(
    ftltest.WithCallsAllowedWithinModule(),
)
```

### PubSub

By default, all subscribers are disabled.
To enable a subscriber:
```go
ctx := ftltest.Context(
    ftltest.WithSubscriber(paymentsSubscription, ProcessPayment),
)
```

Or you can inject a fake subscriber:
```go
ctx := ftltest.Context(
    ftltest.WithSubscriber(paymentsSubscription, func (ctx context.Context, in PaymentEvent) error {
       return fmt.Errorf("failed payment: %v", in)
    }),
)
```

Due to the asynchronous nature of pubsub, your test should wait for subscriptions to consume the published events:
```go
topic.Publish(ctx, Event{Name: "Test"})

ftltest.WaitForSubscriptionsToComplete(ctx)
// Event will have been consumed by now
```

You can check what events were published to a topic:
```go
events := ftltest.EventsForTopic(ctx, topic)
```

You can check what events were consumed by a subscription, and whether a subscriber returned an error:
```go
results := ftltest.ResultsForSubscription(ctx, subscription)
```

If all you wanted to check was whether a subscriber returned an error, this function is simpler:
```go
errs := ftltest.ErrorsForSubscription(ctx, subscription)
```

PubSub also has these different behaviours while testing:
- Publishing to topics in other modules is allowed
- If a subscriber returns an error, no retries will occur regardless of retry policy. 
