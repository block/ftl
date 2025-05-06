//go:build integration

package app

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestConfigsWithController(t *testing.T) {
	in.Run(t, configActions(t)...)
}

func TestConfigsWithoutController(t *testing.T) {
	in.Run(t, configActions(t, in.WithoutController(), in.WithoutTimeline())...)
}

func configActions(t *testing.T, prepend ...in.ActionOrOption) []in.ActionOrOption {
	t.Helper()

	return append(prepend,
		// test setting value without --json flag
		in.Exec("ftl", "config", "set", "test.one", "hello world", "--inline"),
		in.ExecWithExpectedOutput("\"hello world\"\n", "ftl", "config", "get", "test.one"),
		// test updating value with --json flag
		in.Exec("ftl", "config", "set", "test.one", `"hello world 2"`, "--json", "--inline"),
		in.ExecWithExpectedOutput("\"hello world 2\"\n", "ftl", "config", "get", "test.one"),
		// test deleting value
		in.Exec("ftl", "config", "unset", "test.one", "--inline"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"config", "get", "test.one"}, func(output string) {}),
			"failed to get from config manager: not found",
		),
	)
}

func TestSecretsWithController(t *testing.T) {
	in.Run(t, secretActions(t)...)
}

func TestSecretsWithoutController(t *testing.T) {
	in.Run(t, secretActions(t, in.WithoutController(), in.WithoutTimeline())...)
}

func secretActions(t *testing.T, prepend ...in.ActionOrOption) []in.ActionOrOption {
	t.Helper()

	// can not easily use Exec() to enter secure text, using secret import instead
	secretsPath1, err := filepath.Abs("testdata/secrets1.json")
	assert.NoError(t, err)
	secretsPath2, err := filepath.Abs("testdata/secrets2.json")
	assert.NoError(t, err)

	return append(prepend,
		// test setting secret without --json flag
		in.Exec("ftl", "secret", "import", "--inline", secretsPath1),
		in.ExecWithExpectedOutput("\"hello world\"\n", "ftl", "secret", "get", "test.one"),
		// test updating secret
		in.Exec("ftl", "secret", "import", "--inline", secretsPath2),
		in.ExecWithExpectedOutput("\"hello world 2\"\n", "ftl", "secret", "get", "test.one"),
		// test deleting secret
		in.Exec("ftl", "secret", "unset", "test.one", "--inline"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"secret", "get", "test.one"}, func(output string) {}),
			"failed to get from secret manager: not found",
		),
	)
}

func TestSecretImportExport(t *testing.T) {
	testImportExport(t, "secret")
}

func TestConfigImportExport(t *testing.T) {
	testImportExport(t, "config")
}

func testImportExport(t *testing.T, object string) {
	t.Helper()

	firstProjFile := "ftl-project.toml"
	secondProjFile := "ftl-project-2.toml"
	destinationFile := "exported.json"

	importPath, err := filepath.Abs("testdata/import.json")
	assert.NoError(t, err)

	// use a pointer to keep track of the exported json so that i can be modified from within actions
	blank := ""
	exported := &blank

	in.Run(t,
		in.WithoutController(),
		in.WithoutTimeline(),
		// duplicate project file in the temp directory
		in.Exec("cp", firstProjFile, secondProjFile),
		// import into first project file
		in.Exec("ftl", object, "import", "--inline", "--config", firstProjFile, importPath),

		// export from first project file
		in.ExecWithOutput("ftl", []string{object, "export", "--config", firstProjFile}, func(output string) {
			*exported = output

			// make sure the exported json contains a value (otherwise the test could pass with the first import doing nothing)
			assert.Contains(t, output, "test.one")
		}),

		// import into second project file
		// wrapped in a func to avoid capturing the initial valye of *exported
		func(t testing.TB, ic in.TestContext) {
			in.WriteFile(destinationFile, []byte(*exported))(t, ic)
			in.Exec("ftl", object, "import", destinationFile, "--inline", "--config", secondProjFile)(t, ic)
		},

		// export from second project file
		in.ExecWithOutput("ftl", []string{object, "export", "--config", secondProjFile}, func(output string) {
			// check that both exported the same json
			assert.Equal(t, *exported, output)
		}),
	)
}

func TestLocalSchemaDiff(t *testing.T) {
	newVerb := `
//ftl:verb
func NewFunction(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}
`
	in.Run(t,
		in.CopyModule("time"),
		in.Deploy("time"),
		in.ExecWithOutput("ftl", []string{"schema", "diff"}, func(output string) {
			assert.Equal(t, "", output)
		}),
		in.EditFile("time", func(bytes []byte) []byte {
			s := string(bytes)
			s += newVerb
			return []byte(s)
		}, "time.go"),
		in.Build("time"),
		// We exit with code 1 when there is a difference
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"schema", "diff"}, func(output string) {
				assert.Contains(t, output, "-  verb newFunction(time.TimeRequest) time.TimeResponse")
			}), "exit status 1"),
	)
}
