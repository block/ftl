//go:build integration

package compile_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestNonExportedDecls(t *testing.T) {
	in.Run(t,
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("notexportedverb"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"deploy", "notexportedverb"}, func(_ string) {}),
			`unsupported verb parameter type; verbs must have the signature func(Context, Request?, Resources...)`,
		),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	in.Run(t,
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("undefinedverb"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"deploy", "undefinedverb"}, func(_ string) {}),
			`unsupported verb parameter type; verbs must have the signature func(Context, Request?, Resources...)`,
		),
	)
}

func TestNonFTLTypes(t *testing.T) {
	in.Run(t,
		in.CopyModule("external"),
		in.Deploy("external"),
		in.Call("external", "echo", in.Obj{"message": "hello"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["message"])
		}),
	)
}

func TestNonStructRequestResponse(t *testing.T) {
	t.Skip("Skipping until there has been a release with the package change")
	in.Run(t,
		in.CopyModule("two"),
		in.Deploy("two"),
		in.CopyModule("one"),
		in.Deploy("one"),
		in.Call("one", "stringToTime", "1985-04-12T23:20:50.52Z", func(t testing.TB, response time.Time) {
			assert.Equal(t, time.Date(1985, 04, 12, 23, 20, 50, 520_000_000, time.UTC), response)
		}),
	)
}

// TestBuildWithoutQueries ensures that the build command does not fail when queries.ftl.go is missing.
func TestBuildWithoutQueries(t *testing.T) {
	in.Run(t,
		in.CopyModule("missingqueries"),
		func(t testing.TB, ic in.TestContext) {
			assert.NoError(t, os.Remove(filepath.Join(ic.WorkingDir(), "missingqueries/queries.ftl.go")))
		},
		in.Fail(in.FileExists("missingqueries/queries.ftl.go"), "queries.ftl.go should not exist yet"),
		in.Deploy("missingqueries"),
		in.FileExists("missingqueries/queries.ftl.go"),
	)
}
