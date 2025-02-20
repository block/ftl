//go:build !release

package buildengine

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/terminal"
)

var lock = sync.Mutex{}
var built = false

func buildRequiredJARS(ctx context.Context) {
	lock.Lock()
	defer lock.Unlock()
	if built {
		return
	}
	built = true
	sm := terminal.FromContext(ctx)
	logger := log.FromContext(ctx)
	const jvmjars = "Build JVM Jars"
	sm.SetModuleState(jvmjars, terminal.BuildStateBuilding)
	defer func() {
		sm.SetModuleState(jvmjars, terminal.BuildStateTerminated)
	}()
	gitRoot, ok := internal.GitRoot(os.Getenv("FTL_DIR")).Get()
	if !ok {
		logger.Warnf("failed to find Git root")
		return
	}

	// Lock the frontend directory to prevent concurrent builds.
	release, err := flock.Acquire(ctx, filepath.Join(gitRoot, ".jvm.lock"), 2*time.Minute)
	if err != nil {
		logger.Errorf(err, "failed to acquire lock")
		return
	}

	log.FromContext(ctx).Scope("console").Infof("Building JVM Jars...")

	err = exec.Command(ctx, log.Debug, gitRoot, "just", "build-jvm").RunBuffered(ctx)
	if lerr := release(); lerr != nil {
		logger.Errorf(lerr, "failed to release lock")
	}
	if err != nil {
		logger.Errorf(err, "failed to build jars")
	}
}
