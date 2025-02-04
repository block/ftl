---
sidebar_position: 5
title: Visibility
description: Managing visibility of FTL declarations
---

# Visibility

By default all declarations in FTL are visible only to the module they're declared in. The implicit visibility of types is that of the first verb or other declaration that references it.

## Exporting declarations

Exporting a declaration makes it accessible to other modules. Some declarations that are entirely local to a module, such as secrets/config, cannot be exported.

Types that are transitively referenced by an exported declaration will be automatically exported unless they were already defined but unexported. In this case, an error will be raised and the type must be explicitly exported.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="go" label="Go" default>

The following table describes the go directives used to export the corresponding declaration:

| Symbol        | Export syntax            |
| ------------- | ------------------------ |
| Verb          | `//ftl:verb export`      |
| Data          | `//ftl:data export`      |
| Enum/Sum type | `//ftl:enum export`      |
| Typealias     | `//ftl:typealias export` |
| Topic         | `//ftl:export` [^1]      |

```go
//ftl:verb export
func Verb(ctx context.Context, in In) (Out, error)

//ftl:typealias export
type UserID string
```

[^1]: By default, topics do not require any annotations as the declaration itself is sufficient.

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

For Kotlin the `@Export` annotation can be used to export a declaration:

```kotlin
@Verb
@Export
fun time(): TimeResponse {
    // ...
}
```

  </TabItem>
  <TabItem value="java" label="Java">

For Java the `@Export` annotation can be used to export a declaration:

```java
@Verb
@Export
TimeResponse time()  {
    // ...
}
```

  </TabItem>
</Tabs> 
