Declare a cron job, autocompleting on a cron expression.

A cron job is an Empty verb that will be called on a schedule. Supports both cron expressions and duration format.

```kotlin
@Cron("0 * * * *")
class Hourly {
    suspend fun run() {
        // Run every hour
    }
}

@Cron("6h")
class EverySixHours {
    suspend fun run() {
        // Run every 6 hours
    }
}

@Cron("Mon")
class Mondays {
    suspend fun run() {
        // Run every Monday
    }
}
```

See https://block.github.io/ftl/docs/reference/cron/
---

@Cron("${2:Minutes} ${3:Hours} ${4:DayOfMonth} ${5:Month} ${6:DayOfWeek}")
class ${1:Name} {
    suspend fun run() {
        ${7:// TODO: Implement}
    }
} 
