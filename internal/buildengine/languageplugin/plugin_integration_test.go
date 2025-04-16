//go:build integration

package languageplugin

import (
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/result"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/flock"
	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

// These integration tests are meant as a test suite for external plugins.
//
// It is not meant to mirror exactly how FTL calls external plugins, but rather
// to test plugins against the langauge plugin protocol.
//
// It would be great to allow externally built plugins to be tested against this
// test suite.
//
// Each language must provide a module with the name plugintest, which has:
// - A verb which includes in its name the string VERB_NAME_SNIPPET (case sensitive), eg: "func Verbaabbcc(...)"
// - No dependencies, but has lines that include the string "uncommentForDependency:", which enable a
//     on dependency on the module "dependable" when everything before the colon is removed.
//     - dependable.Data is a data type which can be used to avoid unused dependency warnings.

const MODULE_NAME = "plugintest"
const VERB_NAME_SNIPPET = "aabbcc"

type BuildResultType int

const (
	SUCCESS BuildResultType = iota
	FAILURE
	SUCCESSORFAILURE
)

func TestBuilds(t *testing.T) {
	sch := generateInitialSchema(t)
	bctx := &testContext{}
	in.Run(t,
		in.WithLanguages("go"),
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule(MODULE_NAME),
		bctx.startPlugin(),
		bctx.setUpModuleConfig(MODULE_NAME),
		bctx.generateStubs(sch.InternalModules()...),
		bctx.syncStubReferences("builtin", "dependable"),

		// Build once
		bctx.build(false, []string{}, sch, "build-once"),
		bctx.waitForBuildToEnd(SUCCESS, "build-once", false, nil),

		// Sending build context updates should fail if the plugin has no way to send back a build result
		in.Fail(
			bctx.sendUpdatedBuildContext("no-build-stream", []string{}, sch),
			"expected error when sending build context update without a build stream",
		),

		// Build and enable rebuilding automatically
		bctx.build(true, []string{}, sch, "build-and-watch"),
		bctx.waitForBuildToEnd(SUCCESS, "build-and-watch", false, nil),

		// Update verb name and expect auto rebuild started and ended
		bctx.modifyVerbName(MODULE_NAME, VERB_NAME_SNIPPET, "aaabbbccc"),
		in.IfLanguages(bctx.waitForAutoRebuildToStart("build-and-watch"), "go"),
		bctx.waitForBuildToEnd(SUCCESS, "build-and-watch", true, func(t testing.TB, ic in.TestContext, event *langpb.BuildResponse) {
			successEvent, ok := event.Event.(*langpb.BuildResponse_BuildSuccess)
			assert.True(t, ok)
			_, found := slices.Find(successEvent.BuildSuccess.Module.Decls, func(decl *schemapb.Decl) bool {
				verb, ok := decl.Value.(*schemapb.Decl_Verb)
				if !ok {
					return false
				}
				return strings.Contains(verb.Verb.Name, "aaabbbccc")
			})
			assert.True(t, found, "expected verb name to be updated to include %q", "aaabbbccc")
		}),

		// Trigger an auto rebuild, but when we are told of the build being started, send a build context update
		// to force a new build
		bctx.modifyVerbName(MODULE_NAME, "aaabbbccc", "aaaabbbbcccc"),
		in.IfLanguages(bctx.waitForAutoRebuildToStart("build-and-watch"), "go"),
		bctx.sendUpdatedBuildContext("explicit-build", []string{}, sch),
		bctx.waitForBuildToEnd(SUCCESSORFAILURE, "build-and-watch", true, nil),
		bctx.waitForBuildToEnd(SUCCESS, "explicit-build", false, nil),

		// Trigger 2 explicit builds, make sure we get a response for both of them (first one can fail)
		bctx.sendUpdatedBuildContext("double-build-1", []string{}, sch),
		bctx.sendUpdatedBuildContext("double-build-2", []string{}, sch),
		bctx.waitForBuildToEnd(SUCCESSORFAILURE, "double-build-1", false, nil),
		bctx.waitForBuildToEnd(SUCCESS, "double-build-2", false, nil),

		bctx.killPlugin(),
	)
}

func TestDependenciesUpdate(t *testing.T) {
	sch := generateInitialSchema(t)

	bctx := &testContext{}

	in.Run(t,
		in.WithLanguages("go"), //no java support yet, as it relies on writeGenericSchemaFiles
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule(MODULE_NAME),
		bctx.startPlugin(),
		bctx.setUpModuleConfig(MODULE_NAME),
		bctx.generateStubs(sch.InternalModules()...),
		bctx.syncStubReferences("builtin", "dependable"),

		// Build
		bctx.build(false, []string{}, sch, "initial-ctx"),
		bctx.waitForBuildToEnd(SUCCESS, "initial-ctx", false, nil),

		// Add dependency, build, and expect a failure due to invalidated dependencies
		bctx.addDependency(MODULE_NAME, "dependable"),
		bctx.build(false, []string{}, sch, "detect-dep"),
		bctx.waitForBuildToEnd(FAILURE, "detect-dep", false, func(t testing.TB, ic in.TestContext, event *langpb.BuildResponse) {
			failureEvent, ok := event.Event.(*langpb.BuildResponse_BuildFailure)
			assert.True(t, ok)
			assert.True(t, failureEvent.BuildFailure.InvalidateDependencies, "expected dependencies to be invalidated")
		}),

		// Build with new dependency
		bctx.build(false, []string{"dependable"}, sch, "dep-added"),
		bctx.waitForBuildToEnd(SUCCESS, "dep-added", false, nil),

		bctx.killPlugin(),
	)
}

// TestBuildLock tests that the build lock file is created and removed as expected for each build.
func TestBuildLock(t *testing.T) {
	sch := generateInitialSchema(t)

	bctx := &testContext{}

	in.Run(t,
		in.WithLanguages("go"),
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule(MODULE_NAME),
		bctx.startPlugin(),
		bctx.setUpModuleConfig(MODULE_NAME),
		bctx.generateStubs(sch.InternalModules()...),
		bctx.syncStubReferences("builtin", "dependable"),

		// Build and enable rebuilding automatically
		bctx.checkBuildLockLifecycle(
			bctx.build(true, []string{}, sch, "build-and-watch"),
			bctx.waitForBuildToEnd(SUCCESS, "build-and-watch", false, nil),
		),

		// Update verb name and expect auto rebuild started and ended
		bctx.modifyVerbName(MODULE_NAME, VERB_NAME_SNIPPET, "aaabbbccc"),
		bctx.checkBuildLockLifecycle(
			bctx.waitForAutoRebuildToStart("build-and-watch"),
			bctx.waitForBuildToEnd(SUCCESS, "build-and-watch", true, nil),
		),
	)
}

// TestBuildsWhenAlreadyLocked tests how builds work if there are locks already present.
func TestBuildsWhenAlreadyLocked(t *testing.T) {
	sch := generateInitialSchema(t)

	bctx := &testContext{}

	in.Run(t,
		in.WithLanguages("go"),
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule(MODULE_NAME),
		bctx.startPlugin(),
		bctx.setUpModuleConfig(MODULE_NAME),
		bctx.generateStubs(sch.InternalModules()...),
		bctx.syncStubReferences("builtin", "dependable"),

		// Build and enable rebuilding automatically
		bctx.checkBuildLockLifecycle(
			bctx.build(true, []string{}, sch, "build-and-watch"),
			bctx.waitForBuildToEnd(SUCCESS, "build-and-watch", false, nil),
		),

		// Confirm that build lock changes do not trigger a rebuild triggered by file changes
		bctx.obtainAndReleaseBuildLock(3*time.Second),
		bctx.checkForNoEvents(3*time.Second),

		// Confirm that builds fail or stall when a lock file is already present
		bctx.checkLockedBehavior(
			bctx.sendUpdatedBuildContext("updated-ctx", []string{}, sch),
			bctx.waitForBuildToEnd(FAILURE, "updated-ctx", false, nil),
		),
	)
}

func generateInitialSchema(t *testing.T) *schema.Schema {
	t.Helper()

	sch, err := (&schema.Schema{
		Realms: []*schema.Realm{{
			Modules: []*schema.Module{
				{
					Name: "dependable",
					Decls: []schema.Decl{
						&schema.Data{
							Name:   "Data",
							Export: true,
						},
					},
				},
			}},
		},
	}).Validate()
	assert.NoError(t, err)
	return sch
}

type testContext struct {
	client          *pluginClientImpl
	bindURL         *url.URL
	config          moduleconfig.ModuleConfig
	buildChan       chan result.Result[*langpb.BuildResponse]
	buildChanCancel streamCancelFunc
}

func (bctx *testContext) startPlugin() in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Starting plugin")
		bctx.bindURL = must.Get(url.Parse("http://127.0.0.1:8892"))
		var err error
		bctx.client, err = newClientImpl(ic.Context, ic.WorkingDir(), ic.Language, "test")
		assert.NoError(t, err)
	}
}

func (bctx *testContext) killPlugin() in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Killing plugin")
		err := bctx.client.kill()
		assert.NoError(t, err, "could not kill plugin")

		time.Sleep(1 * time.Second)

		// check that the bind port is freed (ie: the plugin has exited)
		var l *net.TCPListener
		_, portStr, err := net.SplitHostPort(bctx.bindURL.Host)
		assert.NoError(t, err)
		port, err := strconv.Atoi(portStr)
		assert.NoError(t, err)
		l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(bctx.bindURL.Hostname()), Port: port})
		if err != nil {
			// panic so that we don't retry, which can hide the real error
			panic("plugin's port is still in use")
		}
		_ = l.Close()
	}
}

func (bctx *testContext) setUpModuleConfig(moduleName string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Setting up config for %s", moduleName)
		path := filepath.Join(ic.WorkingDir(), moduleName)
		defaults, err := GetModuleConfigDefaults(ic.Context, ic.Language, path)
		assert.NoError(t, err)

		unvalidatedConfig, err := moduleconfig.LoadConfig(path)
		assert.NoError(t, err)

		bctx.config, err = unvalidatedConfig.FillDefaultsAndValidate(defaults, projectconfig.Config{Name: "test"})
		assert.NoError(t, err)
	}
}

func (bctx *testContext) generateStubs(moduleSchs ...*schema.Module) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Generating stubs for %v", slices.Map(moduleSchs, func(m *schema.Module) string { return m.Name }))
		wg, wgctx := errgroup.WithContext(ic.Context)
		for _, moduleSch := range moduleSchs {
			wg.Go(func() error {
				var configForStub moduleconfig.ModuleConfig
				if moduleSch.Name == bctx.config.Module {
					configForStub = bctx.config
				} else if moduleSch.Name == "builtin" {
					configForStub = moduleconfig.ModuleConfig{
						Module:   "builtin",
						Language: "go",
					}
				} else {
					configForStub = moduleconfig.ModuleConfig{
						Module:   moduleSch.Name,
						Language: "fake",
					}
				}

				configForStubProto, err := langpb.ModuleConfigToProto(configForStub.Abs())
				assert.NoError(t, err)

				var nativeConfigProto *langpb.ModuleConfig
				if moduleSch.Name != bctx.config.Module {
					nativeConfigProto, err = langpb.ModuleConfigToProto(bctx.config.Abs())
					assert.NoError(t, err)
				}

				path := filepath.Join(ic.WorkingDir(), ".ftl", bctx.config.Language, "modules", configForStub.Module)
				err = os.MkdirAll(path, 0750)
				assert.NoError(t, err)

				_, err = bctx.client.generateStubs(wgctx, connect.NewRequest(&langpb.GenerateStubsRequest{
					Dir:                path,
					Module:             moduleSch.ToProto(),
					ModuleConfig:       configForStubProto,
					NativeModuleConfig: nativeConfigProto,
				}))
				assert.NoError(t, err)
				return nil
			})
		}
		err := wg.Wait()
		assert.NoError(t, err)
	}
}

func (bctx *testContext) syncStubReferences(moduleNames ...string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Syncing stub references for %v", moduleNames)

		configProto, err := langpb.ModuleConfigToProto(bctx.config.Abs())
		assert.NoError(t, err)

		_, err = bctx.client.syncStubReferences(ic.Context, connect.NewRequest(&langpb.SyncStubReferencesRequest{
			ModuleConfig: configProto,
			StubsRoot:    filepath.Join(ic.WorkingDir(), ".ftl", bctx.config.Language, "modules"),
			Modules:      moduleNames,
		}))
		assert.NoError(t, err)
	}
}

func (bctx *testContext) build(rebuildAutomatically bool, dependencies []string, sch *schema.Schema, contextId string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Plugin building: %s", contextId)
		configProto, err := langpb.ModuleConfigToProto(bctx.config.Abs())
		assert.NoError(t, err)

		schemaProto := sch.ToProto()
		bctx.buildChan, bctx.buildChanCancel, err = bctx.client.build(ic.Context, connect.NewRequest(&langpb.BuildRequest{
			ProjectConfig: &langpb.ProjectConfig{
				Dir:  ic.WorkingDir(),
				Name: "test",
			},
			StubsRoot: filepath.Join(ic.WorkingDir(), ".ftl", bctx.config.Language, "modules"),
			BuildContext: &langpb.BuildContext{
				Id:           contextId,
				ModuleConfig: configProto,
				Schema:       schemaProto,
				Dependencies: dependencies,
			},
			RebuildAutomatically: rebuildAutomatically,
		}))
		assert.NoError(t, err)
	}
}

func (bctx *testContext) sendUpdatedBuildContext(contextId string, dependencies []string, sch *schema.Schema) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Sending updated context to plugin: %s", contextId)
		configProto, err := langpb.ModuleConfigToProto(bctx.config.Abs())
		assert.NoError(t, err)

		schemaProto := sch.ToProto()
		_, err = bctx.client.buildContextUpdated(ic.Context, connect.NewRequest(&langpb.BuildContextUpdatedRequest{
			BuildContext: &langpb.BuildContext{
				Id:           contextId,
				ModuleConfig: configProto,
				Schema:       schemaProto,
				Dependencies: dependencies,
			},
		}))
		assert.NoError(t, err)
	}
}

func (bctx *testContext) waitForAutoRebuildToStart(contextId string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Waiting for auto rebuild to start: %s", contextId)
		logger := log.FromContext(ic.Context)
		assert.NotZero(t, bctx.buildChan, "buildChan must be set before calling waitForAutoRebuildStarted")
		for {
			event, err := (<-bctx.buildChan).Result()
			assert.NoError(t, err, "did not expect a build stream error")
			switch event := event.Event.(type) {
			case *langpb.BuildResponse_AutoRebuildStarted:
				if event.AutoRebuildStarted.ContextId == contextId {
					return
				} else {
					logger.Warnf("ignoring automatic rebuild started event for unexpected context %q instead of %q", event.AutoRebuildStarted.ContextId, contextId)
				}
			case *langpb.BuildResponse_BuildSuccess:
				if event.BuildSuccess.ContextId == contextId {
					panic("build succeeded, but expected auto rebuild started event first")
				} else {
					logger.Warnf("ignoring build success for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildSuccess.ContextId, contextId)
				}
			case *langpb.BuildResponse_BuildFailure:
				if event.BuildFailure.ContextId == contextId {
					panic("build failed, but expected auto rebuild started event first")
				} else {
					logger.Warnf("ignoring build failure for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildFailure.ContextId, contextId)
				}
			}
		}
	}
}

func (bctx *testContext) waitForBuildToEnd(success BuildResultType, contextId string, automaticRebuild bool, additionalChecks func(t testing.TB, ic in.TestContext, event *langpb.BuildResponse)) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		switch success {
		case SUCCESSORFAILURE:
			in.Infof("Waiting for build to end: %s", contextId)
		case SUCCESS:
			in.Infof("Waiting for build to succeed: %s", contextId)
		case FAILURE:
			in.Infof("Waiting for build to fail: %s", contextId)
		}
		logger := log.FromContext(ic.Context)
		assert.NotZero(t, bctx.buildChan, "buildChan must be set before calling waitForAutoRebuildStarted")
		for {
			e, err := (<-bctx.buildChan).Result()
			assert.NoError(t, err, "did not expect a build stream error")

			switch event := e.Event.(type) {
			case *langpb.BuildResponse_AutoRebuildStarted:
				if event.AutoRebuildStarted.ContextId != contextId {
					logger.Warnf("Ignoring automatic rebuild started event for unexpected context %q instead of %q", event.AutoRebuildStarted.ContextId, contextId)
					continue
				}
				logger.Debugf("Ignoring auto rebuild started event for the build we are waiting to finish %q", contextId)

			case *langpb.BuildResponse_BuildSuccess:
				if event.BuildSuccess.ContextId != contextId {
					logger.Warnf("Ignoring build success for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildSuccess.ContextId, contextId)
					continue
				}
				if automaticRebuild != event.BuildSuccess.IsAutomaticRebuild {
					logger.Warnf("Ignoring build success for unexpected context %q (IsAutomaticRebuild=%v, expected=%v)", contextId, event.BuildSuccess.IsAutomaticRebuild, automaticRebuild)
					continue
				}
				if success == FAILURE {
					panic(fmt.Sprintf("build succeeded when we expected it to fail: %v", event.BuildSuccess))
				}
				if additionalChecks != nil {
					additionalChecks(t, ic, e)
				}
				return
			case *langpb.BuildResponse_BuildFailure:
				if event.BuildFailure.ContextId != contextId {
					logger.Warnf("Ignoring build failure for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildFailure.ContextId, contextId)
					continue
				}
				if automaticRebuild != event.BuildFailure.IsAutomaticRebuild {
					logger.Warnf("Ignoring build failure for unexpected context %q (IsAutomaticRebuild=%v, expected=%v)", contextId, event.BuildFailure.IsAutomaticRebuild, automaticRebuild)
					continue
				}
				if success == SUCCESS {
					panic(fmt.Sprintf("build failed when we expected it to succeed: %v", event.BuildFailure))
				}
				if additionalChecks != nil {
					additionalChecks(t, ic, e)
				}
				return
			}
		}
	}
}

func (bctx *testContext) checkForNoEvents(duration time.Duration) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Checking for no events for %v", duration)
		for {
			select {
			case result := <-bctx.buildChan:
				e, err := result.Result()
				assert.NoError(t, err, "did not expect a build stream error")
				switch event := e.Event.(type) {
				case *langpb.BuildResponse_AutoRebuildStarted:
					panic(fmt.Sprintf("rebuild started event when expecting no events: %v", event))
				case *langpb.BuildResponse_BuildSuccess:
					panic(fmt.Sprintf("build success event when expecting no events: %v", event))
				case *langpb.BuildResponse_BuildFailure:
					panic(fmt.Sprintf("build failure event when expecting no events: %v", event))
				}
			case <-time.After(duration):
				return
			case <-ic.Context.Done():
				return
			}
		}
	}
}

func (bctx *testContext) addDependency(moduleName, depName string) in.Action {
	searchStr := "uncommentForDependency:"
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Adding dependency: %s", depName)
		found := false
		assert.NoError(t, bctx.walkWatchedFiles(t, ic, moduleName, func(path string) {
			bytes, err := os.ReadFile(path)
			assert.NoError(t, err)

			lines := strings.Split(string(bytes), "\n")

			foundInFile := false
			for i, line := range lines {
				start := strings.Index(line, searchStr)
				if start == -1 {
					continue
				}
				foundInFile = true
				end := start + len(searchStr)
				lines[i] = line[end:]
			}
			if foundInFile {
				found = true
				os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}))
		assert.True(t, found, "could not add dependency because %q was not found in any files that are watched", searchStr)
	}
}

func (bctx *testContext) modifyVerbName(moduleName, old, new string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Modifying verb name: %s -> %s", old, new)
		found := false
		assert.NoError(t, bctx.walkWatchedFiles(t, ic, moduleName, func(path string) {
			bytes, err := os.ReadFile(path)
			assert.NoError(t, err)

			lines := strings.Split(string(bytes), "\n")

			foundInFile := false
			for i, line := range lines {
				start := strings.Index(line, old)
				if start == -1 {
					continue
				}
				foundInFile = true
				end := start + len(old)
				lines[i] = line[:start] + new + line[end:]
			}
			if foundInFile {
				found = true
				os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}))
		assert.True(t, found, "could not modify verb name because %q was not found in any files that are watched", old)
	}
}

func (bctx *testContext) walkWatchedFiles(t testing.TB, ic in.TestContext, moduleName string, visit func(path string)) error {
	path := filepath.Join(ic.WorkingDir(), moduleName)
	return errors.WithStack(watch.WalkDir(path, true, func(srcPath string, entry fs.DirEntry) error {
		if entry.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(path, srcPath)
		assert.NoError(t, err)

		_, matched := slices.Find(bctx.config.Watch, func(pattern string) bool {
			match, err := doublestar.PathMatch(pattern, relativePath)
			assert.NoError(t, err)
			return match
		})
		if matched {
			visit(srcPath)
		}
		return nil
	}))
}

func (bctx *testContext) checkBuildLockLifecycle(childActions ...in.Action) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Checking build lock is unlocked")
		_, err := os.Stat(bctx.config.Abs().BuildLock)
		assert.Error(t, err, "expected build lock file to not exist before building at %v", bctx.config.Abs().BuildLock)

		lockFound := make(chan bool)
		go func() {
			defer close(lockFound)
			startTime := time.Now()
			for {
				select {
				case <-time.After(1 * time.Second):
					if _, err := os.Stat(bctx.config.Abs().BuildLock); err == nil {
						lockFound <- true
						return
					}
					if time.Since(startTime) > 3*time.Second {
						lockFound <- false
						return
					}
				case <-ic.Context.Done():
					lockFound <- false
					return
				}
			}
		}()

		// do build actions
		for _, childAction := range childActions {
			childAction(t, ic)
		}
		// confirm that at some point we did find the lock file
		assert.True(t, (<-lockFound), "never found build lock file at %v while building", bctx.config.Abs().BuildLock)

		in.Infof("Checking build lock is unlocked")
		_, err = os.Stat(bctx.config.Abs().BuildLock)
		assert.Error(t, err, "expected build lock file to not exist after building at %v", bctx.config.Abs().BuildLock)
	}
}

func (bctx *testContext) obtainAndReleaseBuildLock(duration time.Duration) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Obtaining and releasing build lock")
		release, err := flock.Acquire(ic.Context, bctx.config.Abs().BuildLock, BuildLockTimeout)
		assert.NoError(t, err, "could not get build lock")
		time.Sleep(duration)
		err = release()
		assert.NoError(t, err, "could not release build lock")
	}
}

func (bctx *testContext) checkLockedBehavior(buildFailureActions ...in.Action) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Acquiring build lock: %v", bctx.config.Abs().BuildLock)
		release, err := flock.Acquire(ic.Context, bctx.config.Abs().BuildLock, BuildLockTimeout)
		assert.NoError(t, err, "could not get build lock")

		// build on a separate goroutine
		buildEnded := make(chan bool)
		go func() {
			for _, buildAction := range buildFailureActions {
				buildAction(t, ic)
			}
			close(buildEnded)
		}()

		// wait for build to fail due to file lock
		<-buildEnded

		err = release() //nolint:errcheck
		assert.NoError(t, err, "could not release build lock")
	}
}
