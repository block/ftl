//go:build !release

package sqlc

import (
	"archive/zip"

	"github.com/block/ftl/internal"
)

// Files is the SQLC template files.
func Files() *zip.Reader {
	return internal.ZipRelativeToCaller("template")
}
