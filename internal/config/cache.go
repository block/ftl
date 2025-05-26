package config

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/log"
)

// CacheDecorator decorates asynchronous providers with a cache to provide synchronous access.
type CacheDecorator[R Role] struct {
	close         chan struct{}
	lock          sync.RWMutex
	cache         map[Ref][]byte
	asyncProvider AsynchronousProvider[R]
}

var _ Provider[Configuration] = &CacheDecorator[Configuration]{}

// NewCacheDecorator decorates asynchronous providers with a cache to provide synchronous access.
//
// The returned cache will block until the initial synchronization is complete, and will periodically refresh the cache
// based on the provider's refresh interval.
func NewCacheDecorator[R Role](ctx context.Context, provider AsynchronousProvider[R]) *CacheDecorator[R] {
	c := &CacheDecorator[R]{
		close:         make(chan struct{}),
		cache:         map[Ref][]byte{},
		asyncProvider: provider,
	}
	// Lock the cache before starting the goroutine so that we block until the initial synchronization is complete.
	c.lock.Lock()
	go c.run(ctx)
	return c
}

func (c *CacheDecorator[R]) Role() R          { return c.asyncProvider.Role() }
func (c *CacheDecorator[R]) Key() ProviderKey { return c.asyncProvider.Key() }

func (c *CacheDecorator[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, ok := c.cache[ref]
	if ok {
		return value, nil
	}
	return nil, errors.Wrap(ErrNotFound, "cache")
}

func (c *CacheDecorator[R]) Store(ctx context.Context, ref Ref, value []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.asyncProvider.Store(ctx, ref, value); err != nil {
		return errors.Wrap(err, "cache")
	}
	c.cache[ref] = value
	return nil
}

func (c *CacheDecorator[R]) Delete(ctx context.Context, ref Ref) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.asyncProvider.Delete(ctx, ref); err != nil {
		return errors.Wrap(err, "cache")
	}
	delete(c.cache, ref)
	return nil
}

func (c *CacheDecorator[R]) List(ctx context.Context, withValues bool) ([]Value, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	values := make([]Value, 0, len(c.cache))
	for key, value := range c.cache {
		values = append(values, Value{Ref: key, Value: optional.Some(value)})
	}
	slices.SortFunc(values, func(a, b Value) int {
		return strings.Compare(a.Ref.String(), b.Ref.String())
	})
	return values, nil
}

func (c *CacheDecorator[R]) Close(ctx context.Context) error {
	if err := c.asyncProvider.Close(ctx); err != nil {
		return errors.Wrap(err, "cache")
	}
	c.cache = map[Ref][]byte{}
	close(c.close)
	return nil
}

func (c *CacheDecorator[R]) run(ctx context.Context) {
	initialSync := true
	logger := log.FromContext(ctx).Scope(c.Role().String() + ":cache")
	for {
		next := c.asyncProvider.SyncInterval()
		values, err := c.asyncProvider.Sync(ctx)
		if err != nil {
			logger.Warnf("cache: sync failed, retrying in %s: %v", next, err)
		} else {
			logger.Debugf("cache: synced %d values", len(values))
			if !initialSync { // Locked on initial run.
				c.lock.Lock()
			}
			c.cache = map[Ref][]byte{}
			for ref, value := range values {
				c.cache[ref] = value.Value
			}
			// We only unlock if we have successfully synced.
			initialSync = false
			c.lock.Unlock()
		}
		select {
		case <-c.close:
			return
		case <-ctx.Done():
			return
		case <-time.After(next):
			continue
		}
	}
}
