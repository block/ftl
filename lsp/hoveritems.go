// Code generated by ftl-lsp-gen. DO NOT EDIT.
package lsp

var hoverMap = map[string]string{
	"//ftl:enum": "## Type enums (sum types)\n\n[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system, but aren't directly supported by Go. However they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/). To declare a sum type in FTL use the comment directive `//ftl:enum`:\n\n```go\n//ftl:enum\ntype Animal interface { animal() }\n\ntype Cat struct {}\nfunc (Cat) animal() {}\n\ntype Dog struct {}\nfunc (Dog) animal() {}\n```\n## Value enums\n\nA value enum is an enumerated set of string or integer values.\n\n```go\n//ftl:enum\ntype Colour string\n\nconst (\n  Red   Colour = \"red\"\n  Green Colour = \"green\"\n  Blue  Colour = \"blue\"\n)\n```\n",
	"//ftl:cron": "## Cron\n\n\n\nA cron job is an Empty verb that will be called on a schedule. The syntax is described [here](https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html).\n\nYou can also use a shorthand syntax for the cron job, supporting seconds (`s`), minutes (`m`), hours (`h`), and specific days of the week (e.g. `Mon`).\n\n### Examples\n\nThe following function will be called hourly:\n\n```go\n//ftl:cron 0 * * * *\nfunc Hourly(ctx context.Context) error {\n  // ...\n}\n```\n\nEvery 12 hours, starting at UTC midnight:\n\n```go\n//ftl:cron 12h\nfunc TwiceADay(ctx context.Context) error {\n  // ...\n}\n```\n\nEvery Monday at UTC midnight:\n\n```go\n//ftl:cron Mon\nfunc Mondays(ctx context.Context) error {\n  // ...\n}\n```\n\n",
	"//ftl:ingress": "## HTTP Ingress\n\n\n\nVerbs annotated with `ftl:ingress` will be exposed via HTTP (`http` is the default ingress type). These endpoints will then be available on one of our default `ingress` ports (local development defaults to `http://localhost:8891`).\n\nThe following will be available at `http://localhost:8891/http/users/123/posts?postId=456`.\n\n```go\ntype GetRequest struct {\n\tUserID string `json:\"userId\"`\n\tPostID string `json:\"postId\"`\n}\n\ntype GetResponse struct {\n\tMessage string `json:\"msg\"`\n}\n\n//ftl:ingress GET /http/users/{userId}/posts\nfunc Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {\n  // ...\n}\n```\n\n> **NOTE!**\n> The `req` and `resp` types of HTTP `ingress` [verbs](../verbs) must be `builtin.HttpRequest` and `builtin.HttpResponse` respectively. These types provide the necessary fields for HTTP `ingress` (`headers`, `statusCode`, etc.)\n> \n> You will need to import `ftl/builtin`.\n\nKey points to note\n\n- `path`, `query`, and `body` parameters are automatically mapped to the `req` and `resp` structures. In the example above, `{userId}` is extracted from the path parameter and `postId` is extracted from the query parameter.\n- `ingress` verbs will be automatically exported by default.\n",
	"//ftl:subscribe": "## PubSub\n\n\n\nFTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent), subscriptions (a cursor over the topic), and subscribers (functions events are delivered to). Subscribers are, as you would expect, sinks. Each subscription is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each subscription may have multiple subscribers, in which case events will be distributed among them.\n\nFirst, declare a new topic:\n\n```go\nvar invoicesTopic = ftl.Topic[Invoice](\"invoices\")\n```\n\nThen declare each subscription on the topic:\n\n```go\nvar _ = ftl.Subscription(invoicesTopic, \"emailInvoices\")\n```\n\nAnd finally define a Sink to consume from the subscription:\n\n```go\n//ftl:subscribe emailInvoices\nfunc SendInvoiceEmail(ctx context.Context, in Invoice) error {\n  // ...\n}\n```\n\nEvents can be published to a topic like so:\n\n```go\ninvoicesTopic.Publish(ctx, Invoice{...})\n```\n\n> **NOTE!**\n> PubSub topics cannot be published to from outside the module that declared them, they can only be subscribed to. That is, if a topic is declared in module `A`, module `B` cannot publish to it.\n",
	"//ftl:retry": "## Retries\n\n\n\nAny verb called asynchronously (specifically, PubSub subscribers and FSM states), may optionally specify a basic exponential backoff retry policy via a Go comment directive. The directive has the following syntax:\n\n```go\n//ftl:retry [<attempts>] <min-backoff> [<max-backoff>]\n```\n\n`attempts` and `max-backoff` default to unlimited if not specified.\n\nFor example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.\n\n```go\n//ftl:retry 10 5s 1m\nfunc Invoiced(ctx context.Context, in Invoice) error {\n  // ...\n}\n```\n",
	"//ftl:verb": "## Verbs\n\n\n\n## Defining Verbs\n\nTo declare a Verb, write a normal Go function with the following signature, annotated with the Go [comment directive](https://tip.golang.org/doc/comment#syntax) `//ftl:verb`:\n\n```go\n//ftl:verb\nfunc F(context.Context, In) (Out, error) { }\n```\n\neg.\n\n```go\ntype EchoRequest struct {}\n\ntype EchoResponse struct {}\n\n//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {\n  // ...\n}\n```\n\nBy default verbs are only [visible](../visibility) to other verbs in the same module.\n\n## Calling Verbs\n\nTo call a verb use `ftl.Call()`. eg.\n\n```go\nout, err := ftl.Call(ctx, echo.Echo, echo.EchoRequest{})\n```\n",
	"//ftl:typealias": "## Type aliases\n\nA type alias is an alternate name for an existing type. It can be declared like so:\n\n```go\n//ftl:typealias\ntype Alias Target\n```\n\neg.\n\n```go\n//ftl:typealias\ntype UserID string\n```\n",
}
