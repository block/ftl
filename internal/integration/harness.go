//go:build integration || infrastructure || smoketest

package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"github.com/otiai10/copy"
	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/console/v1/consolepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	ftlexec "github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/kube"
	"github.com/block/ftl/internal/rpc"
)

const dumpPath = "/tmp/ftl-kube-report"
const dumpContentsPath = "/tmp/ftl-test-content"

var RedPandaBrokers = []string{"127.0.0.1:19092"}

func (i TestContext) integrationTestTimeout() time.Duration {
	timeout := optional.Zero(os.Getenv("FTL_INTEGRATION_TEST_TIMEOUT")).Default("25s")
	d, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}
	if i.kubeClient.Ok() {
		// kube can be slow, give it some time
		return d * 5
	}
	return d
}

func Infof(format string, args ...any) {
	fmt.Printf("\033[32m\033[1mINFO: "+format+"\033[0m\n", args...)
}

var buildOnce sync.Once

var buildOnceOptions *options

// An Option for configuring the integration test harness.
type Option func(*options)

// ActionOrOption is a type that can be either an Action or an Option.
type ActionOrOption any

// WithLanguages is a Run* option that specifies the languages to test.
//
// Defaults to "go" if not provided.
func WithLanguages(languages ...string) Option {
	return func(o *options) {
		o.languages = languages
	}
}

// WithKubernetes is a Run* option that specifies tests should be run on a kube cluster
func WithKubernetes(helmArgs ...string) Option {
	return func(o *options) {
		o.kube = true
		o.startController = false
		o.helmArgs = helmArgs
		o.envars["FTL_ENDPOINT"] = "http://localhost:8792"
	}
}

func WithDebugLogging() Option {
	return func(o *options) {
		o.debugLogging = true
	}
}

// WithLocalstack is a Run* option that specifies tests should be run on a localstack container
func WithLocalstack() Option {
	return func(o *options) {
		o.localstack = true
	}
}

// WithConsole is a Run* option that specifies tests should build and start the console
func WithConsole() Option {
	return func(o *options) {
		o.console = true
	}
}

// WithPubSub is a Run* option that specifies tests should reset red panda before starting
func WithPubSub() Option {
	return func(o *options) {
		o.resetPubSub = true
	}
}

// WithTestDataDir sets the directory from which to look for test data.
//
// Defaults to "testdata/<language>" if not provided.
func WithTestDataDir(dir string) Option {
	return func(o *options) {
		o.testDataDir = dir
	}
}

// WithFTLConfig is a Run* option that specifies the FTL config to use.
//
// This will set FTL_CONFIG for this test, then pass in the relative
// path based on ./testdata/go/ where "." denotes the directory containing the
// integration test (e.g. for "integration/harness_test.go" supplying
// "database/ftl-project.toml" would set FTL_CONFIG to
// "integration/testdata/go/database/ftl-project.toml").
func WithFTLConfig(path string) Option {
	return func(o *options) {
		o.ftlConfigPath = path
	}
}

// WithEnvar is a Run* option that specifies an environment variable to set.
func WithEnvar(key, value string) Option {
	return func(o *options) {
		o.envars[key] = value
	}
}

// WithJavaBuild is a Run* option that ensures the Java runtime is built.
// If the test languages contain java this is not necessary, as it is implied
// Note that this will not actually add Java as a language under test
func WithJavaBuild() Option {
	return func(o *options) {
		o.requireJava = true
	}
}

// WithoutController is a Run* option that disables starting the controller.
func WithoutController() Option {
	return func(o *options) {
		o.startController = false
	}
}

// WithoutTimeline is a Run* option that disables starting the timeline service.
func WithoutTimeline() Option {
	return func(o *options) {
		o.startTimeline = false
	}
}

// WithProvisionerConfig is a Run* option that specifies the provisioner config to use.
func WithProvisionerConfig(config string) Option {
	return func(o *options) {
		o.provisionerConfig = config
	}
}

// WithDevMode starts the server using FTL dev, so modules are deployed automatically on change
func WithDevMode() Option {
	return func(o *options) {
		o.devMode = true
	}
}

type options struct {
	languages         []string
	testDataDir       string
	ftlConfigPath     string
	startController   bool
	devMode           bool
	startTimeline     bool
	provisionerConfig string
	requireJava       bool
	envars            map[string]string
	kube              bool
	localstack        bool
	console           bool
	resetPubSub       bool
	helmArgs          []string
	debugLogging      bool
}

// Run an integration test.
func Run(t *testing.T, actionsOrOptions ...ActionOrOption) {
	t.Helper()
	run(t, actionsOrOptions...)
}

func run(t *testing.T, actionsOrOptions ...ActionOrOption) {
	t.Helper()
	opts := options{
		startController: true,
		startTimeline:   true,
		languages:       []string{"go"},
		envars:          map[string]string{},
	}
	actions := []Action{}
	for _, opt := range actionsOrOptions {
		switch o := opt.(type) {
		case Action:
			actions = append(actions, o)

		case func(t testing.TB, ic TestContext):
			actions = append(actions, Action(o))

		case Option:
			o(&opts)

		case func(*options):
			o(&opts)

		default:
			panic(fmt.Sprintf("expected Option or Action, not %T", opt))
		}
	}

	for key, value := range opts.envars {
		t.Setenv(key, value)
	}

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir, ok := internal.GitRoot("").Get()
	assert.True(t, ok)

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	binDir := filepath.Join(rootDir, "build", "release")
	println("::group::" + t.Name())
	defer println("::endgroup::")

	var kubeClient *kubernetes.Clientset
	var kubeNamespace string
	buildOnce.Do(func() {
		buildOnceOptions = &opts
		if opts.kube {
			// This command will build a linux/amd64 version of FTL and deploy it to the kube cluster
			Infof("Building FTL and deploying to kube")
			err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "setup-istio-cluster").RunBuffered(ctx)
			assert.NoError(t, err)

			// On CI we always skip the full deploy, as the build is done in the CI pipeline
			skipKubeFullDeploy := os.Getenv("CI") != ""
			var args []string
			if skipKubeFullDeploy {
				Infof("Using apply instead of full-deploy CI is set")
				args = append([]string{}, "apply")
			} else {
				args = append([]string{}, "full-deploy")
			}
			args = append(args, opts.helmArgs...)
			err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", args...).RunBuffered(ctx)
			assert.NoError(t, err)
			if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
				// If we are already on linux/amd64 we don't need to rebuild, otherwise we now need a native one to interact with the kube cluster
				Infof("Building FTL for native OS")
				err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
				assert.NoError(t, err)
			}
			kubeClient, err = kube.CreateClientSet()
			assert.NoError(t, err)
			kubeNamespace, err = kube.GetCurrentNamespace()
			assert.NoError(t, err)
			// We create the client before, as kube itself is up, we are just waiting for the deployments
			err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "wait-for-kube").RunBuffered(ctx)
			if err != nil {
				dumpKubePods(ctx, nil, optional.Ptr(kubeClient), kubeNamespace)
			}
			assert.NoError(t, err)
			ver := os.Getenv("OLD_FTL_VERSION")
			if ver != "" {
				err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "wait-for-version-upgrade", ver).RunBuffered(ctx)
			}
		} else if os.Getenv("USE_RELEASE_BINARIES") == "1" {
			Infof("Using release binaries for integration tests")
		} else if opts.console {
			Infof("Building ftl with console")
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
			assert.NoError(t, err)
		} else {
			Infof("Building ftl without console")
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build-without-frontend", "ftl").RunBuffered(ctx)
			assert.NoError(t, err)
		}
		if opts.requireJava || slices.Contains(opts.languages, "java") || slices.Contains(opts.languages, "kotlin") {
			if os.Getenv("USE_RELEASE_BINARIES") == "1" {
				Infof("Using release binaries for JVM Artifacts")
			} else {
				err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build-jvm", "-DskipTests", "-B").RunBuffered(ctx)
				assert.NoError(t, err)
			}
		}
		if opts.localstack {
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "localstack").RunBuffered(ctx)
			assert.NoError(t, err)
		}
	})

	assert.Equal(t, *buildOnceOptions, opts, "Options changed between test runs")

	adminPort := "8892"
	consolePort := "8892"
	schemaPort := "8892"
	if opts.kube {
		adminPort = "8792"
		consolePort = "8792"
		schemaPort = "8792"
	}
	for _, language := range opts.languages {
		t.Run(language, func(t *testing.T) {
			println("::group::" + t.Name())
			defer println("::endgroup::")
			t.Helper()

			ctx, done := context.WithCancelCause(ctx)
			defer done(errors.Errorf("test complete"))
			tmpDir := initWorkDir(t, cwd, opts)

			verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:"+adminPort, log.Debug)

			var admin adminpbconnect.AdminServiceClient
			var console consolepbconnect.ConsoleServiceClient
			var schema ftlv1connect.SchemaServiceClient
			if opts.startController {
				Infof("Starting ftl cluster")

				command := []string{"serve", "--recreate"}
				if opts.devMode {
					command = []string{"dev"}
				}
				command = append(command, "--log-level=DEBUG", "--log-timestamps")

				args := append([]string{filepath.Join(binDir, "ftl")}, command...)
				if !opts.console {
					args = append(args, "--no-console")
				}

				if opts.provisionerConfig != "" {
					configFile := filepath.Join(tmpDir, "provisioner-plugin-config.toml")
					os.WriteFile(configFile, []byte(opts.provisionerConfig), 0644)
					args = append(args, "--provisioner-plugin-config="+configFile)
				}
				ctx = startProcess(ctx, t, tmpDir, opts.devMode, args...)
			}
			adminEndpoint := "http://localhost:" + adminPort
			if opts.startController || opts.kube {
				admin = rpc.Dial(adminpbconnect.NewAdminServiceClient, adminEndpoint, log.Debug)
				console = rpc.Dial(consolepbconnect.NewConsoleServiceClient, "http://localhost:"+consolePort, log.Debug)
				schema = rpc.Dial(ftlv1connect.NewSchemaServiceClient, "http://localhost:"+schemaPort, log.Debug)
			}

			testData := filepath.Join(cwd, "testdata", language)
			if opts.testDataDir != "" {
				testData = opts.testDataDir
			}

			ic := TestContext{
				Context:       ctx,
				RootDir:       rootDir,
				testData:      testData,
				workDir:       tmpDir,
				binDir:        binDir,
				Verbs:         verbs,
				realT:         t,
				Language:      language,
				kubeNamespace: kubeNamespace,
				devMode:       opts.devMode,
				kubeClient:    optional.Ptr(kubeClient),
			}
			defer dumpTestContents(ctx, ic)
			defer dumpKubePods(ctx, t, ic.kubeClient, ic.kubeNamespace)
			ic.AdminEndpoint = adminEndpoint
			if opts.startController || opts.kube {
				ic.Admin = admin
				ic.Schema = schema
				ic.Console = console

				Infof("Waiting for admin to be ready")
				assert.NoError(t, rpc.Wait(ctx, backoff.Backoff{Max: time.Millisecond * 50}, time.Minute*2, ic.Admin))
			}

			if opts.startTimeline {
				ic.Timeline = rpc.Dial(timelinepbconnect.NewTimelineServiceClient, adminEndpoint, log.Debug)
				Infof("Waiting for timeline to be ready")
				assert.NoError(t, rpc.Wait(ctx, backoff.Backoff{Max: time.Millisecond * 50}, time.Minute*2, ic.Timeline))
			}

			if opts.devMode {
				ic.BuildEngine = rpc.Dial(buildenginepbconnect.NewBuildEngineServiceClient, adminEndpoint, log.Debug)

				Infof("Waiting for build engine client to be ready")
				assert.NoError(t, rpc.Wait(ctx, backoff.Backoff{Max: time.Millisecond * 50}, time.Minute*2, ic.Timeline))
			}

			if opts.resetPubSub {
				Infof("Resetting pubsub")
				envars := []string{"COMPOSE_IGNORE_ORPHANS=True"}
				err = exec.CommandWithEnv(ctx, log.Debug, rootDir, envars, "docker", "compose", "-f", "internal/dev/docker-compose.redpanda.yml", "-p", "ftl", "up", "-d", "--wait").RunBuffered(ctx)
				assert.NoError(t, err)

				client, err := sarama.NewClient(RedPandaBrokers, sarama.NewConfig())
				assert.NoError(t, err)
				defer client.Close()

				clusterAdmin, err := sarama.NewClusterAdminFromClient(client)
				assert.NoError(t, err)
				defer clusterAdmin.Close()

				// There can be a slight delay for consumer groups to be empty after a previous run of FTL ends
				for i := range 15 {
					err = attemptToResetPubSub(clusterAdmin)
					if err == nil {
						break
					}
					if i == 9 {
						assert.Error(t, err, "Error resetting pubsub")
					}
					Infof("Error resetting pubsub, retrying in 1 seconds: %v", err)
					time.Sleep(time.Second)
				}
			}

			Infof("Starting test")

			for _, action := range actions {
				ic.AssertWithRetry(t, action)
			}
		})
	}
}
func attemptToResetPubSub(admin sarama.ClusterAdmin) error {
	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return errors.WithStack(err)
	}
	for name := range groups {
		if strings.HasPrefix(name, "_") {
			continue
		}
		err = admin.DeleteConsumerGroup(name)
		if err != nil {
			return errors.Wrapf(err, "could not delete consumer group %s", name)
		}
	}
	topics, err := admin.ListTopics()
	if err != nil {
		return errors.WithStack(err)
	}
	for name := range topics {
		if strings.HasPrefix(name, "_") {
			continue
		}
		err = admin.DeleteTopic(name)
		if err != nil {
			return errors.Wrapf(err, "could not delete topic %s", name)
		}
	}
	return nil
}

func initWorkDir(t testing.TB, cwd string, opts options) string {
	tmpDir := t.TempDir()

	if opts.ftlConfigPath != "" {
		// TODO: We shouldn't be copying the shared config from the "go" testdata...
		opts.ftlConfigPath = filepath.Join(cwd, "testdata", "go", opts.ftlConfigPath)
		projectPath := filepath.Join(tmpDir, "ftl-project.toml")

		// Copy the specified FTL config to the temporary directory.
		err := copy.Copy(opts.ftlConfigPath, projectPath)
		if err == nil {
			t.Setenv("FTL_CONFIG", projectPath)
			err = copy.Copy(filepath.Join(filepath.Dir(opts.ftlConfigPath), ".ftl"), filepath.Join(filepath.Dir(projectPath), ".ftl")) //nolint
			if err != nil {
				t.Logf("Failed to copy .ftl: %s", err)
			}
		} else {
			// Use a path into the testdata directory instead of one relative to
			// tmpDir. Otherwise we have a chicken and egg situation where the config
			// can't be loaded until the module is copied over, and the config itself
			// is used by FTL during startup.
			// Some tests still rely on this behavior, so we can't remove it entirely.
			t.Logf("Failed to copy %s to %s: %s", opts.ftlConfigPath, projectPath, err)
			t.Setenv("FTL_CONFIG", opts.ftlConfigPath)
		}

	} else {
		err := os.WriteFile(filepath.Join(tmpDir, "ftl-project.toml"), []byte(`name = "ftl"`), 0644)
		assert.NoError(t, err)
	}
	return tmpDir
}

type TestContext struct {
	context.Context
	// Temporary directory the test is executing in.
	workDir string
	// Root of FTL repo.
	RootDir string
	// Path to testdata directory for the current language.
	testData string
	// Path to the "bin" directory.
	binDir string
	// The Language under test
	Language string
	// Set if the test is running on kubernetes
	kubeClient    optional.Option[kubernetes.Clientset]
	kubeNamespace string
	devMode       bool

	AdminEndpoint string
	Admin         adminpbconnect.AdminServiceClient
	Schema        ftlv1connect.SchemaServiceClient
	Console       consolepbconnect.ConsoleServiceClient
	Verbs         ftlv1connect.VerbServiceClient
	Timeline      timelinepbconnect.TimelineServiceClient
	BuildEngine   buildenginepbconnect.BuildEngineServiceClient

	realT *testing.T
}

func (i TestContext) Run(name string, f func(t *testing.T)) bool {
	return i.realT.Run(name, f)
}

// WorkingDir returns the temporary directory the test is executing in.
func (i TestContext) WorkingDir() string { return i.workDir }

// AssertWithRetry asserts that the given action passes within the timeout.
func (i TestContext) AssertWithRetry(t testing.TB, assertion Action) {
	t.Helper()
	i.AssertWithSpecificRetry(t, assertion, i.integrationTestTimeout())
}

// AssertWithSpecificRetry asserts that the given action passes within the timeout.
func (i TestContext) AssertWithSpecificRetry(t testing.TB, assertion Action, timeout time.Duration) {
	t.Helper()
	waitCtx, done := context.WithTimeout(i, timeout)
	defer done()
	for {
		err := i.runAssertionOnce(t, assertion)
		if err == nil {
			return
		}
		select {
		case <-waitCtx.Done():
			t.Fatalf("Timed out waiting for assertion to pass: %s", err)

		case <-time.After(time.Millisecond * 200):
		}
	}
}

// Run an assertion, wrapping testing.TB in an implementation that panics on failure, propagating the error.
func (i TestContext) runAssertionOnce(t testing.TB, assertion Action) (err error) {
	t.Helper()
	defer func() {
		switch r := recover().(type) {
		case TestingError:
			err = errors.WithStack(errors.New(string(r)))
			fmt.Println(string(r))

		case nil:
			return

		default:
			panic(r)
		}
	}()
	assertion(T{t}, i)
	return nil
}

type Action func(t testing.TB, ic TestContext)

type SubTest struct {
	Name   string
	Action Action
}

type logWriter struct {
	mu     sync.Mutex
	logger interface{ Log(...any) }
	buffer []byte
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for {
		index := bytes.IndexByte(p, '\n')
		if index == -1 {
			l.buffer = append(l.buffer, p...)
			return n, nil
		}
		l.buffer = append(l.buffer, p[:index]...)
		l.logger.Log(string(l.buffer))
		l.buffer = l.buffer[:0]
		p = p[index+1:]
	}
}

// startProcess runs a binary in the background and terminates it when the test completes.
func startProcess(ctx context.Context, t testing.TB, tempDir string, devMode bool, args ...string) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancelCause(ctx)
	cmd := ftlexec.Command(ctx, log.Debug, "..", args[0], args[1:]...)
	if devMode {
		cmd.Dir = tempDir
	}
	err := cmd.Start()
	assert.NoError(t, err)
	terminated := make(chan bool)
	go func() {
		err := cmd.Wait()
		select {
		case <-terminated:
		default:
			cancel(errors.Wrap(err, "process terminated"))
			assert.NoError(t, err)
		}
	}()
	t.Cleanup(func() {
		close(terminated)
		err := cmd.Kill(syscall.SIGTERM)
		assert.NoError(t, err)
		cancel(errors.Errorf("test complete"))
	})
	return ctx
}

func dumpTestContents(ctx context.Context, ic TestContext) {
	path := filepath.Join(dumpContentsPath, ic.realT.Name())
	_ = os.RemoveAll(path)               // nolint
	_ = copy.Copy(ic.WorkingDir(), path) // nolint

}

func dumpKubePods(ctx context.Context, t *testing.T, kubeClient optional.Option[kubernetes.Clientset], defaultNs string) {
	if t != nil && !t.Failed() {
		Infof("Skipping dump, test did not fail")
		return
	}

	if client, ok := kubeClient.Get(); ok {
		_ = os.RemoveAll(dumpPath) // #nosec
		namespaceList, err := client.CoreV1().Namespaces().List(ctx, kubemeta.ListOptions{})
		if err != nil {
			Infof("Error listing namespaces: %v", err)
			return
		}
		for _, ns := range namespaceList.Items {
			ok := false
			if ns.Name == defaultNs {
				ok = true
			} else if ns.Labels != nil && ns.Labels["app.kubernetes.io/part-of"] == "ftl" {
				ok = true
			}
			if !ok {
				continue
			}
			pods := client.CoreV1().Pods(ns.Namespace)
			list, err := pods.List(ctx, kubemeta.ListOptions{})
			if err == nil {
				for _, pod := range list.Items {
					Infof("Dumping logs for pod %s", pod.Name)
					podPath := filepath.Join(dumpPath, pod.Name)
					err := os.MkdirAll(podPath, 0755) // #nosec
					if err != nil {
						Infof("Error creating directory %s: %v", podPath, err)
						continue
					}
					podYaml, err := yaml.Marshal(pod)
					if err != nil {
						Infof("Error marshalling pod %s: %v", pod.Name, err)
						continue
					}
					err = os.WriteFile(filepath.Join(podPath, "pod.yaml"), podYaml, 0644) // #nosec
					if err != nil {
						Infof("Error writing pod %s: %v", pod.Name, err)
						continue
					}
					for _, container := range pod.Spec.Containers {
						for _, prev := range []bool{false, true} {
							var path string
							if prev {
								path = filepath.Join(dumpPath, pod.Name, container.Name+"-previous.log")
							} else {
								path = filepath.Join(dumpPath, pod.Name, container.Name+".log")
							}
							req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &kubecore.PodLogOptions{Container: container.Name, Previous: prev})
							podLogs, err := req.Stream(ctx)
							if err != nil {
								if prev {
									// This is pretty normal not to have previous logs
									continue
								}
								Infof("Error getting logs for pod %s, %s: %v", pod.Name, container.Name, err)
								continue
							}
							defer func() {
								_ = podLogs.Close()
							}()
							buf := new(bytes.Buffer)
							_, err = io.Copy(buf, podLogs)
							if err != nil {
								Infof("Error copying logs for pod %s: %v", pod.Name, err)
								continue
							}
							str := buf.String()
							err = os.WriteFile(path, []byte(str), 0644) // #nosec
							if err != nil {
								Infof("Error writing logs for pod %s: %v", pod.Name, err)
							}

						}

					}
				}
			}
			istio, err := kube.CreateIstioClientSet()
			if err != nil {
				Infof("Error creating istio clientset: %v", err)
				return
			}
			auths, err := istio.SecurityV1().AuthorizationPolicies(ns.Namespace).List(ctx, kubemeta.ListOptions{})
			if err == nil {
				for _, policy := range auths.Items {
					Infof("Dumping yamp for auth policy %s", policy.Name)
					policyPath := filepath.Join(dumpPath, policy.Name)
					err := os.MkdirAll(policyPath, 0755) // #nosec
					if err != nil {
						Infof("Error creating directory %s: %v", policyPath, err)
						continue
					}
					podYaml, err := yaml.Marshal(policy)
					if err != nil {
						Infof("Error marshalling pod %s: %v", policy.Name, err)
						continue
					}
					err = os.WriteFile(filepath.Join(policyPath, "policy.yaml"), podYaml, 0644) // #nosec
					if err != nil {
						Infof("Error writing policy %s: %v", policy.Name, err)
						continue
					}

				}
			}
		}
	}
}
