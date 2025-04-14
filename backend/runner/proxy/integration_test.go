//go:build integration

package proxy_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestWorkloadIdentity(t *testing.T) {
	in.Run(t,
		in.WithLanguages("kotlin", "go"),

		in.WithPubSub(),
		in.CopyModule("proxy"),
		in.CopyModule("echo"),
		in.Deploy("proxy", "echo"),

		// publish events
		in.Call("proxy", "proxy", "test", func(t testing.TB, string2 string) {
			assert.Equal(t, "test spiffe://cluster.local/ns/proxy/sa/proxy", string2)
		}))
}
