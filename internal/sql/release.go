//go:build release

package sql

import (
	"archive/zip"
	"bytes"
	_ "embed"
)

//go:embed template.zip
var archive []byte

// Files is the SQLC template files.
func Files() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		panic(err)
	}
	return zr
}
