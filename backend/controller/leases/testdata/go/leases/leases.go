package leases

import (
	"context"
	"time"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/go-runtime/ftl"
)

// Acquire acquires a lease and waits 5s before releasing it.
//
//ftl:verb
func Acquire(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Acquiring lease")
	lease, err := ftl.Lease(ctx, 10*time.Second, "lease")
	if err != nil {
		logger.Warnf("Failed to acquire lease: %s", err)
		return errors.Wrap(err, "failed to acquire lease")
	}
	logger.Infof("Acquired lease!")
	time.Sleep(time.Second * 5)
	return lease.Release()
}
