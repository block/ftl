package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"

	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/maps"
	"github.com/block/ftl/internal/moduleconfig"
)

// A WatchEvent is an event that occurs when a module is added, removed, or
// changed.
type WatchEvent interface{ watchEvent() }

type WatchEventModuleAdded struct {
	Config moduleconfig.UnvalidatedModuleConfig
}

func (WatchEventModuleAdded) watchEvent() {}

type WatchEventModuleRemoved struct {
	Config moduleconfig.UnvalidatedModuleConfig
}

func (WatchEventModuleRemoved) watchEvent() {}

type WatchEventModuleChanged struct {
	Config  moduleconfig.UnvalidatedModuleConfig
	Changes []FileChange
	Time    time.Time
}

func (c WatchEventModuleChanged) String() string {
	return strings.Join(slices.Map(c.Changes, func(change FileChange) string {
		p, err := filepath.Rel(c.Config.Dir, change.Path)
		if err != nil {
			p = change.Path
		}
		return fmt.Sprintf("%s%s", change.Change, p)
	}), ", ")
}

type FileChange struct {
	Change FileChangeType
	Path   string
}

func (WatchEventModuleChanged) watchEvent() {}

type moduleHashes struct {
	Hashes FileHashes
	Config moduleconfig.UnvalidatedModuleConfig
}

type Watcher struct {
	isWatching bool

	// lock path ensures no modules are scaffolded while a Watcher walks over the files
	lockPath optional.Option[string]
	// patterns are relative to each module found
	patterns []string

	// use mutex whenever accessing / modifying existingModules or moduleTransactions
	mutex              sync.Mutex
	existingModules    map[string]moduleHashes
	moduleTransactions map[string][]*modifyFilesTransaction
}

func NewWatcher(lockPath optional.Option[string], patterns ...string) *Watcher {
	svc := &Watcher{
		existingModules:    map[string]moduleHashes{},
		moduleTransactions: map[string][]*modifyFilesTransaction{},
		lockPath:           lockPath,
		patterns:           patterns,
	}

	return svc
}

func (w *Watcher) GetTransaction(moduleDir string) ModifyFilesTransaction {
	return &modifyFilesTransaction{
		watcher:   w,
		moduleDir: moduleDir,
	}
}

// Watch the given directories for new modules, deleted modules, and changes to
// existing modules, publishing a change event for each.
func (w *Watcher) Watch(ctx context.Context, period time.Duration, moduleDirs []string) (*pubsub.Topic[WatchEvent], error) {
	if w.isWatching {
		return nil, errors.Errorf("file watcher is already watching")
	}
	w.isWatching = true

	logger := log.FromContext(ctx)
	topic := pubsub.New[WatchEvent]()
	logger.Debugf("Starting watch %v", moduleDirs)

	go func() {
		wait := topic.Wait()

		isFirstLoop := true
		for {
			var delayChan <-chan time.Time
			if isFirstLoop {
				// No delay on the first loop
				isFirstLoop = false
				delayChan = time.After(0)
			} else {
				delayChan = time.After(period)
			}

			select {
			case <-delayChan:

			case <-wait:
				return

			case <-ctx.Done():
				_ = topic.Close()
				return
			}

			var flockRelease func() error

			if path, ok := w.lockPath.Get(); ok {
				err := os.Mkdir(filepath.Dir(path), 0700)
				if err != nil && !os.IsExist(err) {
					logger.Debugf("error creating lock directory: %v", err)
				}
				flockRelease, err = flock.Acquire(ctx, path, period)
				if err != nil {
					logger.Debugf("error acquiring modules lock to discover modules: %v", err)
					continue
				}
			} else {
				flockRelease = func() error { return nil }
			}
			modules, err := DiscoverModules(ctx, moduleDirs)
			if err != nil {
				logger.Tracef("error discovering modules: %v", err)
				continue
			}
			if err := flockRelease(); err != nil {
				logger.Debugf("error releasing modules lock after discovering modules: %v", err)
			}

			modulesByDir := maps.FromSlice(modules, func(config moduleconfig.UnvalidatedModuleConfig) (string, moduleconfig.UnvalidatedModuleConfig) {
				return config.Dir, config
			})

			w.detectChanges(ctx, topic, modulesByDir)
		}
	}()
	return topic, nil
}

func (w *Watcher) detectChanges(ctx context.Context, topic *pubsub.Topic[WatchEvent], modulesByDir map[string]moduleconfig.UnvalidatedModuleConfig) {
	logger := log.FromContext(ctx)
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Trigger events for removed modules.
	for _, existingModule := range w.existingModules {
		if transactions, ok := w.moduleTransactions[existingModule.Config.Dir]; ok && len(transactions) > 0 {
			// Skip modules that currently have transactions
			continue
		}
		existingConfig := existingModule.Config
		if _, haveModule := modulesByDir[existingConfig.Dir]; !haveModule {
			logger.Debugf("removed %q", existingModule.Config.Module)
			topic.Publish(WatchEventModuleRemoved{Config: existingModule.Config})
			delete(w.existingModules, existingConfig.Dir)
		}
	}

	// Compare the modules to the existing modules.
	for _, config := range modulesByDir {
		if transactions, ok := w.moduleTransactions[config.Dir]; ok && len(transactions) > 0 {
			// Skip modules that currently have transactions
			continue
		}
		existingModule, haveExistingModule := w.existingModules[config.Dir]
		hashes, err := ComputeFileHashes(config.Dir, true, w.patterns)
		if err != nil {
			logger.Tracef("error computing file hashes for %s: %v", config.Dir, err)
			continue
		}

		if haveExistingModule {
			changes := CompareFileHashes(existingModule.Hashes, hashes)
			if len(changes) == 0 {
				continue
			}
			event := WatchEventModuleChanged{Config: existingModule.Config, Changes: changes, Time: time.Now()}
			logger.Debugf("changed %q: %s", config.Module, event)
			topic.Publish(event)
			w.existingModules[config.Dir] = moduleHashes{Hashes: hashes, Config: existingModule.Config}
			continue
		}
		logger.Debugf("added %q", config.Module)
		topic.Publish(WatchEventModuleAdded{Config: config})
		w.existingModules[config.Dir] = moduleHashes{Hashes: hashes, Config: config}
	}
}

// ModifyFilesTransaction allows builds to modify files in a module without triggering a watch event.
// This helps us avoid infinite loops with builds changing files, and those changes triggering new builds.as a no-op
type ModifyFilesTransaction interface {
	Begin() error
	ModifiedFiles(paths ...string) error
	End() error
}

// Implementation of ModifyFilesTransaction protocol
type modifyFilesTransaction struct {
	watcher   *Watcher
	moduleDir string
	isActive  bool
}

var _ ModifyFilesTransaction = (*modifyFilesTransaction)(nil)

func (t *modifyFilesTransaction) Begin() error {
	if t.isActive {
		return errors.Errorf("transaction is already active")
	}
	t.isActive = true

	t.watcher.mutex.Lock()
	defer t.watcher.mutex.Unlock()

	t.watcher.moduleTransactions[t.moduleDir] = append(t.watcher.moduleTransactions[t.moduleDir], t)

	return nil
}

func (t *modifyFilesTransaction) End() error {
	if !t.isActive {
		return errors.Errorf("transaction is not active")
	}

	t.watcher.mutex.Lock()
	defer t.watcher.mutex.Unlock()

	for idx, transaction := range t.watcher.moduleTransactions[t.moduleDir] {
		if transaction != t {
			continue
		}
		t.isActive = false
		t.watcher.moduleTransactions[t.moduleDir] = append(t.watcher.moduleTransactions[t.moduleDir][:idx], t.watcher.moduleTransactions[t.moduleDir][idx+1:]...)
		return nil
	}
	return errors.Errorf("could not end transaction because it was not found")
}

func (t *modifyFilesTransaction) ModifiedFiles(paths ...string) error {
	if !t.isActive {
		return errors.Errorf("can not modify file because transaction is not active: %v", paths)
	}

	t.watcher.mutex.Lock()
	defer t.watcher.mutex.Unlock()

	moduleHashes, ok := t.watcher.existingModules[t.moduleDir]
	if !ok {
		// skip updating hashes because we have not discovered this module yet
		return nil
	}

	for _, path := range paths {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			delete(moduleHashes.Hashes, path)
			continue
		}
		hash, matched, err := computeFileHash(moduleHashes.Config.Dir, path, t.watcher.patterns)
		if err != nil {
			return errors.WithStack(err)
		}
		if !matched {
			continue
		}

		moduleHashes.Hashes[path] = hash
	}
	t.watcher.existingModules[t.moduleDir] = moduleHashes

	return nil
}
