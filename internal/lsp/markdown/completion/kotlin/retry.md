Directive for retrying an async operation.

Any verb called asynchronously (specifically, PubSub subscribers and cron jobs) may specify a basic exponential backoff retry policy. You can optionally specify a catch verb to handle final failures.

```kotlin
// Basic retry
@Retry(attempts = 10, minBackoff = "5s", maxBackoff = "1m")
fun processPayment(payment: Payment) {
	// Process with retries
}

// Retry with catch handler
@Retry(attempts = 5, minBackoff = "1s", catchVerb = "recoverPaymentProcessing")
fun processPayment(payment: Payment) {
	// Process with retries, failures will be sent to recoverPaymentProcessing verb
}

// The catch verb that handles final failures
@Verb
fun recoverPaymentProcessing(req: CatchRequest<Payment>) {
	// Safely handle final failure of the payment
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

@Retry(attempts = ${1:10}, minBackoff = "${2:5s}", maxBackoff = "${3:1m}"${4:, catchVerb = "${5:catchVerb}"})
fun ${6:process}(${7:input}: ${8:Type}) {
	${9:// TODO: Implement}
}

// Optional catch verb handler
@Verb
fun ${5:catchVerb}(req: CatchRequest<${8:Type}>) {
	${10:// Safely handle final failure}
} 
