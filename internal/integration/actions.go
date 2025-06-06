//go:build integration || infrastructure

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/block/scaffolder"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
	"github.com/kballard/go-shellquote"
	"github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/devstate"
	"github.com/block/ftl/internal/dsn"
	ftlexec "github.com/block/ftl/internal/exec"
)

// Scaffold a directory relative to the testdata directory to a directory relative to the working directory.
func Scaffold(src, dest string, tmplCtx any) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Scaffolding %s -> %s", src, dest)
		err := scaffolder.Scaffold(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest), tmplCtx)
		assert.NoError(t, err)
	}
}

// GitInit calls git init on the working directory.
func GitInit() Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Running `git init` on the working directory: %s", ic.workDir)
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "git", "init", ic.workDir).RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// Copy a module from the testdata directory to the working directory.
//
// Ensures that any language-specific local modifications are made correctly,
// such as Go module file replace directives for FTL.
func CopyModule(module string) Action {
	return Chain(
		CopyDir(module, module),
		EditGoMod(module),
	)
}

// Copy a module from the testdata directory to the working directory
// This will always use the specified language regardless of the language under test
//
// Ensures that any language-specific local modifications are made correctly,
// such as Go module file replace directives for FTL.
func CopyModuleWithLanguage(module string, language string) Action {
	return Chain(
		CopyDir(filepath.Join("..", language, module), module),
		EditGoMod(module),
	)
}

func EditGoMod(module string) func(t testing.TB, ic TestContext) {
	return func(t testing.TB, ic TestContext) {
		root := filepath.Join(ic.workDir, module)
		// TODO: Load the module configuration from the module itself and use that to determine the language-specific stuff.
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			err := ftlexec.Command(ic, log.Debug, root, "go", "mod", "edit", "-replace", "github.com/block/ftl="+ic.RootDir).RunBuffered(ic)
			assert.NoError(t, err)
		}
	}
}

// SetEnv sets an environment variable for the duration of the test.
//
// Note that the FTL controller will already be running.
func SetEnv(key string, value func(ic TestContext) string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Setting environment variable %s=%s", key, value(ic))
		t.Setenv(key, value(ic))
	}
}

// Copy a directory from the testdata directory to the working directory.
func CopyDir(src, dest string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Copying %s -> %s", src, dest)
		err := copy.Copy(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest))
		assert.NoError(t, err)
	}
}

// Chain multiple actions together.
func Chain(actions ...Action) Action {
	return func(t testing.TB, ic TestContext) {
		for _, action := range actions {
			action(t, ic)
		}
	}
}

// SubTests runs a list of individual actions as separate tests
func SubTests(tests ...SubTest) Action {
	return func(t testing.TB, ic TestContext) {
		for _, test := range tests {
			ic.Run(test.Name, func(t *testing.T) {
				t.Helper()
				ic.AssertWithRetry(t, test.Action)
			})
		}
	}
}

// Repeat an action N times.
func Repeat(n int, action Action) Action {
	return func(t testing.TB, ic TestContext) {
		for i := 0; i < n; i++ {
			action(t, ic)
		}
	}
}

// Chdir changes the test working directory to the subdirectory for the duration of the action.
func Chdir(dir string, a Action) Action {
	return func(t testing.TB, ic TestContext) {
		dir := filepath.Join(ic.workDir, dir)
		Infof("Changing directory to %s", dir)
		cwd, err := os.Getwd()
		assert.NoError(t, err)
		ic.workDir = dir
		err = os.Chdir(dir)
		assert.NoError(t, err)
		defer os.Chdir(cwd)
		a(t, ic)
	}
}

// DebugShell opens a new Terminal window in the test working directory.
func DebugShell() Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Starting debug shell")
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "open", "-n", "-W", "-a", "Terminal", ".").RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// Exec runs a command from the test working directory.
func Exec(cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing (in %s): %s %s", ic.workDir, cmd, shellquote.Join(args...))
		err := ftlexec.Command(ic, log.Debug, ic.workDir, cmd, args...).RunStderrError(ic)
		assert.NoError(t, err)
	}
}

// ExecWithExpectedOutput runs a command from the test working directory.
// The output is captured and is compared with the expected output.
func ExecWithExpectedOutput(want string, cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.NoError(t, err)
		assert.Equal(t, output, []byte(want))
	}
}

// ExecWithExpectedError runs a command from the test working directory, and
// expects it to fail with the given error message.
func ExecWithExpectedError(want string, cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.Error(t, err)
		assert.Contains(t, string(output), want)
	}
}

// ExecWithOutput runs a command from the test working directory.
// On success capture() is executed with the output
// On error, an error with the output is returned.
func ExecWithOutput(cmd string, args []string, capture func(output string)) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.NoError(t, err, "%s", string(output))
		capture(string(output))
	}
}

// ExpectError wraps an action and expects it to return an error containing the given messages.
func ExpectError(action Action, expectedErrorMsg ...string) Action {
	return func(t testing.TB, ic TestContext) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(TestingError); ok {
					for _, msg := range expectedErrorMsg {
						assert.Contains(t, string(e), msg)
					}
				} else {
					panic(r)
				}
			}
		}()
		action(t, ic)
	}
}

// Deploy a module from the working directory and wait for it to become available.
func Deploy(modules ...string) Action {
	return Chain(
		func(t testing.TB, ic TestContext) {
			args := []string{"deploy", "-t", "4m"}
			if ic.kubeClient.Ok() {
				args = append(args, "--build-env", "GOOS=linux", "--build-env", "GOARCH=amd64", "--build-env", "CGO_ENABLED=0")
			}
			args = append(args, modules...)

			Exec("ftl", args...)(t, ic)
		},
		func(t testing.TB, ic TestContext) {
			// Wait for all modules to deploy
			wg := &errgroup.Group{}
			for _, module := range modules {
				wg.Go(func() error {
					WaitWithTimeout(module, time.Minute*2)(t, ic)
					return nil
				})
			}
			assert.NoError(t, wg.Wait())
		},
		Chain(islices.Map(modules, func(module string) Action {
			return WaitWithTimeout(module, time.Minute*2)
		})...),
	)
}

// Deploy a module from the working directory and expect it to fail
func DeployBroken(module string) Action {
	return Chain(
		func(t testing.TB, ic TestContext) {
			args := []string{"deploy", "-t", "4m"}
			if ic.kubeClient.Ok() {
				args = append(args, "--build-env", "GOOS=linux", "--build-env", "GOARCH=amd64", "--build-env", "CGO_ENABLED=0")
			}
			args = append(args, module)
			ExecWithExpectedError("failed to deploy", "ftl", args...)(t, ic)
		},
		WaitWithTimeout(module, time.Minute),
	)
}

// Build modules from the working directory and wait for it to become available.
func Build(modules ...string) Action {
	args := []string{"build"}
	args = append(args, modules...)
	return Exec("ftl", args...)
}

// FtlNew creates a new FTL module
func FtlNew(language, name string) Action {
	return func(t testing.TB, ic TestContext) {
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "ftl", "module", "new", language, name, ic.workDir).RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// Wait for the given module to deploy.
func Wait(module string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Waiting for %s to become ready", module)
		// There's a bit of a bug here: wait() is already being retried by the
		// test harness, so in the error case we'll be waiting N^2 times. This
		// is fine for now, but we should fix this in the future.
		ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) {
			status, err := ic.Admin.GetSchema(ic, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
			assert.NoError(t, err)

			schema, err := schema.FromProto(status.Msg.GetSchema())
			assert.NoError(t, err)
			for _, deployment := range schema.InternalModules() {
				if deployment.Name == module {
					return
				}
			}
			t.Fatalf("deployment of module %q not found", module)
		})
	}
}

// WaitWithTimeout for the given module to deploy.
func WaitWithTimeout(module string, timeout time.Duration) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Waiting for %s to become ready", module)
		deadline := time.After(timeout)
		tick := time.NewTicker(time.Millisecond * 100)
		defer tick.Stop()
		for {
			select {
			case <-deadline:
				t.Fatalf("deployment of module %q not found", module)
				return
			case <-tick.C:
				status, err := ic.Admin.GetSchema(ic, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
				assert.NoError(t, err)
				sch, err := schema.FromProto(status.Msg.GetSchema())
				assert.NoError(t, err)

				for _, deployment := range sch.InternalModules() {
					if deployment.Name == module {
						return
					}
				}
			}
		}
	}
}

func WaitForDev(noErrors bool, msgAndArgs ...any) Action {
	return func(t testing.TB, ic TestContext) {
		if noErrors {
			Infof("Waiting for FTL Dev state with no errors")
		} else {
			Infof("Waiting for FTL Dev state with errors")
		}
		result, err := devstate.WaitForDevState(ic.Context, ic.BuildEngine, ic.Admin, true)
		assert.NoError(t, err, "failed to wait for dev state")

		var errs []builderrors.Error
		for _, m := range result.Modules {
			if m.Errors != nil {
				errs = append(errs, m.Errors...)
			}
		}
		if noErrors {
			assert.Zero(t, errs, msgAndArgs...)
		} else {
			assert.NotZero(t, errs, msgAndArgs...)
		}
	}
}

func Sleep(duration time.Duration) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Sleeping for %s", duration)
		time.Sleep(duration)
	}
}

// Assert that a file exists in the working directory.
func FileExists(path string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Checking that %s exists", path)
		_, err := os.Stat(filepath.Join(ic.workDir, path))
		assert.NoError(t, err)
	}
}

// Assert that a file exists and its content contains the given text.
//
// If "path" is relative it will be to the working directory.
func FileContains(path, needle string) Action {
	return func(t testing.TB, ic TestContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		Infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		assert.NoError(t, err)
		actual := string(data)
		assert.Contains(t, actual, needle)
	}
}

// Assert that a file exists and its content is equal to the given text.
//
// If "path" is relative it will be to the working directory.
func FileContent(path, expected string) Action {
	return func(t testing.TB, ic TestContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		Infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		assert.NoError(t, err)
		expected = strings.TrimSpace(expected)
		actual := strings.TrimSpace(string(data))
		assert.Equal(t, expected, actual)
	}
}

// WriteFile writes a file to the working directory.
func WriteFile(path string, content []byte) Action {
	return func(t testing.TB, ic TestContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		Infof("Writing to %s", path)
		err := os.WriteFile(absPath, content, 0600)
		assert.NoError(t, err)
	}
}

// CreateFile creates a file in a module
func CreateFile(module string, contents string, path ...string) Action {
	return func(t testing.TB, ic TestContext) {
		parts := []string{ic.workDir, module}
		parts = append(parts, path...)
		file := filepath.Join(parts...)
		Infof("Creating %s", file)
		err := os.WriteFile(file, []byte(contents), os.FileMode(0644)) //nolint:gosec
		assert.NoError(t, err)
	}
}

// EditFile edits a file in a module
func EditFile(module string, editFunc func([]byte) []byte, path ...string) Action {
	return func(t testing.TB, ic TestContext) {
		parts := []string{ic.workDir, module}
		parts = append(parts, path...)
		file := filepath.Join(parts...)
		Infof("Editing %s", file)
		contents, err := os.ReadFile(file)
		assert.NoError(t, err)
		contents = editFunc(contents)
		err = os.WriteFile(file, contents, os.FileMode(0))
		assert.NoError(t, err)
	}
}

// EditNamedFile edits a file in a based on the name without an extension. This allows for similar edits for Kotlin and Java files for example
func EditNamedFile(module string, fileName string, editFunc func([]byte) []byte) Action {
	return func(t testing.TB, ic TestContext) {
		EditFiles(module, func(path string, contents []byte) (bool, []byte) {
			_, file := filepath.Split(path)
			ext := filepath.Ext(file)
			if ext == ".class" {
				return false, nil
			}
			if ext != "" {
				file = file[0 : len(file)-len(ext)]
			}
			if fileName == file {
				Infof("Editing %s %s", path, fileName)
				return true, editFunc(contents)
			}
			return false, nil
		}, "")(t, ic)
	}
}

// EditFile edits a files in a modules directory
func EditFiles(module string, editFunc func(path string, contents []byte) (bool, []byte), path ...string) Action {
	return func(t testing.TB, ic TestContext) {
		parts := []string{ic.workDir, module}
		parts = append(parts, path...)
		dir := filepath.Join(parts...)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			assert.NoError(t, err)
			if info.IsDir() {
				return nil
			}
			contents, err := os.ReadFile(path)
			assert.NoError(t, err)
			ok, contents := editFunc(path, contents)
			if ok {
				Infof("Editing %s", path)
				err = os.WriteFile(path, contents, os.FileMode(0))
				assert.NoError(t, err)
			}
			return nil
		})

	}
}

// MoveFile moves a file within a module
func MoveFile(module, from, to string) Action {
	return func(t testing.TB, ic TestContext) {
		err := os.Rename(filepath.Join(ic.WorkingDir(), module, from), filepath.Join(ic.WorkingDir(), module, to))
		assert.NoError(t, err)
	}
}

// MkdirAll creates the given directory under the working dir
func MkdirAll(module string, dir string) Action {
	return func(t testing.TB, ic TestContext) {
		err := os.MkdirAll(filepath.Join(ic.WorkingDir(), module, dir), 0700)
		assert.NoError(t, err)
	}
}

// RemoveDir removes the given directory and all of its contents under the working dir
func RemoveDir(dir string) Action {
	return func(t testing.TB, ic TestContext) {
		err := os.RemoveAll(filepath.Join(ic.WorkingDir(), dir))
		assert.NoError(t, err)
	}
}

type Obj map[string]any

// Call a verb.
//
// "check" may be nil
func Call[Req any, Resp any](module, verb string, request Req, check func(t testing.TB, response Resp)) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Calling %s.%s", module, verb)
		assert.False(t, unicode.IsUpper([]rune(verb)[0]), "verb %q must start with an lowercase letter", verb)
		data, err := encoding.Marshal(request)
		assert.NoError(t, err)
		resp, err := ic.Verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
			Verb: &schemapb.Ref{Module: module, Name: verb},
			Body: data,
		}))
		assert.NoError(t, err)
		var response Resp
		assert.Zero(t, resp.Msg.GetError(), "verb failed: %s", resp.Msg.GetError().GetMessage())
		err = encoding.Unmarshal(resp.Msg.GetBody(), &response)
		assert.NoError(t, err)
		if check != nil {
			check(t, response)
		}
	}
}

// VerifyKubeState lets you test the current kube state
func VerifyKubeState(check func(ctx context.Context, t testing.TB, client kubernetes.Clientset)) Action {
	return func(t testing.TB, ic TestContext) {
		check(ic.Context, t, ic.kubeClient.MustGet())
	}
}

// VerifySchema lets you test the current schema
func VerifySchema(check func(ctx context.Context, t testing.TB, sch *schema.Schema)) Action {
	return func(t testing.TB, ic TestContext) {
		sch, err := ic.Admin.GetSchema(ic, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
		if err != nil {
			t.Errorf("failed to get schema: %v", err)
			return
		}
		schema, err := schema.FromProto(sch.Msg.GetSchema())
		if err != nil {
			t.Errorf("failed to parse schema: %v", err)
			return
		}
		check(ic.Context, t, schema)
	}
}

// VerifySchemaVerb lets you test the current schema for a specific verb
func VerifySchemaVerb(module string, verb string, check func(ctx context.Context, t testing.TB, schema *schemapb.Schema, verb *schemapb.Verb)) Action {
	return func(t testing.TB, ic TestContext) {
		resp, err := ic.Schema.GetSchema(ic, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
		if err != nil {
			t.Errorf("failed to get schema: %v", err)
			return
		}
		sch, err := schema.FromProto(resp.Msg.GetSchema())
		if err != nil {
			t.Errorf("failed to parse schema: %v", err)
			return
		}
		for _, m := range sch.InternalModules() {
			if m.Name == module {
				for _, v := range m.Decls {
					if v, ok := v.(*schema.Verb); ok {
						if v.Name == verb {
							check(ic.Context, t, sch.ToProto(), v.ToProto())
							return
						}
					}
				}
			}
		}
		t.Errorf("verb %s.%s not found in schema", module, verb)
	}
}

// VerifyTimeline lets you test the current timeline
func VerifyTimeline(limit int, filters []*timelinepb.TimelineQuery_Filter, check func(ctx context.Context, t testing.TB, events []*timelinepb.Event)) Action {
	return func(t testing.TB, ic TestContext) {
		resp, err := ic.Timeline.GetTimeline(ic, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Query: &timelinepb.TimelineQuery{
				Filters: filters,
				Limit:   int32(limit),
			}}))
		if err != nil {
			t.Errorf("failed to get timeline: %v", err)
			return
		}
		check(ic.Context, t, resp.Msg.Events)
	}
}

// DeleteOldTimelineEvents deletes old events from the timeline
func DeleteOldTimelineEvents(ageSeconds int, _type timelinepb.EventType, check func(ctx context.Context, t testing.TB, expectDeleted int, events []*timelinepb.Event)) Action {
	return func(t testing.TB, ic TestContext) {
		resp, err := ic.Timeline.DeleteOldEvents(ic, connect.NewRequest(&timelinepb.DeleteOldEventsRequest{
			EventType:  _type,
			AgeSeconds: int64(ageSeconds),
		}))
		if err != nil {
			t.Errorf("failed to delete old timeline events: %v", err)
			return
		}
		remaining, err := ic.Timeline.GetTimeline(ic, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Query: &timelinepb.TimelineQuery{
				Limit: 1000,
			},
		}))
		if err != nil {
			t.Errorf("failed to get timeline: %v", err)
			return
		}
		check(ic.Context, t, int(resp.Msg.DeletedCount), remaining.Msg.Events)
	}
}

// Fail expects the next action to Fail.
func Fail(next Action, msg string, args ...any) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Expecting failure of nested action")
		panicked := true
		defer func() {
			if !panicked {
				t.Fatalf("expected action to fail: "+msg, args...)
			} else {
				recover()
			}
		}()
		next(t, ic)
		panicked = false
	}
}

// fetched and returns a row's column values
func GetRow(t testing.TB, ic TestContext, database, query string, fieldCount int) []any {
	Infof("Querying %s: %s", database, query)
	db, err := sql.Open("pgx", dsn.PostgresDSN(database))
	assert.NoError(t, err)
	defer db.Close()
	actual := make([]any, fieldCount)
	for i := range actual {
		actual[i] = new(any)
	}
	err = db.QueryRowContext(ic, query).Scan(actual...)
	assert.NoError(t, err)
	for i := range actual {
		actual[i] = *actual[i].(*any)
	}
	return actual
}

// Query a single row from a database.
func QueryRow(database string, query string, expected ...interface{}) Action {
	return func(t testing.TB, ic TestContext) {
		actual := GetRow(t, ic, database, query, len(expected))
		for i, a := range actual {
			assert.Equal(t, expected[i], a)
		}
	}
}

func terminateDanglingConnections(t testing.TB, db *sql.DB, dbName string) {
	t.Helper()

	_, err := db.Exec(`
		SELECT pid, pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()`,
		dbName)
	assert.NoError(t, err)
}

func DropDBAction(t testing.TB, dbName string) Action {
	return func(t testing.TB, ic TestContext) {
		DropDB(t, dbName)
	}
}

func DropDB(t testing.TB, dbName string) {
	Infof("Dropping database %s", dbName)

	db, err := sql.Open("pgx", dsn.PostgresDSN("postgres"))
	if err != nil {
		// We just assume there is no DB running
		return
	}
	_, err = db.Exec("SELECT 1")
	if err != nil {
		// We just assume there is no DB running
		return
	}

	terminateDanglingConnections(t, db, dbName)

	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	assert.NoError(t, err, "failed to delete existing database")

	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})
}

// Create a directory in the working directory
func Mkdir(dir string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Creating directory %s", dir)
		err := os.MkdirAll(filepath.Join(ic.workDir, dir), 0700)
		assert.NoError(t, err)
	}
}

type HTTPResponse struct {
	Status    int
	Headers   map[string][]string
	JsonBody  map[string]any
	BodyBytes []byte
}

func JsonData(t testing.TB, body interface{}) []byte {
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	return b
}

// HttpCall makes an HTTP call to the running FTL ingress endpoint.
func HttpCall(method string, path string, headers map[string][]string, body []byte, onResponse func(t testing.TB, resp *HTTPResponse)) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("HTTP %s %s", method, path)
		baseURL, err := url.Parse("http://localhost:8891")
		assert.NoError(t, err)

		u, err := baseURL.Parse(path)
		assert.NoError(t, err)
		r, err := http.NewRequestWithContext(ic, method, u.String(), bytes.NewReader(body))
		assert.NoError(t, err)

		r.Header.Add("Content-Type", "application/json")
		for k, vs := range headers {
			for _, v := range vs {
				r.Header.Add(k, v)
			}
		}

		client := http.Client{}
		resp, err := client.Do(r)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var resBody map[string]any
		// ignore the error here since some responses are just `[]byte`.
		_ = json.Unmarshal(bodyBytes, &resBody)

		onResponse(t, &HTTPResponse{
			Status:    resp.StatusCode,
			Headers:   resp.Header,
			JsonBody:  resBody,
			BodyBytes: bodyBytes,
		})
	}
}

func IfLanguage(language string, action Action) Action {
	return IfLanguages(action, language)
}

func IfLanguages(action Action, languages ...string) Action {
	return func(t testing.TB, ic TestContext) {
		if slices.Contains(languages, ic.Language) {
			action(t, ic)
		}
	}
}

// Run "go test" in the given module.
func ExecModuleTest(module string) Action {
	return Chdir(module, Exec("go", "test", "./..."))
}
