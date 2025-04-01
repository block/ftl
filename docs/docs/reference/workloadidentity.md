---
sidebar_position: 18
title: Workload Identity
description: Workload Identity
---

# Workload Identity

FTL relies on Istio to mTLS to authenticate services, and provide the workload identity of a caller of the service. FTL exposes
this as a Spiffe ID for each request. This is useful for services that need to know the identity of the caller. This is achieved
by FTL parsing the `x-forwarded-client-cert` header that is set by Istio.

When running in local `dev` or `serve` mode FTL injected a fake Spiffe ID for the service. This is of the form `spiffe://cluster.local/ns/<module>/sa/<module>`.

## Examples

Examples of retrieving the Spiffe ID in different languages:

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
func CallerIdentity(ctx context.Context) (string, error) {
identity := ftl.WorkloadIdentity(ctx)
id, err := identity.SpiffeID()
if err != nil {
return "", err
}
return id.String(), nil
}

```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
import xyz.block.ftl.WorkloadIdentity

@Verb
@Export
fun callerIdentity(identity: WorkloadIdentity): String {
    return identity.spiffeID().toString()
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
import xyz.block.ftl.WorkloadIdentity;

@Verb
@Export
public String callerIdentity(WorkloadIdentity identity) {
    return identity.spiffeID().toStrin();
}
```

  </TabItem>
</Tabs>
