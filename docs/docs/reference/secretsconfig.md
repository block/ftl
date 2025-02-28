---
sidebar_position: 13
title: Secrets/Config
description: Secrets and Configuration values
---

# Secrets and Configuration

## Configuration

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
<TabItem value="go" label="Go" default>

To declare a configuration value use the following syntax:

```go
// Simple string configuration
type ApiUrl = ftl.Config[string]

// Type-safe configuration
type DefaultUser = ftl.Config[Username]
```

Note that the name of the configuration value as represented in the FTL schema is the lower camel case version of the type name (e.g., `ApiUrl` becomes `apiUrl`).

Configuration values can be injected into FTL methods, such as //ftl:verb, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```go
//ftl:verb
func Hello(ctx context.Context, req Request, defaultUser DefaultUser) error {
    username := defaultUser.Get(ctx)
    // ...
}
```

</TabItem>
<TabItem value="kotlin" label="Kotlin">

Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```kotlin
@Export
@Verb
fun hello(helloRequest: HelloRequest, @Config("defaultUser") defaultUser: String): HelloResponse {
    return HelloResponse("Hello, $defaultUser")
}
```

</TabItem>
<TabItem value="java" label="Java">

Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```java
@Export
@Verb
HelloResponse hello(HelloRequest helloRequest, @Config("defaultUser") String defaultUser)  {
    return new HelloResponse("Hello, " + defaultUser);
}
```

</TabItem>
<TabItem value="schema" label="Schema">

In the FTL schema, configuration values are declared as follows:

```schema
module example {
  config defaultUser String
  
  verb hello(Unit) String
    +config example.defaultUser
}
```

Configuration values have a name, a type, and can be injected into verbs using the `+config` annotation.
</TabItem>
</Tabs>

## Secrets

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line.

<Tabs groupId="languages">
<TabItem value="go" label="Go" default>

Declare a secret with the following:

```go
// Simple string secret
type ApiToken = ftl.Secret[string]

// Type-safe secret
type ApiKey = ftl.Secret[Credentials]
```

Like configuration values, the name of the secret as represented in the FTL schema is the lower camel case version of the type name (e.g., `ApiToken` becomes `apiToken`).

Secrets can be injected into FTL methods, such as //ftl:verb, HTTP ingress, Cron etc. To inject a secret value, use the following syntax:

```go
//ftl:verb
func CallApi(ctx context.Context, req Request, apiKey ApiKey) error {
    credentials := apiKey.Get(ctx)
    // ...
}
```

</TabItem>
<TabItem value="kotlin" label="Kotlin">

Secrets can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a secret value, use the following syntax:

```kotlin
@Export
@Verb
fun hello(helloRequest: HelloRequest, @Secret("apiKey") apiKey: String): HelloResponse {
    return HelloResponse("Hello, ${api.call(apiKey)}")
}
```

</TabItem>
<TabItem value="java" label="Java">

Secrets can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a secret value, use the following syntax:

```java
@Export
@Verb
HelloResponse hello(HelloRequest helloRequest, @Secret("apiKey") String apiKey)  {
    return new HelloResponse("Hello, " + api.call(apiKey));
}
```

</TabItem>
<TabItem value="schema" label="Schema">

In the FTL schema, secrets are declared as follows:

```schema
module example {
  // Secret declaration
  secret apiToken String
  secret apiKey example.Credentials
  
  // Using a secret in a verb
  verb callApi(example.Request) Unit
    +secret apiKey
}
```

Secrets have a name, a type, and can be injected into verbs using the `+secret` annotation.
</TabItem>
</Tabs>

## Transforming secrets/configuration

Often, raw secret/configuration values aren't directly useful. For example, raw credentials might be used to create an API client. For those situations `ftl.Map()` can be used to transform a configuration or secret value into another type:

```go
var client = ftl.Map(ftl.Secret[Credentials]("credentials"),
                     func(ctx context.Context, creds Credentials) (*api.Client, error) {
    return api.NewClient(creds)
})
```

This is not currently supported in Kotlin or Java. 
