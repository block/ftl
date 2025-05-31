package deploymentcontext

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
)

type diskSecretsProvider struct {
	secretsPath string
}

type diskConfigProvider struct {
	configsPath string
}

type diskRouteProvider struct {
	schemaPath string
	schema     atomic.Value[*schema.Schema]
}

var _ RouteProvider = (*diskRouteProvider)(nil)

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
		// Marshal the file content as a JSON string before storing
		// This ensures that if the file contains "foo", it's stored as "\"foo\""
		// so that GetSecret(..., &myString) correctly unmarshals it.
		jsonEncodedValue, err := json.Marshal(string(fileContentBytes))
		if err != nil {
			logger.Errorf(err, "Failed to JSON marshal secret value from file %s", filePath)
			continue
		}
		secrets[entry.Name()] = jsonEncodedValue
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
		// Marshal the file content as a JSON string
		jsonEncodedValue, err := json.Marshal(string(fileContentBytes))
		if err != nil {
			logger.Errorf(err, "Failed to JSON marshal config value from file %s", filePath)
			continue
		}
		configs[entry.Name()] = jsonEncodedValue
	}
	return configs
}

// NewDiskRouteProvider creates a new RouteProvider that reads from a schema file on disk
func NewDiskRouteProvider(ctx context.Context, schemaPath string) (RouteProvider, error) {
	logger := log.FromContext(ctx)

	provider := &diskRouteProvider{
		schemaPath: schemaPath,
	}

	if schemaPath != "" {
		sch, err := provider.loadSchema() // loadSchema reads from config.SchemaPath
		if err != nil {
			logger.Warnf("Failed to load initial schema from %s: %v", schemaPath, err)
		} else {
			provider.schema.Store(sch)
			logger.Debugf("Successfully loaded initial schema from %s", schemaPath)
		}
	}

	return provider, nil
}

func (p *diskRouteProvider) Route(module string) string {
	sch := p.schema.Load()
	if sch == nil {
		return ""
	}

	routes := make(map[string]string)
	for _, realm := range sch.Realms {
		for _, mod := range realm.Modules {
			for _, verb := range mod.Verbs() {
				for _, metadata := range verb.Metadata {
					if ingress, ok := metadata.(*schema.MetadataIngress); ok {
						routes[mod.Name] = ingress.Method + " " + ingress.PathString()
						break
					}
				}
			}
		}
	}

	return routes[module]
}

func (p *diskRouteProvider) Subscribe() chan string {
	return nil
}

func (p *diskRouteProvider) Unsubscribe(c chan string) {
}

func (p *diskRouteProvider) loadSchema() (*schema.Schema, error) {
	if p.schemaPath == "" {
		return nil, errors.New("schema path not provided")
	}

	data, err := os.ReadFile(p.schemaPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read schema file %s", p.schemaPath)
	}

	schemaProto := &schemapb.Schema{}
	err = proto.Unmarshal(data, schemaProto)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal schema file %s", p.schemaPath)
	}

	sch, err := schema.FromProto(schemaProto)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse schema file %s", p.schemaPath)
	}

	return sch, nil
}
