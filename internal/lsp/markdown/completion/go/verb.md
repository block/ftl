Declare a Verb function.

A Verb is a remotely callable function that takes an input and returns an output. Verbs are the primary way to expose functionality in FTL. They are strongly typed and support both synchronous and asynchronous execution.

The structure of a verb is:
```go
//ftl:verb
func F(context.Context, In) (Out, error) { }
```

Here's an example of a verb:
```go
type EchoRequest struct {}

type EchoResponse struct {}

//ftl:verb
func Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {
	// Process the request
	return EchoResponse{}, nil
}
```

See https://block.github.io/ftl/docs/reference/verbs/
---

type ${1:Request} struct {}

type ${2:Response} struct {}

//ftl:verb
func ${3:Name}(ctx context.Context, in ${1:Request}) (${2:Response}, error) {
	${4:// TODO: Implement}
	return ${2:Response}{}, nil
}
