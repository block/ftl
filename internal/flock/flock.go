package flock

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alecthomas/errors"
	"github.com/jpillora/backoff"
	"golang.org/x/sys/unix"

	"github.com/block/ftl/common/log"
)

var ErrLocked = errors.New("locked")

// Acquire a lock on the given path.
//
// The lock is released when the returned function is called.
func Acquire(ctx context.Context, path string, timeout time.Duration) (release func() error, err error) {
	logger := log.FromContext(ctx)
	logger.Tracef("Acquiring lock %s", path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	retry := backoff.Backoff{Max: time.Second * 1, Jitter: true}
	end := time.Now().Add(timeout)
	for {
		release, err := acquire(absPath)
		if err == nil {
			logger.Tracef("Locked %s", path)
			return func() error {
				logger.Tracef("Releasing lock %s", path)
				return errors.WithStack(release())
			}, nil
		}
		if !errors.Is(err, ErrLocked) {
			return nil, errors.Wrapf(err, "failed to acquire lock %s", absPath)
		}
		if time.Now().After(end) {
			pid, _ := os.ReadFile(absPath) //nolint:errcheck
			return nil, errors.Wrapf(err, "timed out acquiring lock %s, locked by pid %s", absPath, pid)
		}
		select {
		case <-ctx.Done():
			return nil, errors.WithStack(ctx.Err())
		case <-time.After(retry.Duration()):
		}
	}
}

func acquire(path string) (release func() error, err error) {
	pid := os.Getpid()
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_SYNC, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "open failed")
	}

	err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		_ = unix.Close(fd)
		return nil, errors.Join(err, ErrLocked)
	}

	_, err = unix.Write(fd, []byte(strconv.Itoa(pid)))
	if err != nil {
		return nil, errors.Wrap(err, "write failed")
	}
	return func() error {
		return errors.WithStack(errors.Join(unix.Flock(fd, unix.LOCK_UN), unix.Close(fd), os.Remove(path)))
	}, nil
}
