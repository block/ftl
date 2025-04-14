//go:build integration

package pubsub_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/block/ftl/internal/exec"
	in "github.com/block/ftl/internal/integration"
)

func TestPubSub(t *testing.T) {
	in.Run(t,
		append(setupPubsubTests(), in.WithLanguages("java", "go", "kotlin"))...,
	)
}

func TestRetry(t *testing.T) {
	retriesPerCall := 2
	in.Run(t,
		in.WithLanguages("java", "go"),

		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher", "subscriber"),

		// publish events
		in.Call("publisher", "publishOneToTopic2", map[string]any{"haystack": "firstCall"}, func(t testing.TB, resp in.Obj) {}),
		in.Call("publisher", "publishOneToTopic2", map[string]any{"haystack": "secondCall"}, func(t testing.TB, resp in.Obj) {}),

		in.Sleep(time.Second*7),

		checkConsumed("subscriber", "consumeButFailAndRetry", false, retriesPerCall+1, optional.Some("firstCall")),
		checkConsumed("subscriber", "consumeButFailAndRetry", false, retriesPerCall+1, optional.Some("secondCall")),
		checkPublished("subscriber", "consumeButFailAndRetryFailed", 2),

		checkConsumed("subscriber", "consumeFromDeadLetter", true, 2, optional.None[string]()),
	)
}

func TestExternalPublishRuntimeCheck(t *testing.T) {
	// No java as there is no API for this
	in.Run(t,
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher", "subscriber"),

		in.ExpectError(
			in.Call("subscriber", "publishToExternalModule", in.Obj{}, func(t testing.TB, resp in.Obj) {}),
			"can not publish to another module's topic",
		),
	)
}

// TestConsumerGroupMembership tests that when a runner ends, the consumer group is properly exited.
func TestConsumerGroupMembership(t *testing.T) {
	var deploymentKilledTime *time.Time
	in.Run(t,
		in.WithLanguages("go"),
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher", "subscriber"),

		// consumer group must now have a member for each partition
		checkGroupMembership("subscriber", "consumeSlow", 1),

		// publish events that will take a long time to process on the first subscriber deployment
		// to test that rebalancing doesnt cause consumption to fail and skip events
		in.Repeat(100, in.Call("publisher", "publishSlow", in.Obj{}, func(t testing.TB, resp in.Obj) {})),

		// Upgrade deployment
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Modifying code")
			path := filepath.Join(ic.WorkingDir(), "subscriber", "subscriber.go")

			bytes, err := os.ReadFile(path)
			assert.NoError(t, err)
			output := strings.ReplaceAll(string(bytes), "This deployment is TheFirstDeployment", "This deployment is TheSecondDeployment")
			assert.NoError(t, os.WriteFile(path, []byte(output), 0644))
		},
		in.Deploy("subscriber"),

		// Currently old deployment runs for a little bit longer.
		// During this time we expect the consumer group to have 2 members (old deployment and new deployment).
		// This will probably change when we have proper draining of the old deployment.
		checkGroupMembership("subscriber", "consumeSlow", 2),
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Waiting for old deployment to be killed")
			start := time.Now()
			for {
				assert.True(t, time.Since(start) < 15*time.Second)
				ps, err := exec.Capture(ic.Context, ".", "ftl", "ps")
				assert.NoError(t, err)
				if strings.Count(string(ps), "dpl-default-subscriber-") == 1 {
					// original deployment has ended
					now := time.Now()
					deploymentKilledTime = &now
					return
				}
			}
		},
		// Once old deployment has ended, the consumer group should only have 1 member per partition (the new deployment)
		// This should happen fairly quickly. If it takes a while it could be because the previous deployment did not close
		// the group properly.
		checkGroupMembership("subscriber", "consumeSlow", 1),
		func(t testing.TB, ic in.TestContext) {
			assert.True(t, time.Since(*deploymentKilledTime) < 10*time.Second, "make sure old deployment was removed from consumer group fast enough")
		},

		// confirm that each message was consumed successfully
		checkConsumed("subscriber", "consumeSlow", true, 100, optional.None[string]()),
	)
}
