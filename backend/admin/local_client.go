package admin

import (
	"context"
	"fmt"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/either"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/schema/builder"
	cf "github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
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
func NewLocalClient(projConfig projectconfig.Config, cm *manager.Manager[cf.Configuration], sm *manager.Manager[cf.Secrets]) EnvironmentClient {
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
	sch := builder.Schema()
	realmBuilder := builder.Realm(s.projConfig.Name)
	errs := []error{}
	for range len(modules) {
		result := <-moduleSchemas
		switch result := result.(type) {
		case either.Left[*schema.Module, error]:
			realmBuilder = realmBuilder.Module(result.Get())
		case either.Right[*schema.Module, error]:
			errs = append(errs, result.Get())
		default:
			panic(fmt.Sprintf("unexpected type %T", result))
		}
	}
	if len(errs) > 0 {
		return nil, errors.WithStack(errors.Join(errs...))
	}
	realm, err := realmBuilder.Build()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return errors.WithStack2(sch.Realm(realm).Build())
}
