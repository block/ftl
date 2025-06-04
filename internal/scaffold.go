package internal

import (
	"archive/zip"
	"os"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/block/scaffolder"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/strcase"
)

// ScaffoldZip is a convenience function for scaffolding a zip archive with scaffolder.
func ScaffoldZip(source *zip.Reader, destination string, ctx any, options ...scaffolder.Option) error {
	tmpDir, err := os.MkdirTemp("", "scaffold-")
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := UnzipDir(source, tmpDir); err != nil {
		return errors.WithStack(err)
	}
	options = append(options, scaffolder.Functions(scaffoldFuncs))
	return errors.WithStack(scaffolder.Scaffold(tmpDir, destination, ctx, options...))
}

var scaffoldFuncs = scaffolder.FuncMap{
	"snake":          strcase.ToLowerSnake,
	"screamingSnake": strcase.ToUpperSnake,
	"camel":          strcase.ToUpperCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"strippedCamel":  strcase.ToUpperStrippedCamel,
	"kebab":          strcase.ToLowerKebab,
	"screamingKebab": strcase.ToUpperKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename":       schema.TypeName,
}
