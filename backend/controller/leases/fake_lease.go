package leases

import (
	"context"
	"time"

	"github.com/alecthomas/errors"
	"github.com/puzpuzpuz/xsync/v3"
)

func NewFakeLeaser() *FakeLeaser {
	return &FakeLeaser{
		leases: xsync.NewMapOf[string, *FakeLease](),
	}
}

var _ Leaser = (*FakeLeaser)(nil)

// FakeLeaser is a fake implementation of the [Leaser] interface.
type FakeLeaser struct {
	leases *xsync.MapOf[string, *FakeLease]
}

func (f *FakeLeaser) AcquireLease(ctx context.Context, key Key, ttl time.Duration) (Lease, context.Context, error) {
	leaseCtx, cancelCtx := context.WithCancelCause(ctx)
	newLease := &FakeLease{
		leaser:    f,
		key:       key,
		cancelCtx: cancelCtx,
		ttl:       ttl,
	}
	if _, loaded := f.leases.LoadOrStore(key.String(), newLease); loaded {
		cancelCtx(errors.Wrapf(ErrConflict, "lease with key %q already exists", key))
		return nil, nil, errors.WithStack(ErrConflict)
	}
	return newLease, leaseCtx, nil
}

type FakeLease struct {
	leaser    *FakeLeaser
	key       Key
	cancelCtx context.CancelCauseFunc
	ttl       time.Duration
}

func (f *FakeLease) Release() error {
	f.leaser.leases.Delete(f.key.String())
	f.cancelCtx(errors.Errorf("lease with key %q released", f.key))
	return nil
}

func (f *FakeLease) String() string { return f.key.String() }
