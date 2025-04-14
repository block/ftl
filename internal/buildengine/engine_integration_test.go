//go:build integration

package buildengine_test

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/strcase"
	in "github.com/block/ftl/internal/integration"
)

func TestCycleDetection(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("depcycle1"),
		in.CopyModule("depcycle2"),

		in.ExpectError(
			in.Build("depcycle1", "depcycle2"),
			`detected a module dependency cycle that impacts these modules:`,
		),
	)
}

func TestInt64BuildError(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("integer"),

		in.ExpectError(
			in.Build("integer"),
			`unsupported type "int64" for field "Input"`,
			`unsupported type "int64" for field "Output"`,
		),
	)
}

// Tests how build engine reacts to a module changing its exported verbs.
func TestModuleInterfaceChanges(t *testing.T) {
	in.Run(t,
		in.WithDevMode(),
		in.WithLanguages("go", "kotlin"),

		in.CopyModule("parent"),
		in.CopyModule("child"),
		in.Wait("parent"),
		in.Wait("child"),

		in.WaitForDev(true, "should have no errors at the start"),

		updateVerb("parent", "verb1", "verb2"),
		in.WaitForDev(false, "child should fail to build because verb1 is now verb2"),
		updateVerb("child", "verb1", "verb2"),
		in.WaitForDev(true, "child should succeed because it now uses verb2"),
		updateVerb("parent", "verb2", "verb3"),
		in.WaitForDev(false, "child should fail to build because verb2 is now verb3"),
		updateVerb("parent", "verb3", "verb2"),
		in.WaitForDev(true, "child should now build because verb2 is back (verb3 reverted)"),
		updateVerb("parent", "verb2", "verb4"),
		updateVerb("child", "verb2", "verb4"),
		in.WaitForDev(true, "should have no errors after both modules updated to verb4"),
	)
}

// updateVerb can replace decls or clients
func updateVerb(module, old, new string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		var file string
		switch ic.Language {
		case "go":
			file = filepath.Join(".", module, module+".go")
		case "kotlin":
			file = filepath.Join(".", module, "src", "main", "kotlin", "ftl", module, strcase.ToUpperCamel(module)+".kt")
		}
		assert.NotEqual(t, "", file, "unsupported language: %s", ic.Language)

		in.Exec("sed", "-i.bak", "s/"+old+"/"+new+"/g", file)(t, ic)
		in.Exec("sed", "-i.bak", "s/"+strcase.ToUpperCamel(old)+"/"+strcase.ToUpperCamel(new)+"/g", file)(t, ic)
	}
}
