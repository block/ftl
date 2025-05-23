package omitempty

import (
	"context"

	"ftl/builtin"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Response struct {
	Error   string `json:"error,omitempty"` // Should be omitted from marshaled JSON
	MustSet string `json:"mustset"`         // Should marshal to `"mustset":""`
}

//ftl:ingress http GET /get
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[Response, string], error) {
	return builtin.HttpResponse[Response, string]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body:    ftl.Some[Response](Response{}),
	}, nil
}
