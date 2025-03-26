//go:build infrastructure

package pubsub_test

import (
	in "github.com/block/ftl/internal/integration"
	"testing"
)

func TestPubSubOnKube(t *testing.T) {
	in.Run(t,
		append(setupPubsubTests(), in.WithKubernetes())...,
	)
}
