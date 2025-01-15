Directive for retrying an async operation.

Any verb called asynchronously (specifically, PubSub subscribers and cron jobs) may specify a basic exponential backoff retry policy.

```java
@Retry(attempts = 10, minBackoff = "5s", maxBackoff = "1m")
public class InvoiceProcessor {
    public void process(Context ctx, Invoice invoice) throws Exception {
        // Process with retries
    }
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

@Retry(attempts = ${1:10}, minBackoff = "${2:5s}", maxBackoff = "${3:1m}")
public class ${4:Processor} {
    public void ${5:process}(Context ctx, ${6:Type} input) throws Exception {
        ${7:// TODO: Implement}
    }
} 
