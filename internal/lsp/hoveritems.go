// Code generated by 'just lsp-generate'. DO NOT EDIT.
package lsp

var hoverMap = map[string]string{
	"//ftl:cron": "## Cron\n\nA cron job is an Empty verb that will be called on a schedule. The syntax is described [here](https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html).\n\nYou can also use a shorthand syntax for the cron job, supporting seconds (`s`), minutes (`m`), hours (`h`), and specific days of the week (e.g. `Mon`).\n\n### Examples\n\nThe following function will be called hourly:\n\n```go\n//ftl:cron 0 * * * *\nfunc Hourly(ctx context.Context) error {\n  // ...\n}\n```\nEvery 12 hours, starting at UTC midnight:\n\n```go\n//ftl:cron 12h\nfunc TwiceADay(ctx context.Context) error {\n  // ...\n}\n```\n\nEvery Monday at UTC midnight:\n\n```go\n//ftl:cron Mon\nfunc Mondays(ctx context.Context) error {\n  // ...\n}\n```",
	"//ftl:enum": "## Type enums (sum types)\n\n[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system, but aren't directly supported by Go. However they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/). To declare a sum type in FTL use the comment directive `//ftl:enum`:\n\n```go\n//ftl:enum\ntype Animal interface { animal() }\n\ntype Cat struct {}\nfunc (Cat) animal() {}\n\ntype Dog struct {}\nfunc (Dog) animal() {}\n```\n## Value enums\n\nA value enum is an enumerated set of string or integer values.\n\n```go\n//ftl:enum\ntype Colour string\n\nconst (\n  Red   Colour = \"red\"\n  Green Colour = \"green\"\n  Blue  Colour = \"blue\"\n)\n```\n",
	"//ftl:ingress": "## HTTP Ingress\n\nVerbs annotated with `ftl:ingress` will be exposed via HTTP (`http` is the default ingress type). These endpoints will then be available on one of our default `ingress` ports (local development defaults to `http://localhost:8891`).\n\nThe following will be available at `http://localhost:8891/http/users/123/posts?postId=456`.\n\n\n```go\ntype GetRequestPathParams struct {\n\tUserID string `json:\"userId\"`\n}\n\ntype GetRequestQueryParams struct {\n\tPostID string `json:\"postId\"`\n}\n\ntype GetResponse struct {\n\tMessage string `json:\"msg\"`\n}\n\n//ftl:ingress GET /http/users/{userId}/posts\nfunc Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetRequestPathParams, GetRequestQueryParams]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {\n  // ...\n}\n```\n\nBecause the example above only has a single path parameter it can be simplified by just using a scalar such as `string` or `int64` as the path parameter type:\n\n```go\n\n//ftl:ingress GET /http/users/{userId}/posts\nfunc Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, int64, GetRequestQueryParams]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {\n  // ...\n}\n```\n\n> **NOTE!**\n> The `req` and `resp` types of HTTP `ingress` [verbs](../verbs) must be `builtin.HttpRequest` and `builtin.HttpResponse` respectively. These types provide the necessary fields for HTTP `ingress` (`headers`, `statusCode`, etc.)\n>\n> You will need to import `ftl/builtin`.\n\nKey points:\n\n- `ingress` verbs will be automatically exported by default.\n\n## Field mapping\n\nThe `HttpRequest` request object takes 3 type parameters, the body, the path parameters and the query parameters.\n\nGiven the following request verb:\n\n```go\n\ntype PostBody struct{\n\tTitle string               `json:\"title\"`\n\tContent string             `json:\"content\"`\n\tTag ftl.Option[string]     `json:\"tag\"`\n}\ntype PostPathParams struct {\n\tUserID string             `json:\"userId\"`\n\tPostID string             `json:\"postId\"`\n}\n\ntype PostQueryParams struct {\n\tPublish boolean `json:\"publish\"`\n}\n\n//ftl:ingress http PUT /users/{userId}/posts/{postId}\nfunc Get(ctx context.Context, req builtin.HttpRequest[PostBody, PostPathParams, PostQueryParams]) (builtin.HttpResponse[GetResponse, string], error) {\n\treturn builtin.HttpResponse[GetResponse, string]{\n\t\tHeaders: map[string][]string{\"Get\": {\"Header from FTL\"}},\n\t\tBody: ftl.Some(GetResponse{\n\t\t\tMessage: fmt.Sprintf(\"UserID: %s, PostID: %s, Tag: %s\", req.pathParameters.UserID, req.pathParameters.PostID, req.Body.Tag.Default(\"none\")),\n\t\t}),\n\t}, nil\n}\n```\n\nThe rules for how each element is mapped are slightly different, as they have a different structure:\n\n- The body is mapped directly to the body of the request, generally as a JSON object. Scalars are also supported, as well as []byte to get the raw body. If they type is `any` then it will be assumed to be JSON and mapped to the appropriate types based on the JSON structure.\n- The path parameters can be mapped directly to an object with field names corresponding to the name of the path parameter. If there is only a single path parameter it can be injected directly as a scalar. They can also be injected as a `map[string]string`.\n- The path parameters can also be mapped directly to an object with field names corresponding to the name of the path parameter. They can also be injected directly as a `map[string]string`, or `map[string][]string` for multiple values.\n\n#### Optional fields\n\nOptional fields are represented by the `ftl.Option` type. The `Option` type is a wrapper around the actual type and can be `Some` or `None`. In the example above, the `Tag` field is optional.\n\n```sh\ncurl -i http://localhost:8891/users/123/posts/456\n```\n\nBecause the `tag` query parameter is not provided, the response will be:\n\n```json\n{\n  \"msg\": \"UserID: 123, PostID: 456, Tag: none\"\n}\n```\n\n#### Casing\n\nField names use lowerCamelCase by default. You can override this by using the `json` tag.\n\n## SumTypes\n\nGiven the following request verb:\n\n```go\n//ftl:enum export\ntype SumType interface {\n\ttag()\n}\n\ntype A string\n\nfunc (A) tag() {}\n\ntype B []string\n\nfunc (B) tag() {}\n\n//ftl:ingress http POST /typeenum\nfunc TypeEnum(ctx context.Context, req builtin.HttpRequest[SumType, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[SumType, string], error) {\n\treturn builtin.HttpResponse[SumType, string]{Body: ftl.Some(req.Body)}, nil\n}\n```\n\nThe following curl request will map the `SumType` name and value to the `req.Body`:\n\n```sh\ncurl -X POST \"http://localhost:8891/typeenum\" \\\n     -H \"Content-Type: application/json\" \\\n     --data '{\"name\": \"A\", \"value\": \"sample\"}'\n```\n\nThe response will be:\n\n```json\n{\n  \"name\": \"A\",\n  \"value\": \"sample\"\n}\n```\n\n## Encoding query params as JSON\n\nComplex query params can also be encoded as JSON using the `@json` query parameter. For example:\n\n> `{\"tag\":\"ftl\"}` url-encoded is `%7B%22tag%22%3A%22ftl%22%7D`\n\n```bash\ncurl -i http://localhost:8891/users/123/posts/456?@json=%7B%22tag%22%3A%22ftl%22%7D\n```\n\n\n\n",
	"//ftl:retry": "## Retries\n\nSome FTL features allow specifying a retry policy via a Go comment directive. Retries back off exponentially until the maximum is reached.\n\nThe directive has the following syntax:\n\n\n```go\n//ftl:retry [<attempts=10>] <min-backoff> [<max-backoff=1hr>] [catch <catchVerb>]\n```\n\n\nFor example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.\n\n\n```go\n//ftl:retry 10 5s 1m\nfunc Process(ctx context.Context, in Invoice) error {\n  // ...\n}\n```\n\n### PubSub\n\nSubscribers can have a retry policy. For example:\n\n\n```go\n//ftl:retry 5 1s catch recoverPaymentProcessing\nfunc ProcessPayment(ctx context.Context, payment Payment) error {\n...\n}\n```\n\n\n## Catching\nAfter all retries have failed, a catch verb can be used to safely recover.\n\nThese catch verbs have a request type of `builtin.CatchRequest<Req>` and no response type. If a catch verb returns an error, it will be retried until it succeeds so it is important to handle errors carefully.\n\n\n\n```go\n//ftl:retry 5 1s catch recoverPaymentProcessing\nfunc ProcessPayment(ctx context.Context, payment Payment) error {\n...\n}\n\n//ftl:verb\nfunc RecoverPaymentProcessing(ctx context.Context, request builtin.CatchRequest[Payment]) error {\n// safely handle final failure of the payment\n}\n```\n",
	"//ftl:subscribe": "## PubSub\n\nFTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent) and subscribers (a verb which consumes events). Subscribers are, as you would expect, sinks. Each subscriber is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each published event has an at least once delivery guarantee for each subscription.\n\n\nFirst, declare a new topic:\n\n```go\npackage payments\n\nimport (\n  \"github.com/block/ftl/go-runtime/ftl\"\n)\ntype Invoice struct {\n  InvoiceNo string\n}\n\n//ftl:export\ntype Invoices = ftl.TopicHandle[Invoice, ftl.SinglePartitionMap[Invoice]]\n```\n\nNote that the name of the topic as represented in the FTL schema is the lower camel case version of the type name.\n\nThe `Invoices` type is a handle to the topic. It is a generic type that takes two arguments: the event type and the partition map type. The partition map type is used to map events to partitions. In this case, we are using a single partition map, which means that all events are sent to the same partition.\n\nThen define a Sink to consume from the topic:\n\n```go\n//ftl:subscribe payments.invoices from=beginning\nfunc SendInvoiceEmail(ctx context.Context, in Invoice) error {\n  // ...\n}\n```\n\nEvents can be published to a topic by injecting the topic type into a verb:\n\n```go\n//ftl:verb\nfunc PublishInvoice(ctx context.Context, topic Invoices) error {\n   topic.Publish(ctx, Invoice{...})\n   // ...\n}\n```\n\n> **NOTE!**\n> PubSub topics cannot be published to from outside the module that declared them, they can only be subscribed to. That is, if a topic is declared in module `A`, module `B` cannot publish to it.\n",
	"//ftl:typealias": "## Type aliases\n\nA type alias is an alternate name for an existing type. It can be declared like so:\n\n```go\n//ftl:typealias\ntype Alias Target\n```\nor\n```go\n//ftl:typealias\ntype Alias = Target\n```\n\neg.\n\n```go\n//ftl:typealias\ntype UserID string\n\n//ftl:typealias\ntype UserToken = string\n```\n",
	"//ftl:verb": "## Verbs\n\n## Defining Verbs\n\n\nTo declare a Verb, write a normal Go function with the following signature, annotated with the Go [comment directive](https://tip.golang.org/doc/comment#syntax) `//ftl:verb`:\n\n```go\n//ftl:verb\nfunc F(context.Context, In) (Out, error) { }\n```\n\neg.\n\n```go\ntype EchoRequest struct {}\n\ntype EchoResponse struct {}\n\n//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {\n  // ...\n}\n```\n\n\nBy default verbs are only [visible](../visibility) to other verbs in the same module.\n\n## Calling Verbs\n\n\nTo call a verb, import the module's verb client (`{ModuleName}.{VerbName}Client`), add it to your verb's signature, then invoke it as a function. eg.\n\n```go\n//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest, tc time.TimeClient) (EchoResponse, error) {\n\tout, err := tc(ctx, TimeRequest{...})\n}\n```\n\nVerb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must build the \nmodule first (with callee verb defined) in order to generate its client for use by the caller. Local verb clients are \navailable in the generated `types.ftl.go` file as `{VerbName}Client`.\n\n",
}
