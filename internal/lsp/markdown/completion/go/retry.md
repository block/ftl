Directive for retrying an async operation.

Any verb called asynchronously (specifically, PubSub subscribers and cron jobs) may specify a basic exponential backoff retry policy. You can optionally specify a catch verb to handle final failures.

```go
// Basic retry
//ftl:retry 10 5s 1m
func processPayment(ctx context.Context, payment Payment) error {
	// Process with retries
	return nil
}

// Retry with catch handler
//ftl:retry 5 1s catch recoverPaymentProcessing
func processPayment(ctx context.Context, payment Payment) error {
	// Process with retries, failures will be sent to recoverPaymentProcessing verb
	return nil
}

// The catch verb that handles final failures
//ftl:verb
func recoverPaymentProcessing(ctx context.Context, request builtin.CatchRequest[Payment]) error {
	// Safely handle final failure of the payment
	return nil
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

//ftl:retry ${1:attempts} ${2:minBackoff} ${3:maxBackoff}${4: catch ${5:catchVerb}}
func ${6:process}(ctx context.Context, in ${7:Type}) error {
	${8:// TODO: Implement}
	return nil
}

// Optional catch verb handler
//ftl:verb
func ${5:catchVerb}(ctx context.Context, request builtin.CatchRequest[${7:Type}]) error {
	${9:// Safely handle final failure}
	return nil
}
