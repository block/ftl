package deploymentcontext

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/block/ftl/common/log"
)

type diskSecretsProvider struct {
	secretsPath string
}

type diskConfigProvider struct {
	configsPath string
}

func NewDiskSecretsProvider(path string) SecretsProvider {
	provider := &diskSecretsProvider{
		secretsPath: path,
	}
	return provider.GetSecrets
}

// GetSecrets reads secrets from individual files in the configured directory
func (p *diskSecretsProvider) GetSecrets(ctx context.Context) map[string][]byte {
	logger := log.FromContext(ctx)
	secrets := make(map[string][]byte)

	entries, err := os.ReadDir(p.secretsPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Secrets directory %s does not exist", p.secretsPath)
			return secrets
		}
		logger.Errorf(err, "Failed to read secrets directory %s", p.secretsPath)
		return secrets
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") { // Skip directories and hidden files
			continue
		}
		filePath := filepath.Join(p.secretsPath, entry.Name())
		fileContentBytes, err := os.ReadFile(filePath)
		if err != nil {
			logger.Errorf(err, "Failed to read secret file %s", filePath)
			continue
		}
		secrets[entry.Name()] = fileContentBytes
	}
	return secrets
}

func NewDiskConfigProvider(path string) ConfigProvider {
	provider := &diskConfigProvider{
		configsPath: path,
	}
	return provider.GetConfig
}

// GetConfig reads configs from individual files in the configured directory
func (p *diskConfigProvider) GetConfig(ctx context.Context) map[string][]byte {
	logger := log.FromContext(ctx)
	configs := make(map[string][]byte)

	entries, err := os.ReadDir(p.configsPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Configs directory %s does not exist", p.configsPath)
			return configs
		}
		logger.Errorf(err, "Failed to read configs directory %s", p.configsPath)
		return configs
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") { // Skip directories and hidden files
			continue
		}
		filePath := filepath.Join(p.configsPath, entry.Name())
		fileContentBytes, err := os.ReadFile(filePath)
		if err != nil {
			logger.Errorf(err, "Failed to read config file %s", filePath)
			continue
		}
		// ASSUMPTION: fileContentBytes from os.ReadFile is already the
		// JSON-encoded value (e.g., a json.RawMessage if the individual config files are fragments of a larger JSON structure,
		// or if the file itself contains a pre-encoded JSON string like "\"myValue\"")
		// ready for FTL's unmarshaler.
		configs[entry.Name()] = fileContentBytes
	}
	return configs
}
