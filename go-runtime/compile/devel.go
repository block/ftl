//go:build !release

package compile

import (
	"archive/zip"

	"github.com/block/ftl/internal"
)

func mainWorkTemplateFiles() *zip.Reader {
	return internal.ZipRelativeToCaller("main-work-template")
}

func externalModuleTemplateFiles() *zip.Reader {
	return internal.ZipRelativeToCaller("external-module-template")
}

func buildTemplateFiles() *zip.Reader {
	return internal.ZipRelativeToCaller("build-template")
}

func queriesTemplateFiles() *zip.Reader {
	return internal.ZipRelativeToCaller("queries-template")
}
