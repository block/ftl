Directive for retrying an async operation.

Any verb called asynchronously (specifically, PubSub subscribers and cron jobs) may specify a basic exponential backoff retry policy. You can optionally specify a catch verb to handle final failures.

```java
// Basic retry
@Retry(attempts = 10, minBackoff = "5s", maxBackoff = "1m")
public void processPayment(Payment payment) throws Exception {
	// Process with retries
}

// Retry with catch handler
@Retry(attempts = 5, minBackoff = "1s", catchVerb = "recoverPaymentProcessing")
public void processPayment(Payment payment) throws Exception {
	// Process with retries, failures will be sent to recoverPaymentProcessing verb
}

// The catch verb that handles final failures
@Verb
public void recoverPaymentProcessing(CatchRequest<Payment> req) {
	// Safely handle final failure of the payment
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

@Retry(attempts = ${1:10}, minBackoff = "${2:5s}", maxBackoff = "${3:1m}"${4:, catchVerb = "${5:catchVerb}"})
public void ${6:process}(${7:Type} input) throws Exception {
	${8:// TODO: Implement}
}

// Optional catch verb handler
@Verb
public void ${5:catchVerb}(CatchRequest<${7:Type}> req) {
	${9:// Safely handle final failure}
} 
