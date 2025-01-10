Directive for retrying an async operation.

Any verb called asynchronously (specifically, PubSub subscribers and cron jobs) may specify a basic exponential backoff retry policy.

```go
//ftl:retry 10 5s 1m
func Process(ctx context.Context, in Invoice) error {
    // Process with retries
    return nil
}
```

See https://block.github.io/ftl/docs/reference/retries/
---

//ftl:retry ${1:10} ${2:5s} ${3:1m}
func ${4:Process}(ctx context.Context, in ${5:Type}) error {
	${6:// TODO: Implement}
	return nil
}
