package compile

import (
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/errors"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/watch"
)

func ExtractDependencies(config moduleconfig.AbsModuleConfig) ([]string, error) {
	deps, _, err := extractDependenciesAndImports(config)
	return deps, errors.WithStack(err)
}

func extractDependenciesAndImports(config moduleconfig.AbsModuleConfig) (deps []string, imports []string, err error) {
	importsMap := map[string]bool{}
	dependencies := map[string]bool{}
	fset := token.NewFileSet()
	err = watch.WalkDir(config.Dir, true, func(path string, d fs.DirEntry) error {
		if !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), "_") || d.Name() == "testdata" {
			return errors.WithStack(watch.ErrSkip)
		}
		pkgs, err := parser.ParseDir(fset, path, isNotGenerated, parser.ImportsOnly)
		if pkgs == nil {
			return errors.Wrap(err, "could parse directory in search of dependencies")
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					path, err := strconv.Unquote(imp.Path.Value)
					if err != nil {
						continue
					}
					importsMap[path] = true
					if !strings.HasPrefix(path, "ftl/") {
						continue
					}
					module := strings.Split(strings.TrimPrefix(path, "ftl/"), "/")[0]
					if module == config.Module {
						continue
					}
					dependencies[module] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "%s: failed to extract dependencies from Go module", config.Module)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	imports = maps.Keys(importsMap)
	sort.Strings(imports)
	return modules, imports, nil
}

func isNotGenerated(fi fs.FileInfo) bool {
	return fi.Name() != "types.ftl.go" && fi.Name() != "queries.ftl.go"
}
