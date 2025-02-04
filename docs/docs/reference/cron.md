---
sidebar_position: 6
title: Cron
description: Cron Jobs
---

# Cron

A cron job is an Empty verb that will be called on a schedule. The syntax is described [here](https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html).

You can also use a shorthand syntax for the cron job, supporting seconds (`s`), minutes (`m`), hours (`h`), and specific days of the week (e.g. `Mon`).

## Examples

The following function will be called hourly:

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="go" label="Go" default>

```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {
  // ...
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
import xyz.block.ftl.Cron

@Cron("0 * * * *")
fun hourly() {
    
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
import xyz.block.ftl.Cron;

class MyCron {
    @Cron("0 * * * *")
    void hourly() {
        
    }
}
```

  </TabItem>
</Tabs>

Every 12 hours, starting at UTC midnight:

<Tabs>
  <TabItem value="go" label="Go" default>

```go
//ftl:cron 12h
func TwiceADay(ctx context.Context) error {
  // ...
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
import xyz.block.ftl.Cron

@Cron("12h")
fun twiceADay() {
    
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
import xyz.block.ftl.Cron;

class MyCron {
    @Cron("12h")
    void twiceADay() {
        
    }
}
```

  </TabItem>
</Tabs>

Every Monday at UTC midnight:

<Tabs>
  <TabItem value="go" label="Go" default>

```go
//ftl:cron Mon
func Mondays(ctx context.Context) error {
  // ...
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
import xyz.block.ftl.Cron

@Cron("Mon")
fun mondays() {
    
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
import xyz.block.ftl.Cron;

class MyCron {
    @Cron("Mon")
    void mondays() {
        
    }
}
```

  </TabItem>
</Tabs> 
