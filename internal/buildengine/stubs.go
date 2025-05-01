package buildengine

import (
	"context"
	"github.com/block/ftl/internal/buildengine/module"
	"os"
	"path/filepath"

	errors "github.com/alecthomas/errors"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

var buildDirName = ".ftl"

// GenerateStubs generates stubs for the given modules.
//
// Currently, only Go stubs are supported. Kotlin and other language stubs can be added in the future.
func GenerateStubs(ctx context.Context, projectRoot string, modules []*schema.Module, metas map[string]module.EngineModule) error {
	err := generateStubsForEachLanguage(ctx, projectRoot, modules, metas)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CleanStubs removes all generated stubs.
func CleanStubs(ctx context.Context, projectRoot string, configs []moduleconfig.UnvalidatedModuleConfig) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Deleting all generated stubs")
	sharedFtlDir := filepath.Join(projectRoot, buildDirName)

	resourcesDir := filepath.Join(sharedFtlDir, "resources")
	err := os.RemoveAll(resourcesDir)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to remove %s", resourcesDir)
	}

	// Figure out which languages we need to clean.
	languages := make(map[string]struct{})
	for _, config := range configs {
		languages[config.Language] = struct{}{}
	}

	for lang := range languages {
		stubsDir := filepath.Join(sharedFtlDir, lang, "modules")
		err := os.RemoveAll(stubsDir)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to remove %s", stubsDir)
		}
	}

	return nil
}

// SyncStubReferences syncs the references in the generated stubs.
//
// For Go, this means updating all the go.work files to include all known modules in the shared stubbed modules directory.
func SyncStubReferences(ctx context.Context, projectRoot string, moduleNames []string, metas map[string]module.EngineModule, view *schema.Schema) error {
	wg, wgctx := errgroup.WithContext(ctx)
	for _, meta := range metas {
		wg.Go(func() error {
			return meta.SyncStubReferences(wgctx, projectRoot, moduleNames, view)
		})
	}
	err := wg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to sync go stub references")
	}
	return nil
}

func generateStubsForEachLanguage(ctx context.Context, projectRoot string, modules []*schema.Module, metas map[string]module.EngineModule) error {
	modulesByName := map[string]*schema.Module{}
	for _, module := range modules {
		modulesByName[module.Name] = module
	}
	metasByLanguage := map[string][]module.EngineModule{}
	for _, meta := range metas {
		metasByLanguage[meta.Language()] = append(metasByLanguage[meta.Language()], meta)
	}
	wg, wgctx := errgroup.WithContext(ctx)
	for _, metasForLang := range metasByLanguage {
		for idx, module := range modules {
			// spread the load across plugins
			assignedMeta := metasForLang[idx%len(metasForLang)]
			meta := metas[module.Name]
			wg.Go(func() error {
				if err := assignedMeta.GenerateStubs(wgctx, projectRoot, &meta, module); err != nil {
					return errors.WithStack(err) //nolint:wrapcheck
				}
				return nil
			})
		}
	}
	err := wg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to generate language stubs")
	}
	return nil
}
