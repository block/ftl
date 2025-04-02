package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

var buildDirName = ".ftl"

// GenerateStubs generates stubs for the given modules.
//
// Currently, only Go stubs are supported. Kotlin and other language stubs can be added in the future.
func GenerateStubs(ctx context.Context, projectRoot string, modules []*schema.Module, metas map[string]moduleMeta) error {
	err := generateStubsForEachLanguage(ctx, projectRoot, modules, metas)
	if err != nil {
		return err
	}
	return nil
}

// CleanStubs removes all generated stubs.
func CleanStubs(ctx context.Context, projectRoot string, configs []moduleconfig.UnvalidatedModuleConfig) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Deleting all generated stubs")
	sharedFtlDir := filepath.Join(projectRoot, buildDirName)

	// Figure out which languages we need to clean.
	languages := make(map[string]struct{})
	for _, config := range configs {
		languages[config.Language] = struct{}{}
	}

	for lang := range languages {
		stubsDir := filepath.Join(sharedFtlDir, lang, "modules")
		err := os.RemoveAll(stubsDir)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", stubsDir, err)
		}
	}

	return nil
}

// SyncStubReferences syncs the references in the generated stubs.
//
// For Go, this means updating all the go.work files to include all known modules in the shared stubbed modules directory.
func SyncStubReferences(ctx context.Context, projectRoot string, moduleNames []string, metas map[string]moduleMeta, view *schema.Schema) error {
	wg, wgctx := errgroup.WithContext(ctx)
	for _, meta := range metas {
		wg.Go(func() error {
			stubsRoot := stubsLanguageDir(projectRoot, meta.module.Config.Language)
			if err := meta.plugin.SyncStubReferences(wgctx, meta.module.Config, stubsRoot, moduleNames, view); err != nil {
				return fmt.Errorf("failed to sync go stub references for %s: %w", meta.module.Config.Module, err)
			}
			return nil
		})
	}
	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("failed to sync go stub references: %w", err)
	}
	return nil
}

func stubsLanguageDir(projectRoot, language string) string {
	return filepath.Join(projectRoot, buildDirName, language, "modules")
}

func stubsModuleDir(projectRoot, language, module string) string {
	return filepath.Join(stubsLanguageDir(projectRoot, language), module)
}

func generateStubsForEachLanguage(ctx context.Context, projectRoot string, modules []*schema.Module, metas map[string]moduleMeta) error {
	modulesByName := map[string]*schema.Module{}
	for _, module := range modules {
		modulesByName[module.Name] = module
	}
	metasByLanguage := map[string][]moduleMeta{}
	for _, meta := range metas {
		metasByLanguage[meta.module.Config.Language] = append(metasByLanguage[meta.module.Config.Language], meta)
	}
	wg, wgctx := errgroup.WithContext(ctx)
	for language, metasForLang := range metasByLanguage {
		for idx, module := range modules {
			// spread the load across plugins
			assignedMeta := metasForLang[idx%len(metasForLang)]
			config := metas[module.Name].module.Config
			wg.Go(func() error {
				path := stubsModuleDir(projectRoot, language, module.Name)
				var nativeConfig optional.Option[moduleconfig.ModuleConfig]
				if config.Module == "builtin" || config.Language != language {
					nativeConfig = optional.Some(assignedMeta.module.Config)
				}
				if err := assignedMeta.plugin.GenerateStubs(wgctx, path, module, config, nativeConfig); err != nil {
					return err //nolint:wrapcheck
				}
				return nil
			})
		}
	}
	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("failed to generate language stubs: %w", err)
	}
	return nil
}
