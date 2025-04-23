//go:build integration

package goplugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestGoBuildClearsBuildDir(t *testing.T) {
	file := "./another/.ftl/test-clear-build.tmp"
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("another"),
		in.Build("another"),
		in.WriteFile(file, []byte{1}),
		in.FileExists(file),
		in.Build("another"),
		in.ExpectError(in.FileExists(file), "no such file"),
	)
}

func TestExternalType(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("external"),
		in.ExpectError(in.Build("external"),
			`unsupported type "time.Month" for field "Month"`,
			`unsupported external type "time.Month"; see FTL docs on using external types: block.github.io/ftl/docs/reference/externaltypes/`,
			`unsupported response type "ftl/external.ExternalResponse"`,
		),
	)
}

func TestGeneratedTypeRegistry(t *testing.T) {
	t.Skip("Skipping until there has been a release with the package change")
	expected, err := os.ReadFile("testdata/type_registry_main.go")
	assert.NoError(t, err)

	file := "other/.ftl/go/main/main.go"

	in.Run(t,
		in.WithTestDataDir("testdata"),
		// Deploy dependency
		in.CopyModule("another"),
		in.Deploy("another"),
		// Build the module under test
		in.CopyModule("other"),
		in.ExpectError(in.FileExists(file), "no such file"),
		in.Build("other"),
		// Validate the generated main.go
		in.FileContent(file, string(expected)),
	)
}

// TestGoSQLInterfaces tests the generation of SQL interfaces for Go modules
//
// This is used by the MCP to notify the LLM of changes to the sql interface after updating queries or schema.
func TestGoSQLInterfaces(t *testing.T) {
	in.Run(t,
		in.WithoutController(),
		in.WithoutTimeline(),
		in.WithTestDataDir("testdata"),
		in.CopyModule("mysql"),
		in.Build("mysql"),
		func(t testing.TB, ic in.TestContext) {
			expected := map[string]string{
				"CreateDemoRowClient": `type CreateDemoRowClient func(context.Context, CreateDemoRowQuery) error`,
				"ListDemoRowsClient":  `type ListDemoRowsClient func(context.Context) ([]Demo, error)`,
				"CreateDemoRowQuery": `type CreateDemoRowQuery struct {
	RequiredString string
	OptionalString ftl.Option[string]
	NumberValue    int
	TimestampValue stdtime.Time
	FloatValue     float64
}`,
				"Demo": `type Demo struct {
	Id             int
	RequiredString string
	OptionalString ftl.Option[string]
	NumberValue    int
	TimestampValue stdtime.Time
	FloatValue     float64
}`,
			}

			result, err := sqlInterfaces(filepath.Join(ic.WorkingDir(), "mysql"))
			assert.NoError(t, err)
			resultMap := make(map[string]string)
			for _, r := range result {
				resultMap[r.Name] = r.Interface
			}
			assert.Equal(t, expected, resultMap, "Generated interfaces do not match expected values")

			// interfacesForGeneratedFiles should not return an error if the generated files do not exist yet
			result, err = sqlInterfaces(filepath.Join(ic.WorkingDir(), "nonexistent"))
			assert.NoError(t, err)
			assert.Zero(t, len(result), "Expected no interfaces for nonexistent module")
		},
	)
}
