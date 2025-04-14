package sql

import (
	"embed"
	"io"
	"os"
	"path/filepath"

	errors "github.com/alecthomas/errors"
)

//go:embed resources/sqlc-gen-ftl.wasm
var embeddedResources embed.FS

func extractEmbeddedFile(resourceName, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	srcFile, err := embeddedResources.Open(filepath.Join("resources", resourceName))
	if err != nil {
		return errors.Wrapf(err, "failed to open embedded resource %s", resourceName)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to create destination file")
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return errors.Wrap(err, "failed to copy file contents")
	}

	return nil
}
