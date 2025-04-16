package initialize

import (
	"go/token"
	"go/types"
	"reflect"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl-golang-tools/go/packages"
	"github.com/block/ftl/common/slices"
)

// Analyzer prepares data prior to the schema extractor runs, e.g. loads FTL types for reference by other
// analyzers.
var Analyzer = &analysis.Analyzer{
	Name:             "initialize",
	Doc:              "loads data to be used by other analyzers in the schema extractor pass",
	Run:              Run,
	ResultType:       reflect.TypeFor[Result](),
	RunDespiteErrors: true,
}

type Result struct {
	types map[string]*types.Interface
}

// IsStdlibErrorType will return true if the provided type is assertable to the stdlib `error` type.
func (r Result) IsStdlibErrorType(typ types.Type) bool {
	return r.assertableToType(typ, "builtin", "error")
}

// IsContextType will return true if the provided type is assertable to the `context.Context` type.
func (r Result) IsContextType(typ types.Type) bool {
	return r.assertableToType(typ, "context", "Context")
}

func (r Result) assertableToType(typ types.Type, pkg string, name string) bool {
	ityp, ok := r.types[pkg+"."+name]
	if !ok {
		return false
	}
	return types.AssertableTo(ityp, typ)
}

func Run(pass *analysis.Pass) (interface{}, error) {
	ctxType, err := loadRef(pass, "context", "Context")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	errType, err := loadRef(pass, "builtin", "error")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return Result{types: map[string]*types.Interface{
		"context.Context": ctxType,
		"builtin.error":   errType,
	}}, nil
}

// Lazy load the compile-time reference from a package. First attempts to derive the package from imports, then if not
// found attempts to load the package directly.
func loadRef(pass *analysis.Pass, pkg, name string) (*types.Interface, error) {
	var importedPkg *types.Package
	for _, p := range pass.Pkg.Imports() {
		if p.Path() == pkg {
			importedPkg = p
		}
	}

	// if the package is not imported, attempt to load it.
	if importedPkg == nil {
		pkgs, err := packages.Load(&packages.Config{Fset: token.NewFileSet(), Mode: packages.NeedTypes}, pkg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load package %q", pkg)
		}
		if len(pkgs) != 1 {
			return nil, errors.Errorf("expected one package, got %s",
				strings.Join(slices.Map(pkgs, func(p *packages.Package) string { return p.Name }), ", "))
		}
		importedPkg = pkgs[0].Types
	}
	if importedPkg == nil {
		return nil, errors.Errorf("package %q not found", pkg)
	}

	obj := importedPkg.Scope().Lookup(name)
	if obj == nil {
		return nil, errors.Errorf("interface %q not found", name)
	}
	ifaceType, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, errors.Errorf("expected an interface, got %s", obj.Type())
	}
	return ifaceType, nil
}
