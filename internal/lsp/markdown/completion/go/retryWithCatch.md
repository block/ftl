Directive for retrying an async operation with a catch handler.

Specify a catch handler verb to safely handle final failures after all retries are exhausted.

```go
//ftl:retry 5 1s catch recoverPaymentProcessing
func ProcessPayment(ctx context.Context, payment Payment) error {
    // Process payment with retries
    return nil
}

//ftl:verb
func RecoverPaymentProcessing(ctx context.Context, request builtin.CatchRequest[Payment]) error {
    // Safely handle final failure of the payment
    return nil
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

//ftl:retry ${1:5} ${2:1s} catch ${3:RecoverProcessing}
func ${4:Process}(ctx context.Context, in ${5:Type}) error {
	${6:// TODO: Implement}
	return nil
}

//ftl:verb
func ${3:RecoverProcessing}(ctx context.Context, request builtin.CatchRequest[${5:Type}]) error {
	${7:// Handle final failure}
	return nil
} 
