---
sidebar_position: 6
title: Fixtures
description: Fixtures
---

# Fixtures

Fixtures are a way to define a set of data that can be pre-populated for dev mode. Fixtures are essentially a verb that is called when the service is started in dev mode.

When not running in dev mode fixtures are not called, and are not present in the schema. This is to prevent fixtures from being called in production. This may change in the future
to allow fixtures to be used in contract testing between services.

FTL also supports manual fixtures that can be called by the user through the console or CLI. This is useful for defining dev mode helper functions that are not usable in production.

Note: Test fixtures are not implemented yet, this will be implemented in the future.

## Examples

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
//ftl:fixture
func Fixture(ctx context.Context) error {
  // ...
}
//ftl:fixture manual
func ManualFixture(ctx context.Context) error {
// ...
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
import xyz.block.ftl.Fixture

@Fixture
fun fixture() {
    
}
@Fixture(manual=true)
fun manualFixture() {

}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
import xyz.block.ftl.Fixture;

class MyFixture {
    @Fixture
    void fixture() {
        
    }
    @Fixture(manual=true)
    void manualFixture() {

    }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

```schema
module example {
  verb fixture(Unit) Unit
    +fixture
  verb manualFixture(Unit) Unit
    +fixture manual
}
```

  </TabItem>
</Tabs>
