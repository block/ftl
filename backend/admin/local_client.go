package admin

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/errors"
	"github.com/block/ftl/common/schema"
	cf "github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

type diskSchemaRetriever struct {
	// Omit to use the project root as the deploy root (used in tests)
	deployRoot optional.Option[string]
}

// NewLocalClient creates a admin client that reads and writes from the provided config and secret managers
func NewLocalClient(cm *manager.Manager[cf.Configuration], sm *manager.Manager[cf.Secrets]) EnvironmentClient {
	return NewEnvironmentClient(cm, sm, &diskSchemaRetriever{})
}

func (s *diskSchemaRetriever) GetSchema(ctx context.Context) (*schema.Schema, error) {
	path, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return nil, fmt.Errorf("no project config path available")
	}
	projConfig, err := projectconfig.Load(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("could not load project config: %w", err)
	}
	modules, err := watch.DiscoverModules(ctx, projConfig.AbsModuleDirs())
	if err != nil {
		return nil, fmt.Errorf("could not discover modules: %w", err)
	}

	moduleSchemas := make(chan either.Either[*schema.Module, error], 32)
	defer close(moduleSchemas)

	for _, m := range modules {
		go func() {
			module, err := schema.ModuleFromProtoFile(projConfig.SchemaPath(m.Module))
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](fmt.Errorf("could not load module schema: %w", err))
				return
			}
			moduleSchemas <- either.LeftOf[error](module)
		}()
	}
	sch := &schema.Schema{}
	errs := []error{}
	for range len(modules) {
		result := <-moduleSchemas
		switch result := result.(type) {
		case either.Left[*schema.Module, error]:
			sch.Upsert(result.Get())
		case either.Right[*schema.Module, error]:
			errs = append(errs, result.Get())
		default:
			panic(fmt.Sprintf("unexpected type %T", result))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return sch, nil
}
