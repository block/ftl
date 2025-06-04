package deploymentcontext

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/errors"
)

func NewDiskProvider(path string) func(ctx context.Context) (map[string][]byte, error) {
	// GetSecrets reads secrets from individual files in the configured directory
	return func(ctx context.Context) (map[string][]byte, error) {
		data := make(map[string][]byte)

		entries, err := os.ReadDir(path)
		if err != nil {
			return data, errors.Wrapf(err, "failed to read directory %s", path)
		}

		for _, entry := range entries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") { // Skip directories and hidden files
				continue
			}
			filePath := filepath.Join(path, entry.Name())
			fileContentBytes, err := os.ReadFile(filePath)
			if err != nil {
				return data, errors.Wrapf(err, "failed to read file %s", filePath)
			}
			data[entry.Name()] = fileContentBytes
		}
		return data, nil
	}
}
