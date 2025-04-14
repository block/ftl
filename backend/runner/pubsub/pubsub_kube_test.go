//go:build infrastructure

package pubsub_test

import (
	"testing"

	in "github.com/block/ftl/internal/integration"
)

func TestPubSubOnKube(t *testing.T) {
	in.Run(t,
		append(setupPubsubTests(), in.WithKubernetes())...,
	)
}
