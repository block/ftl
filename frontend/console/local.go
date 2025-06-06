//go:build !release

package console

import (
	"context"
	glog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/cors"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/terminal"
)

var proxyURL, _ = url.Parse("http://localhost:5173") //nolint:errcheck
var proxy = httputil.NewSingleHostReverseProxy(proxyURL)

func PrepareServer(ctx context.Context) error {
	sm := terminal.FromContext(ctx)
	// This looks the same as the default logger, but it will use a redirected STDERR
	proxy.ErrorLog = glog.New(os.Stderr, "", glog.LstdFlags)
	const console = "FTL Console (dev)"
	sm.SetModuleState(console, terminal.BuildStateBuilding)
	defer func() {
		sm.SetModuleState(console, terminal.BuildStateTerminated)
	}()
	gitRoot, ok := internal.GitRoot(os.Getenv("FTL_DIR")).Get()
	if !ok {
		return errors.Errorf("failed to find Git root")
	}

	// Lock the frontend directory to prevent concurrent builds.
	release, err := flock.Acquire(ctx, filepath.Join(gitRoot, ".frontend.lock"), 2*time.Minute)
	if err != nil {
		return errors.Wrap(err, "failed to acquire lock")
	}

	log.FromContext(ctx).Scope("console").Infof("Building console...")

	err = exec.Command(ctx, log.Debug, gitRoot, "just", "build-frontend").RunBuffered(ctx)
	if lerr := release(); lerr != nil {
		return errors.WithStack(errors.Join(errors.Wrap(lerr, "failed to release lock")))
	}
	if err != nil {
		return errors.Wrap(err, "failed to build frontend")
	}

	return nil
}

func Server(ctx context.Context, timestamp time.Time, allowOrigin *url.URL) (http.Handler, error) {
	gitRoot, ok := internal.GitRoot(os.Getenv("FTL_DIR")).Get()
	if !ok {
		return nil, errors.Errorf("failed to find Git root")
	}

	err := exec.Command(ctx, log.Debug, path.Join(gitRoot, "frontend", "console"), "pnpm", "run", "dev").Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if allowOrigin == nil {
		return proxy, nil
	}

	return cors.Middleware([]string{allowOrigin.String()}, nil, proxy), nil
}
