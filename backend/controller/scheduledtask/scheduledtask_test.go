package scheduledtask

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"

	"github.com/block/ftl/backend/controller/leases"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

func TestScheduledTask(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancelCause(ctx)
	t.Cleanup(func() { cancel(fmt.Errorf("test ended: %w", context.Canceled)) })

	var singletonCount atomic.Int64
	var multiCount atomic.Int64

	type controller struct {
		controllerKey key.Controller
		cron          *Scheduler
	}

	controllers := []*controller{
		{controllerKey: key.NewControllerKey("localhost", "8080")},
		{controllerKey: key.NewControllerKey("localhost", "8081")},
		{controllerKey: key.NewControllerKey("localhost", "8082")},
		{controllerKey: key.NewControllerKey("localhost", "8083")},
	}

	clock := clock.NewMock()
	leaser := leases.NewFakeLeaser()

	for _, c := range controllers {
		c.cron = NewForTesting(ctx, c.controllerKey, clock, leaser)
		c.cron.Singleton("singleton", backoff.Backoff{}, func(ctx context.Context) (time.Duration, error) {
			singletonCount.Add(1)
			return time.Second, nil
		})
		c.cron.Parallel("parallel", backoff.Backoff{}, func(ctx context.Context) (time.Duration, error) {
			multiCount.Add(1)
			return time.Second, nil
		})
	}

	for range 6 {
		clock.Add(time.Second)
		time.Sleep(time.Millisecond * 100)
	}

	assert.True(t, singletonCount.Load() >= 5 && singletonCount.Load() < 10, "expected singletonCount to be >= 5 but was %d", singletonCount.Load())
	assert.True(t, multiCount.Load() >= 20 && multiCount.Load() < 30, "expected multiCount to be >= 20 but was %d", multiCount.Load())
}
