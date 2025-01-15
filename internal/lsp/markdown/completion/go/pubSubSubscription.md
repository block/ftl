Declare a subscription to a topic.

A subscription consumes events from a topic, starting either from the beginning of the topic or only new events.

```go
//ftl:subscribe payments.invoices from=beginning
func SendInvoiceEmail(ctx context.Context, in Invoice) error {
    // Process the invoice event
    return nil
}
```

See https://block.github.io/ftl/docs/reference/pubsub/
---

//ftl:subscribe ${1:topicName} from=${2:beginning}
func ${3:OnEvent}(ctx context.Context, in ${4:EventType}) error {
	${5:// TODO: Implement}
	return nil
}

