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
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/result"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	languagepb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/flock"
	in "github.com/block/ftl/internal/integration"
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

func expectBuildSuccess(t *testing.T) func(res *languagepb.BuildResponse) {
	return func(res *languagepb.BuildResponse) {
		_, ok := res.Event.(*languagepb.BuildResponse_BuildSuccess)
		assert.True(t, ok, "Expecting build success event")
	}
}
func expectBuildFailure(t *testing.T) func(res *languagepb.BuildResponse) {
	return func(res *languagepb.BuildResponse) {
		_, ok := res.Event.(*languagepb.BuildResponse_BuildFailure)
		assert.True(t, ok, "Expecting build success event")
	}
}

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
		bctx.build(false, []string{}, sch, expectBuildSuccess(t)),

		// Build and enable rebuilding automatically
		bctx.build(true, []string{}, sch, expectBuildSuccess(t)),

		// Update verb name and expect auto rebuild started and ended
		bctx.modifyVerbName(MODULE_NAME, VERB_NAME_SNIPPET, "aaabbbccc"),
		bctx.build(false, []string{}, sch, func(event *langpb.BuildResponse) {
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
		bctx.build(true, []string{}, sch, expectBuildSuccess(t)),

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
		bctx.build(false, []string{}, sch, expectBuildSuccess(t)),

		// Add dependency, build, and expect a failure due to invalidated dependencies
		bctx.addDependency(MODULE_NAME, "dependable"),
		bctx.build(false, []string{}, sch, expectBuildFailure(t)),

		// Build with new dependency
		bctx.build(false, []string{"dependable"}, sch, expectBuildSuccess(t)),

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
			bctx.build(false, []string{}, sch, expectBuildSuccess(t)),
		),

		// Update verb name and expect auto rebuild started and ended
		bctx.modifyVerbName(MODULE_NAME, VERB_NAME_SNIPPET, "aaabbbccc"),
		bctx.checkBuildLockLifecycle(
			bctx.build(false, []string{}, sch, expectBuildSuccess(t)),
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
							Name:       "Data",
							Visibility: schema.VisibilityScopeModule,
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

func (bctx *testContext) build(rebuildAutomatically bool, dependencies []string, sch *schema.Schema, resultHandler func(res *languagepb.BuildResponse)) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		configProto, err := langpb.ModuleConfigToProto(bctx.config.Abs())
		assert.NoError(t, err)

		schemaProto := sch.ToProto()
		res, err := bctx.client.build(ic.Context, connect.NewRequest(&langpb.BuildRequest{
			ProjectConfig: &langpb.ProjectConfig{
				Dir:  ic.WorkingDir(),
				Name: "test",
			},
			StubsRoot: filepath.Join(ic.WorkingDir(), ".ftl", bctx.config.Language, "modules"),
			BuildContext: &langpb.BuildContext{
				ModuleConfig: configProto,
				Schema:       schemaProto,
				Dependencies: dependencies,
			},
			DevModeBuild: rebuildAutomatically,
		}))
		assert.NoError(t, err)
		resultHandler(res.Msg)
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
