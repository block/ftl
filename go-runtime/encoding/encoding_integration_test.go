//go:build integration

package encoding_test

import (
	"net/http"
	"testing"

	in "github.com/TBD54566975/ftl/integration"
	"github.com/alecthomas/assert/v2"
)

func TestHttpEncodeOmitempty(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("omitempty"),
		in.Deploy("omitempty"),
		in.HttpCall(http.MethodGet, "/get", in.JsonData(t, in.Obj{}), func(t testing.TB, resp *in.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			_, ok := resp.JsonBody["mustset"]
			assert.True(t, ok)
			_, ok = resp.JsonBody["error"]
			assert.False(t, ok)
		}),
	)
}
