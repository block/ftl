package admin

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/either"

	"github.com/block/ftl/common/schema"
	configuration "github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

type diskSchemaRetriever struct {
	projConfig projectconfig.Config
}

func newDiskSchemaRetriever(projConfig projectconfig.Config) *diskSchemaRetriever {
	return &diskSchemaRetriever{projConfig: projConfig}
}

// NewLocalClient creates a admin client that reads and writes from the provided config and secret managers
func NewLocalClient(projConfig projectconfig.Config, cm configuration.Provider[configuration.Configuration], sm configuration.Provider[configuration.Secrets]) EnvironmentClient {
	return NewEnvironmentClient(cm, sm, newDiskSchemaRetriever(projConfig))
}

func (s *diskSchemaRetriever) GetSchema(ctx context.Context) (*schema.Schema, error) {
	modules, err := watch.DiscoverModules(ctx, s.projConfig.AbsModuleDirs())
	if err != nil {
		return nil, errors.Wrap(err, "could not discover modules")
	}

	moduleSchemas := make(chan either.Either[*schema.Module, error], 32)
	defer close(moduleSchemas)

	for _, m := range modules {
		go func() {
			module, err := schema.ModuleFromProtoFile(s.projConfig.SchemaPath(m.Module))
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](errors.Wrap(err, "could not load module schema"))
				return
			}
			moduleSchemas <- either.LeftOf[error](module)
		}()
	}
	realm := &schema.Realm{
		Name:    s.projConfig.Name,
		Modules: []*schema.Module{},
	}
	sch := &schema.Schema{Realms: []*schema.Realm{realm}}
	errs := []error{}
	for range len(modules) {
		result := <-moduleSchemas
		switch result := result.(type) {
		case either.Left[*schema.Module, error]:
			realm.Upsert(result.Get())
		case either.Right[*schema.Module, error]:
			errs = append(errs, result.Get())
		default:
			panic(fmt.Sprintf("unexpected type %T", result))
		}
	}
	if len(errs) > 0 {
		return nil, errors.WithStack(errors.Join(errs...))
	}
	return sch, nil
}
