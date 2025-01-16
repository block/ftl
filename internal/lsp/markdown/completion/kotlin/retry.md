Directive for retrying an async operation.

Any verb called asynchronously (specifically, PubSub subscribers and cron jobs) may specify a basic exponential backoff retry policy.

```kotlin
@Retry(attempts = 10, minBackoff = "5s", maxBackoff = "1m")
class InvoiceProcessor {
    suspend fun process(ctx: Context, invoice: Invoice) {
        // Process with retries
    }
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

@Retry(attempts = ${1:10}, minBackoff = "${2:5s}", maxBackoff = "${3:1m}")
class ${4:Processor} {
    suspend fun ${5:process}(ctx: Context, input: ${6:Type}) {
        ${7:// TODO: Implement}
    }
} 
