package sql

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:embed resources/sqlc-gen-ftl.wasm
var embeddedResources embed.FS

func extractEmbeddedFile(resourceName, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	srcFile, err := embeddedResources.Open(filepath.Join("resources", resourceName))
	if err != nil {
		return fmt.Errorf("failed to open embedded resource %s: %w", resourceName, err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
