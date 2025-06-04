package leases

import (
	"context"
	"time"

	"github.com/alecthomas/errors"

	leasepb "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
)

var _ Leaser = (*clientLeaser)(nil)
var _ Lease = (*clientLease)(nil)

func NewClientLeaser(ctx context.Context, client leasepbconnect.LeaseServiceClient) Leaser {
	return &clientLeaser{
		client: client,
	}
}

type clientLeaser struct {
	client leasepbconnect.LeaseServiceClient
}

func (c clientLeaser) AcquireLease(ctx context.Context, key Key, ttl time.Duration) (Lease, context.Context, error) {
	if len(key) == 0 {
		return nil, nil, errors.WithStack(errors.New("lease key must not be empty"))
	}
	if ttl.Seconds() < 5 {
		return nil, nil, errors.WithStack(errors.New("ttl must be at least 5 seconds"))
	}
	lease := c.client.AcquireLease(ctx)
	// Send the initial request to acquire the lease.
	err := lease.Send(&leasepb.AcquireLeaseRequest{
		Key: key,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to send acquire lease request")
	}
	_, err = lease.Receive()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to send receive lease response")
	}
	// We have got the lease, we need a goroutine to keep renewing the lease.
	ret := &clientLease{}
	ctx, cancel := context.WithCancelCause(ctx)
	done := func(err error) {
		cancel(err)
		_ = lease.CloseResponse() //nolint:errcheck
		_ = lease.CloseRequest()  //nolint:errcheck
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				done(ctx.Err())
				return
			case <-time.After(ttl / 2):
				err := lease.Send(&leasepb.AcquireLeaseRequest{
					Key: key,
				})
				if err != nil {
					done(errors.Wrap(errors.Join(context.Canceled, err), "failed to send acquire lease request"))
					return
				}
				_, err = lease.Receive()
				if err != nil {
					done(errors.Wrap(errors.Join(context.Canceled, err), "failed to receive lease response"))
					return
				}
			}

		}
	}()
	return ret, ctx, nil
}

type clientLease struct {
	done func()
}

func (c *clientLease) Release() error {
	c.done()
	return nil
}
