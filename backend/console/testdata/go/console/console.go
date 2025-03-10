package console

import (
	"context"
	"fmt"

	"ftl/builtin"

	lib "github.com/block/ftl/backend/console/testdata/go"
	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type External = lib.External

type Response struct {
	Message string
}

//ftl:ingress http GET /test
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, External]) (builtin.HttpResponse[Response, string], error) {
	return builtin.HttpResponse[Response, string]{
		Body: ftl.Some(Response{
			Message: fmt.Sprintf("Hello, %s", req.Query.Message),
		}),
	}, nil
}

//ftl:verb
func Verb(ctx context.Context, req External) (External, error) {
	return External{
		Message: fmt.Sprintf("Hello, %s", req.Message),
	}, nil
}
