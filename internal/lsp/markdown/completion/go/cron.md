Declare a cron job.

A cron job is an Empty verb that will be called on a schedule. Supports both cron expressions and duration format.

```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {}

//ftl:cron 6h
func EverySixHours(ctx context.Context) error {}
```

See https://block.github.io/ftl/docs/reference/cron/
---

//ftl:cron ${2:Schedule}
func ${1:Name}(ctx context.Context) error {
	${3:// TODO: Implement}
	return nil
}
