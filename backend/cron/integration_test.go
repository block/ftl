//go:build infrastructure

package cron_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	in "github.com/block/ftl/internal/integration"
)

func TestDistributedCron(t *testing.T) {
	in.Run(t,
		in.WithKubernetes(),
		in.WithLanguages("go"),
		in.CopyModule("cron"),
		in.Deploy("cron"),

		// Wait for cron to execute at least once
		in.Sleep(2*time.Second),

		in.ExecWithOutput("kubectl", []string{"logs", "-l", "ftl.dev/module=cron"}, func(output string) {
			assert.Contains(t, output, "Cron job executed")
		}),
	)
}
