---
sidebar_position: 19
title: Egress
description: Egress
---

# Egress

FTL requires verbs to declare any outbound egress they wish to do. At present no enforcement is done, however eventually
FTL will use this to provision appropriate Istio policies when running on kube. 

## Examples

Examples of defining egress targets in different languages:

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
package example

import (
"context"

"github.com/block/ftl/go-runtime/ftl"
)
//ftl:verb export
//ftl:egress url="https://example.com"
func Fixed(ctx context.Context, url ftl.EgressTarget) (string, error) {
 targetUrl := url.GetString(ctx)  // Equal https://example.com
 // Do HTTP requests with targetUrl
}

// For now in go you must explicitly declare config items used in egress urls
type Url = flt.Config[string]

//ftl:verb export
//ftl:egress url="${url}"
func FromConfig(ctx context.Context, url ftl.EgressTarget) (string, error) {
  // This automatically creates a config item called url in the schema
	targetUrl := url.GetString(ctx) // Equal to the value of the URL config item
	// Do HTTP requests with targetUrl
}

```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
import xyz.block.ftl.Egress
import xyz.block.ftl.Verb
import java.net.URL

@Verb
fun fixed(@Egress("https//example.com") url: String) {
    // do HTTP requests with url
}

@Verb
fun fromConfig(@Egress("${url}") url: String) {
    // This automatically creates a config item called url in the schema
}

@Verb
fun fromConfigAndURLType(@Egress("http://${host}:${port}") url: URL) {
    // This automatically creates a config item called host and port in the schema
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
import xyz.block.ftl.Egress;
import xyz.block.ftl.Verb;
import java.net.URL;

@Verb
public void fixed(String @Egress("https//example.com") url) {
    // do HTTP requests with url
}

@Verb
public void fromConfig(String @Egress("${url}") url) {
    // This automatically creates a config item called url in the schema
}

@Verb
public void fromConfigAndURLType(URL @Egress("http://${host}:${port}") url) {
    // This automatically creates a config item called host and port in the schema
}
```

  </TabItem>
</Tabs>
