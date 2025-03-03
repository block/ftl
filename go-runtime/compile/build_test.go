package compile

import (
	"os"
	"path/filepath"
	stdslices "slices"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/common/strcase"
	goschema "github.com/block/ftl/go-runtime/schema"
	"github.com/block/ftl/go-runtime/schema/common"
	"github.com/block/ftl/internal/watch"
)

func TestImportAliases(t *testing.T) {
	actual, err := schema.ParseModuleString("", `
	module typealias {
		typealias FooBar1 Any
		+typemap go "github.com/one1/foo/bar/package.Type"

		typealias FooBar2 Any
		+typemap go "github.com/two2/foo/bar/package.Type"

		typealias Unique Any
		+typemap go "github.com/two2/foo/bar/unique.Type"

		typealias UniqueDir Any
		+typemap go "github.com/some/pkg.uniquedir.Type"

		typealias NonUniqueDir Any
		+typemap go "github.com/example/path/to/pkg.last.Type"

		typealias ConflictsWithDir Any
		+typemap go "github.com/last.Type"

		// import aliases can't have a number as the first character
		typealias StartsWithANumber1 Any
		+typemap go "github.com/11/numeric.Type"

		typealias StartsWithANumber2 Any
		+typemap go "github.com/22/numeric.Type"

		// two different directories with the same import path, first one wins
		typealias SamePackageDiffDir1 Any
		+typemap go "github.com/same.dir1.Type"

		typealias SamePackageDiffDir2 Any
		+typemap go "github.com/same.dir2.Type"

		// two aliases that are part of the same external package
		typealias TwoAliasesWithOnePkg1 Any
		+typemap go "github.com/two/aliaseswithonepkg.Type1"

		typealias TwoAliasesWithOnePkg2 Any
		+typemap go "github.com/two/aliaseswithonepkg.Type2"

		// references ftl/moduleclash, which is also the name of an external library
		export data ExampleData {
			something moduleclash.ExampleType
		}

		typealias ClashesWithModuleImport Any
		+typemap go "github.com/ftlmoduleclash.Type2"
	}
	`)
	assert.NoError(t, err)
	imports := imports(actual, false)
	assert.Equal(t, map[string]string{
		"github.com/one1/foo/bar/package":  "one1_foo_bar_package",
		"github.com/two2/foo/bar/package":  "two2_foo_bar_package",
		"github.com/two2/foo/bar/unique":   "unique",
		"github.com/some/pkg":              "uniquedir",
		"github.com/example/path/to/pkg":   "pkg_last",
		"github.com/last":                  "github_com_last",
		"github.com/11/numeric":            "_11_numeric",
		"github.com/22/numeric":            "_22_numeric",
		"github.com/same":                  "dir1",
		"github.com/two/aliaseswithonepkg": "aliaseswithonepkg",
		"ftl/moduleclash":                  "ftlmoduleclash",
		"github.com/ftlmoduleclash":        "github_com_ftlmoduleclash",
	}, imports)
}

func TestUpdateGoModuleValidatesModuleName(t *testing.T) {
	dir := t.TempDir()
	goModPath := filepath.Join(dir, "go.mod")

	// Matching module name
	err := os.WriteFile(goModPath, []byte(`module ftl/mymodule

go 1.24.0
`), 0600)
	assert.NoError(t, err)

	_, _, err = updateGoModule(goModPath, "mymodule", optional.None[watch.ModifyFilesTransaction]())
	assert.NoError(t, err)

	// Mismatched module name
	err = os.WriteFile(goModPath, []byte(`module ftl/wrongname

go 1.24.0
`), 0600)
	assert.NoError(t, err)

	_, _, err = updateGoModule(goModPath, "mymodule", optional.None[watch.ModifyFilesTransaction]())
	assert.Contains(t, err.Error(), "module name mismatch: expected 'ftl/mymodule' but got 'ftl/wrongname'")
}

func TestCylicVerbs(t *testing.T) {
	module, err := schema.ParseModuleString("", `
module cyclic {
  verb a(Unit) Unit
    +calls cyclic.b

  verb b(Unit) Unit
    +calls cyclic.c

  verb c(Unit) Unit
    +calls cyclic.a
}
	`)
	assert.NoError(t, err)
	verbDecls := stdslices.Collect(slices.FilterVariants[*schema.Verb](module.Decls))
	_, err = buildMainDeploymentContext(&schema.Schema{}, goschema.Result{
		Module: module,
		VerbResourceParamOrder: map[*schema.Verb][]common.VerbResourceParam{
			verbDecls[0]: {},
			verbDecls[1]: {},
			verbDecls[2]: {},
		},
		NativeNames: goschema.NativeNames{
			verbDecls[0]: "ftl/cyclic." + strcase.ToUpperCamel(module.Decls[0].GetName()),
			verbDecls[1]: "ftl/cyclic." + strcase.ToUpperCamel(module.Decls[1].GetName()),
			verbDecls[2]: "ftl/cyclic." + strcase.ToUpperCamel(module.Decls[2].GetName()),
		},
	},
		"", "projectname", nil, nil)

	assert.EqualError(t, err, "cyclic references are not allowed: cyclic.a refers to cyclic.b refers to cyclic.c refers to cyclic.a")
}
