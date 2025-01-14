package compile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	stdreflect "reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/block/scaffolder"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	extract "github.com/block/ftl/go-runtime/schema"
	"github.com/block/ftl/go-runtime/schema/common"
	"github.com/block/ftl/go-runtime/schema/finalize"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

var ErrInvalidateDependencies = errors.New("dependencies need to be updated")
var ftlTypesFilename = "types.ftl.go"

type MainWorkContext struct {
	GoVersion          string
	SharedModulesPaths []string
	IncludeMainPackage bool
}

type mainDeploymentContext struct {
	GoVersion          string
	FTLVersion         string
	Name               string
	SharedModulesPaths []string
	Verbs              []goVerb
	Databases          []goDBHandle
	Replacements       []*modfile.Replace
	MainCtx            mainFileContext
	TypesCtx           typesFileContext
}

func (c *mainDeploymentContext) withImports(mainModuleImport string) {
	c.MainCtx.Imports = c.generateMainImports()
	c.TypesCtx.Imports = c.generateTypesImports(mainModuleImport)
}

func (c *mainDeploymentContext) generateMainImports() []string {
	imports := sets.NewSet[string]()
	imports.Add(`"context"`)
	imports.Add(`"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"`)
	imports.Add(`"github.com/block/ftl/common/plugin"`)
	imports.Add(`"github.com/block/ftl/go-runtime/server"`)
	if len(c.MainCtx.SumTypes) > 0 || len(c.MainCtx.ExternalTypes) > 0 {
		imports.Add(`"github.com/block/ftl/common/reflection"`)
	}

	for _, v := range c.Verbs {
		imports.Append(verbImports(v)...)
	}
	for _, st := range c.MainCtx.SumTypes {
		imports.Add(st.importStatement())
		for _, v := range st.Variants {
			imports.Add(v.importStatement())
		}
	}
	for _, e := range c.MainCtx.ExternalTypes {
		imports.Add(e.importStatement())
	}
	out := imports.ToSlice()
	slices.Sort(out)
	return out
}

func (c *mainDeploymentContext) generateTypesImports(mainModuleImport string) []string {
	imports := sets.NewSet[string]()
	if len(c.TypesCtx.SumTypes) > 0 || len(c.TypesCtx.ExternalTypes) > 0 {
		imports.Add(`"github.com/block/ftl/common/reflection"`)
	}
	if len(c.Verbs) > 0 {
		imports.Add(`"context"`)
	}
	if len(c.Databases) > 0 {
		imports.Add(`"github.com/block/ftl/go-runtime/server"`)
	}
	for _, st := range c.TypesCtx.SumTypes {
		imports.Add(st.importStatement())
		for _, v := range st.Variants {
			imports.Add(v.importStatement())
		}
	}
	for _, et := range c.TypesCtx.ExternalTypes {
		imports.Add(et.importStatement())
	}
	for _, v := range c.Verbs {
		if len(v.Resources) > 0 {
			imports.Add(`"github.com/block/ftl/go-runtime/server"`)
		}
		imports.Append(verbImports(v)...)
	}

	var filteredImports []string
	for _, im := range imports.ToSlice() {
		if im == mainModuleImport {
			continue
		}
		filteredImports = append(filteredImports, im)
	}
	slices.Sort(filteredImports)
	return filteredImports
}

func typeImports(t goSchemaType) []string {
	imports := sets.NewSet[string]()
	if nt, ok := t.nativeType.Get(); ok {
		imports.Add(nt.importStatement())
	}
	for _, c := range t.children {
		imports.Append(typeImports(c)...)
	}
	return imports.ToSlice()
}

func verbImports(v goVerb) []string {
	imports := sets.NewSet[string]()
	imports.Add(v.importStatement())
	imports.Add(`"github.com/block/ftl/common/reflection"`)

	if nt, ok := v.Request.nativeType.Get(); ok && v.Request.TypeName != "ftl.Unit" {
		imports.Add(nt.importStatement())
	}
	if nt, ok := v.Response.nativeType.Get(); ok && v.Response.TypeName != "ftl.Unit" {
		imports.Add(nt.importStatement())
	}
	for _, r := range v.Request.children {
		imports.Append(typeImports(r)...)
	}
	for _, r := range v.Response.children {
		imports.Append(typeImports(r)...)
	}

	for _, r := range v.Resources {
		switch r := r.(type) {
		case verbClient:
			imports.Add(`"github.com/block/ftl/go-runtime/server"`)
			imports.Append(verbImports(r.goVerb)...)
		case goTopicHandle:
			imports.Add(r.MapperType.importStatement())
		}
	}
	return imports.ToSlice()
}

type mainFileContext struct {
	Imports []string

	ProjectName   string
	SumTypes      []goSumType
	ExternalTypes []goExternalType
}

type typesFileContext struct {
	Imports       []string
	MainModulePkg string

	SumTypes      []goSumType
	ExternalTypes []goExternalType
}

type goType interface {
	getNativeType() nativeType
}

type nativeType struct {
	Name       string
	pkg        string
	importPath string
	// true if the package name differs from the directory provided by the import path
	importAlias bool
}

func (n nativeType) importStatement() string {
	if n.importAlias {
		return fmt.Sprintf("%s %q", n.pkg, n.importPath)
	}
	return strconv.Quote(n.importPath)
}

func (n nativeType) TypeName() string {
	return n.pkg + "." + n.Name
}

type goVerb struct {
	Request   goSchemaType
	Response  goSchemaType
	Resources []verbResource

	nativeType
}

type goSchemaType struct {
	TypeName      string
	LocalTypeName string
	children      []goSchemaType

	nativeType optional.Option[nativeType]
}

func (g goVerb) getNativeType() nativeType { return g.nativeType }

type goExternalType struct {
	nativeType
}

func (g goExternalType) getNativeType() nativeType { return g.nativeType }

type goSumType struct {
	Variants []goSumTypeVariant

	nativeType
}

func (g goSumType) getNativeType() nativeType { return g.nativeType }

type goSumTypeVariant struct {
	Type goSchemaType

	nativeType
}

func (g goSumTypeVariant) getNativeType() nativeType { return g.nativeType }

type verbResource interface {
	resource()
}

type verbClient struct {
	goVerb
}

func (v verbClient) resource() {}

type goDBHandle struct {
	Type   string
	Name   string
	Module string

	nativeType
}

func (d goDBHandle) resource() {}

func (d goDBHandle) getNativeType() nativeType {
	return d.nativeType
}

type goTopicHandle struct {
	Name      string
	Module    string
	EventType goSchemaType

	// Types for the topics partition mapper
	// TODO: we should support multiple levels of associated types, rather than just one level.
	MapperType           nativeType
	MapperAssociatedType optional.Option[nativeType]

	nativeType
}

func (d goTopicHandle) resource() {}

func (d goTopicHandle) getNativeType() nativeType {
	return d.nativeType
}

func (d goTopicHandle) MapperTypeName(trimModuleName string) string {
	name := trimModuleQualifier(trimModuleName, d.MapperType.TypeName())
	if at, ok := d.MapperAssociatedType.Get(); ok {
		name += "[" + trimModuleQualifier(trimModuleName, at.TypeName()) + "]"
	}
	return name
}

type goConfigHandle struct {
	Name   string
	Module string
	Type   goSchemaType

	nativeType
}

func (c goConfigHandle) resource() {}

func (c goConfigHandle) getNativeType() nativeType {
	return c.nativeType
}

type goSecretHandle struct {
	Module string
	Name   string
	Type   goSchemaType

	nativeType
}

func (s goSecretHandle) resource() {}

func (s goSecretHandle) getNativeType() nativeType {
	return s.nativeType
}

const buildDirName = ".ftl"

func buildDir(moduleDir string) string {
	return filepath.Join(moduleDir, buildDirName)
}

// OngoingState maintains state between builds, allowing the Build function to skip steps if nothing has changed.
type OngoingState struct {
	imports   []string
	moduleCtx mainDeploymentContext
}

func (s *OngoingState) checkIfImportsChanged(imports []string) (changed bool) {
	if slices.Equal(s.imports, imports) {
		return false
	}
	s.imports = imports
	return true
}

func (s *OngoingState) checkIfMainDeploymentContextChanged(moduleCtx mainDeploymentContext) (changed bool) {
	if stdreflect.DeepEqual(s.moduleCtx, moduleCtx) {
		return false
	}
	s.moduleCtx = moduleCtx
	return true
}

// DetectedFileChanges should be called whenever file changes are detected outside of the Build() function.
// This allows the OngoingState to detect if files need to be reprocessed.
func (s *OngoingState) DetectedFileChanges(config moduleconfig.AbsModuleConfig, changes []watch.FileChange) {
	paths := []string{
		filepath.Join(config.Dir, ftlTypesFilename),
		filepath.Join(config.Dir, "go.mod"),
		filepath.Join(config.Dir, "go.sum"),
	}
	for _, change := range changes {
		if !slices.Contains(paths, change.Path) {
			continue
		}
		// If files altered by Build() have been manually changed, reset state to make sure we correct them if needed.
		s.reset()
		return
	}
}

func (s *OngoingState) reset() {
	s.imports = nil
	s.moduleCtx = mainDeploymentContext{}
}

// Build the given module.
func Build(ctx context.Context, projectRootDir, stubsRoot string, config moduleconfig.AbsModuleConfig,
	sch *schema.Schema, deps, buildEnv []string, filesTransaction watch.ModifyFilesTransaction, ongoingState *OngoingState,
	devMode bool) (moduleSch optional.Option[*schema.Module], buildErrors []builderrors.Error, err error) {
	if err := filesTransaction.Begin(); err != nil {
		return moduleSch, nil, fmt.Errorf("could not start a file transaction: %w", err)
	}
	defer func() {
		if terr := filesTransaction.End(); terr != nil {
			err = fmt.Errorf("failed to end file transaction: %w", terr)
		}
		if err != nil {
			// If we failed, reset the state to ensure we don't skip steps on the next build.
			// Example: If `go mod tidy` fails due to a network failure, we need to try again next time, even if nothing else has changed.
			ongoingState.reset()
		}
	}()

	// Check dependencies
	newDeps, imports, err := extractDependenciesAndImports(config)
	if err != nil {
		return moduleSch, nil, fmt.Errorf("could not extract dependencies: %w", err)
	}
	importsChanged := ongoingState.checkIfImportsChanged(imports)
	if !slices.Equal(islices.Sort(newDeps), islices.Sort(deps)) {
		// dependencies have changed
		return moduleSch, nil, ErrInvalidateDependencies
	}

	replacements, goModVersion, err := updateGoModule(filepath.Join(config.Dir, "go.mod"), config.Module)
	if err != nil {
		return moduleSch, nil, err
	}

	goVersion := runtime.Version()[2:]
	if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
		return moduleSch, nil, fmt.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion)
	}

	logger := log.FromContext(ctx)
	funcs := maps.Clone(scaffoldFuncs)

	buildDir := buildDir(config.Dir)
	mainDir := filepath.Join(buildDir, "go", "main")

	err = os.MkdirAll(buildDir, 0750)
	if err != nil {
		return moduleSch, nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	var sharedModulesPaths []string
	for _, mod := range sch.Modules {
		if mod.Name == config.Module {
			continue
		}
		sharedModulesPaths = append(sharedModulesPaths, filepath.Join(stubsRoot, mod.Name))
	}

	if err := internal.ScaffoldZip(mainWorkTemplateFiles(), config.Dir, MainWorkContext{
		GoVersion:          goModVersion,
		SharedModulesPaths: sharedModulesPaths,
		IncludeMainPackage: mainPackageExists(config),
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return moduleSch, nil, fmt.Errorf("failed to scaffold zip: %w", err)
	}

	// In parallel, extract schema and optimistically compile.
	// These two steps take the longest, and only sometimes depend on each other.
	// After both have completed, we will scaffold out the build template and only use the optimistic compile
	// if the extracted schema has not caused any changes.
	extractResultChan := make(chan result.Result[extract.Result], 1)
	go func() {
		logger.Debugf("Extracting schema")
		extractResultChan <- result.From(extract.Extract(config.Dir))
	}()
	optimisticHashesChan := make(chan watch.FileHashes, 1)
	optimisticCompileChan := make(chan error, 1)
	go func() {
		hashes, err := fileHashesForOptimisticCompilation(config)
		if err != nil {
			optimisticHashesChan <- watch.FileHashes{}
			return
		}
		optimisticHashesChan <- hashes

		logger.Debugf("Optimistically compiling")
		optimisticCompileChan <- compile(ctx, mainDir, buildEnv, devMode)
	}()

	// wait for schema extraction to complete
	extractResult, err := (<-extractResultChan).Result()
	if err != nil {
		return moduleSch, nil, fmt.Errorf("could not extract schema: %w", err)
	}
	if builderrors.ContainsTerminalError(extractResult.Errors) {
		// Only bail if schema errors contain elements at level ERROR.
		// If errors are only at levels below ERROR (e.g. INFO, WARN), the schema can still be used.
		return moduleSch, extractResult.Errors, nil
	}

	logger.Debugf("Generating main package")
	projectName := ""
	if pcpath, ok := projectconfig.DefaultConfigPath().Get(); ok {
		pc, err := projectconfig.Load(ctx, pcpath)
		if err != nil {
			return moduleSch, nil, fmt.Errorf("failed to load project config: %w", err)
		}
		projectName = pc.Name
	}
	mctx, err := buildMainDeploymentContext(sch, extractResult, goModVersion, projectName, sharedModulesPaths, replacements)
	if err != nil {
		return moduleSch, nil, err
	}
	mainModuleCtxChanged := ongoingState.checkIfMainDeploymentContextChanged(mctx)
	if err := scaffoldBuildTemplateAndTidy(ctx, config, mainDir, importsChanged, mainModuleCtxChanged, mctx, funcs, filesTransaction); err != nil {
		return moduleSch, nil, err // nolint:wrapcheck
	}

	logger.Debugf("Writing launch script")
	if err := writeLaunchScript(buildDir); err != nil {
		return moduleSch, nil, err
	}

	// Compare main package hashes to when we optimistically compiled
	if originalHashes := (<-optimisticHashesChan); len(originalHashes) > 0 {
		currentHashes, err := fileHashesForOptimisticCompilation(config)
		if err == nil {
			changes := watch.CompareFileHashes(originalHashes, currentHashes)
			// Wait for optimistic compile to complete if there has been no changes
			if len(changes) == 0 && (<-optimisticCompileChan) == nil {
				logger.Debugf("Accepting optimistic compilation")
				return optional.Some(extractResult.Module), extractResult.Errors, nil
			}
			logger.Debugf("Discarding optimistic compilation due to file changes: %s", strings.Join(islices.Map(changes, func(change watch.FileChange) string {
				p, err := filepath.Rel(config.Dir, change.Path)
				if err != nil {
					p = change.Path
				}
				return fmt.Sprintf("%s%s", change.Change, p)
			}), ", "))
		}
	}

	logger.Debugf("Compiling")
	err = compile(ctx, mainDir, buildEnv, devMode)
	if err != nil {
		return moduleSch, nil, err
	}
	return optional.Some(extractResult.Module), extractResult.Errors, nil
}

func fileHashesForOptimisticCompilation(config moduleconfig.AbsModuleConfig) (watch.FileHashes, error) {
	// Include every file that may change while scaffolding the build template or tidying.
	hashes, err := watch.ComputeFileHashes(config.Dir, false, []string{filepath.Join(buildDirName, "go", "main", "*"), "go.mod", "go.tidy", ftlTypesFilename})
	if err != nil {
		return nil, fmt.Errorf("could not calculate hashes for optimistic compilation: %w", err)
	}
	if _, ok := hashes[filepath.Join(config.Dir, buildDirName, "go", "main", "main.go")]; !ok {
		return nil, fmt.Errorf("main package not scaffolded yet")
	}
	return hashes, nil
}

func compile(ctx context.Context, mainDir string, buildEnv []string, devMode bool) error {
	args := []string{"build", "-o", "../../main", "."}
	if devMode {
		args = []string{"build", "-gcflags=all=-N -l", "-o", "../../main", "."}
	}
	// We have seen lots of upstream HTTP/2 failures that make CI unstable.
	// Disable HTTP/2 for now during the build. This can probably be removed later
	buildEnv = slices.Clone(buildEnv)
	buildEnv = append(buildEnv, "GODEBUG=http2client=0")
	err := exec.CommandWithEnv(ctx, log.Debug, mainDir, buildEnv, "go", args...).RunStderrError(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile: %w", err)
	}
	return nil
}

func scaffoldBuildTemplateAndTidy(ctx context.Context, config moduleconfig.AbsModuleConfig, mainDir string, importsChanged,
	mainModuleCtxChanged bool, mctx mainDeploymentContext, funcs scaffolder.FuncMap, filesTransaction watch.ModifyFilesTransaction) error {
	logger := log.FromContext(ctx)
	if mainModuleCtxChanged {
		if err := internal.ScaffoldZip(buildTemplateFiles(), config.Dir, mctx, scaffolder.Exclude("^go.mod$"),
			scaffolder.Functions(funcs)); err != nil {
			return fmt.Errorf("failed to scaffold build template: %w", err)
		}
		if err := filesTransaction.ModifiedFiles(filepath.Join(config.Dir, ftlTypesFilename)); err != nil {
			return fmt.Errorf("failed to mark %s as modified: %w", ftlTypesFilename, err)
		}
	} else {
		logger.Debugf("Skipped scaffolding build template")
	}
	logger.Debugf("Tidying go.mod files")
	wg, wgctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		if !importsChanged {
			log.FromContext(ctx).Debugf("Skipped go mod tidy (module dir)")
			return nil
		}
		if err := exec.Command(wgctx, log.Debug, config.Dir, "go", "mod", "tidy").RunStderrError(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", config.Dir, err)
		}
		if err := exec.Command(wgctx, log.Debug, config.Dir, "go", "fmt", ftlTypesFilename).RunStderrError(wgctx); err != nil {
			return fmt.Errorf("%s: failed to format module dir: %w", config.Dir, err)
		}
		if err := filesTransaction.ModifiedFiles(filepath.Join(config.Dir, "go.mod"), filepath.Join(config.Dir, "go.sum"), filepath.Join(config.Dir, ftlTypesFilename)); err != nil {
			return fmt.Errorf("could not files as modified after tidying module package: %w", err)
		}
		return nil
	})
	wg.Go(func() error {
		if !mainModuleCtxChanged {
			log.FromContext(ctx).Debugf("Skipped go mod tidy (build dir)")
			return nil
		}
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "mod", "tidy").RunStderrError(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", mainDir, err)
		}
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "fmt", "./...").RunStderrError(wgctx); err != nil {
			return fmt.Errorf("%s: failed to format main dir: %w", mainDir, err)
		}
		if err := filesTransaction.ModifiedFiles(filepath.Join(mainDir, "go.mod"), filepath.Join(mainDir, "go.sum")); err != nil {
			return fmt.Errorf("could not files as modified after tidying main package: %w", err)
		}
		return nil
	})
	return wg.Wait() //nolint:wrapcheck
}

type mainDeploymentContextBuilder struct {
	sch                     *schema.Schema
	mainModule              *schema.Module
	nativeNames             extract.NativeNames
	topicMapperNames        map[*schema.Topic]finalize.TopicMapperQualifiedNames
	verbResourceParamOrders map[*schema.Verb][]common.VerbResourceParam
	imports                 map[string]string
}

func buildMainDeploymentContext(sch *schema.Schema, result extract.Result, goModVersion, projectName string,
	sharedModulesPaths []string, replacements []*modfile.Replace) (mainDeploymentContext, error) {
	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}
	combinedSch := &schema.Schema{
		Modules: append(sch.Modules, result.Module),
	}
	builder := &mainDeploymentContextBuilder{
		sch:                     combinedSch,
		mainModule:              result.Module,
		nativeNames:             result.NativeNames,
		topicMapperNames:        result.TopicPartitionMapperNames,
		verbResourceParamOrders: result.VerbResourceParamOrder,
		imports:                 imports(result.Module, false),
	}
	return builder.build(goModVersion, ftlVersion, projectName, sharedModulesPaths, replacements)
}

func (b *mainDeploymentContextBuilder) build(goModVersion, ftlVersion, projectName string,
	sharedModulesPaths []string, replacements []*modfile.Replace) (mainDeploymentContext, error) {
	ctx := &mainDeploymentContext{
		GoVersion:          goModVersion,
		FTLVersion:         ftlVersion,
		Name:               b.mainModule.Name,
		SharedModulesPaths: sharedModulesPaths,
		Replacements:       replacements,
		Verbs:              make([]goVerb, 0, len(b.mainModule.Decls)),
		Databases:          make([]goDBHandle, 0, len(b.mainModule.Decls)),
		MainCtx: mainFileContext{
			ProjectName:   projectName,
			SumTypes:      []goSumType{},
			ExternalTypes: []goExternalType{},
		},
		TypesCtx: typesFileContext{
			SumTypes:      []goSumType{},
			ExternalTypes: []goExternalType{},
		},
	}

	visited := sets.NewSet[string]()
	err := b.visit(ctx, b.mainModule, b.mainModule, visited)
	if err != nil {
		return mainDeploymentContext{}, err
	}

	slices.SortFunc(ctx.MainCtx.SumTypes, func(a, b goSumType) int {
		return strings.Compare(a.TypeName(), b.TypeName())
	})
	slices.SortFunc(ctx.TypesCtx.SumTypes, func(a, b goSumType) int {
		return strings.Compare(a.TypeName(), b.TypeName())
	})

	ctx.TypesCtx.MainModulePkg = b.mainModule.Name
	mainModuleImport := fmt.Sprintf("ftl/%s", b.mainModule.Name)
	if alias, ok := b.imports[mainModuleImport]; ok {
		mainModuleImport = fmt.Sprintf("%s %q", alias, mainModuleImport)
		ctx.TypesCtx.MainModulePkg = alias
	}
	ctx.withImports(mainModuleImport)
	return *ctx, nil
}

func writeLaunchScript(buildDir string) error {
	err := os.WriteFile(filepath.Join(buildDir, "launch"), []byte(`#!/bin/bash
	if [ -n "$FTL_DEBUG_PORT" ] && command -v dlv &> /dev/null ; then
	    dlv --listen=localhost:$FTL_DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --allow-non-terminal-interactive exec --continue ./main
	else
		exec ./main
	fi
	`), 0770) // #nosec
	if err != nil {
		return fmt.Errorf("failed to write launch script: %w", err)
	}
	return nil
}

func (b *mainDeploymentContextBuilder) visit(
	ctx *mainDeploymentContext,
	module *schema.Module,
	node schema.Node,
	visited sets.Set[string],
) error {
	err := schema.Visit(node, func(node schema.Node, next func() error) error {
		if ref, ok := node.(*schema.Ref); ok {
			maybeResolved, maybeModule := b.sch.ResolveWithModule(ref)
			resolved, ok := maybeResolved.Get()
			if !ok {
				return next()
			}
			m, ok := maybeModule.Get()
			if !ok {
				return next()
			}
			err := b.visit(ctx, m, resolved, visited)
			if err != nil {
				return fmt.Errorf("failed to visit children of %s: %w", ref, err)
			}
			return next()
		}

		maybeGoType, isLocal, err := b.getGoType(module, node)
		if err != nil {
			return err
		}
		gotype, ok := maybeGoType.Get()
		if !ok {
			return next()
		}
		if visited.Contains(gotype.getNativeType().TypeName()) {
			return next()
		}
		visited.Add(gotype.getNativeType().TypeName())

		switch n := gotype.(type) {
		case goVerb:
			ctx.Verbs = append(ctx.Verbs, n)
		case goSumType:
			if isLocal {
				ctx.TypesCtx.SumTypes = append(ctx.TypesCtx.SumTypes, n)
			}
			ctx.MainCtx.SumTypes = append(ctx.MainCtx.SumTypes, n)
		case goExternalType:
			ctx.TypesCtx.ExternalTypes = append(ctx.TypesCtx.ExternalTypes, n)
			ctx.MainCtx.ExternalTypes = append(ctx.MainCtx.ExternalTypes, n)
		case goDBHandle:
			ctx.Databases = append(ctx.Databases, n)
		}
		return next()
	})
	if err != nil {
		return fmt.Errorf("failed to build main module context: %w", err)
	}
	return nil
}

func (b *mainDeploymentContextBuilder) getGoType(module *schema.Module, node schema.Node) (gotype optional.Option[goType], isLocal bool, err error) {
	isLocal = b.visitingMainModule(module.Name)
	switch n := node.(type) {
	case *schema.Verb:
		if !isLocal {
			return optional.None[goType](), false, nil
		}
		goverb, err := b.processVerb(n)
		if err != nil {
			return optional.None[goType](), isLocal, err
		}
		return optional.Some[goType](goverb), isLocal, nil

	case *schema.Enum:
		if n.IsValueEnum() {
			return optional.None[goType](), isLocal, nil
		}
		st, err := b.processSumType(module, n)
		if err != nil {
			return optional.None[goType](), isLocal, err
		}
		return optional.Some[goType](st), isLocal, nil

	case *schema.TypeAlias:
		if len(n.Metadata) == 0 {
			return optional.None[goType](), isLocal, nil
		}
		return b.processExternalTypeAlias(n), isLocal, nil
	case *schema.Database:
		if !isLocal {
			return optional.None[goType](), false, nil
		}
		dbHandle, err := b.processDatabase(module.Name, n)
		if err != nil {
			return optional.None[goType](), isLocal, err
		}
		return optional.Some[goType](dbHandle), isLocal, nil

	default:
	}
	return optional.None[goType](), isLocal, nil
}

func (b *mainDeploymentContextBuilder) visitingMainModule(moduleName string) bool {
	return moduleName == b.mainModule.Name
}

func (b *mainDeploymentContextBuilder) processSumType(module *schema.Module, enum *schema.Enum) (goSumType, error) {
	moduleName := module.Name
	var nt nativeType
	var err error
	if !b.visitingMainModule(moduleName) {
		nt, err = nativeTypeFromQualifiedName("ftl/" + moduleName + "." + enum.Name)
	} else if nn, ok := b.nativeNames[enum]; ok {
		nt, err = b.getNativeType(nn)
	} else {
		return goSumType{}, fmt.Errorf("missing native name for enum %s", enum.Name)
	}
	if err != nil {
		return goSumType{}, err
	}

	variants := make([]goSumTypeVariant, 0, len(enum.Variants))
	for _, v := range enum.Variants {
		var vnt nativeType
		if !b.visitingMainModule(moduleName) {
			vnt, err = nativeTypeFromQualifiedName("ftl/" + moduleName + "." + v.Name)
		} else if nn, ok := b.nativeNames[v]; ok {
			vnt, err = b.getNativeType(nn)
		} else {
			return goSumType{}, fmt.Errorf("missing native name for enum variant %s", enum.Name)
		}
		if err != nil {
			return goSumType{}, err
		}

		typ, err := b.getGoSchemaType(v.Value.(*schema.TypeValue).Value)
		if err != nil {
			return goSumType{}, err
		}
		variants = append(variants, goSumTypeVariant{
			Type:       typ,
			nativeType: vnt,
		})
	}

	return goSumType{
		Variants:   variants,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processExternalTypeAlias(alias *schema.TypeAlias) optional.Option[goType] {
	for _, m := range alias.Metadata {
		if m, ok := m.(*schema.MetadataTypeMap); ok && m.Runtime == "go" {
			nt, ok := nativeTypeForWidenedType(alias)
			if !ok {
				continue
			}
			return optional.Some[goType](goExternalType{
				nativeType: nt,
			})
		}
	}
	return optional.None[goType]()
}

func (b *mainDeploymentContextBuilder) processVerb(verb *schema.Verb) (goVerb, error) {
	var resources []verbResource
	verbResourceParams, ok := b.verbResourceParamOrders[verb]
	if !ok {
		return goVerb{}, fmt.Errorf("missing verb resource param order for %s", verb.Name)
	}
	for _, m := range verbResourceParams {
		resource, err := b.getVerbResource(verb, m)
		if err != nil {
			return goVerb{}, err
		}
		resources = append(resources, resource)
	}
	nativeName, ok := b.nativeNames[verb]
	if !ok {
		return goVerb{}, fmt.Errorf("missing native name for verb %s", verb.Name)
	}
	return b.getGoVerb(nativeName, verb, resources...)
}

func (b *mainDeploymentContextBuilder) getVerbResource(verb *schema.Verb, param common.VerbResourceParam) (verbResource, error) {
	ref := param.Ref
	resolved, ok := b.sch.Resolve(ref).Get()
	if !ok {
		return nil, fmt.Errorf("failed to resolve %s resource, used by %s.%s", ref,
			b.mainModule.Name, verb.Name)
	}

	switch param.Type.(type) {
	case *schema.MetadataCalls:
		callee, ok := resolved.(*schema.Verb)
		if !ok {
			return verbClient{}, fmt.Errorf("%s.%s uses %s client, but %s is not a verb",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		calleeNativeName, ok := b.nativeNames[ref]
		if !ok {
			return verbClient{}, fmt.Errorf("missing native name for verb client %s", ref)
		}
		calleeverb, err := b.getGoVerb(calleeNativeName, callee)
		if err != nil {
			return verbClient{}, err
		}
		return verbClient{
			calleeverb,
		}, nil
	case *schema.MetadataDatabases:
		db, ok := resolved.(*schema.Database)
		if !ok {
			return goDBHandle{}, fmt.Errorf("%s.%s uses %s database handle, but %s is not a database",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return b.processDatabase(ref.Module, db)
	case *schema.MetadataPublisher:
		topic, ok := resolved.(*schema.Topic)
		if !ok {
			return goTopicHandle{}, fmt.Errorf("%s.%s uses %s topic handle, but %s is not a topic",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return b.processTopic(ref.Module, ref, topic)
	case *schema.MetadataConfig:
		cfg, ok := resolved.(*schema.Config)
		if !ok {
			return goConfigHandle{}, fmt.Errorf("%s.%s uses %s config handle, but %s is not a config",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return b.processConfig(ref.Module, ref, cfg)
	case *schema.MetadataSecrets:
		secret, ok := resolved.(*schema.Secret)
		if !ok {
			return goSecretHandle{}, fmt.Errorf("%s.%s uses %s secret handle, but %s is not a secret",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return b.processSecret(ref.Module, ref, secret)

	default:
		return nil, fmt.Errorf("unsupported resource type for verb %q", verb.Name)
	}
}

func (b *mainDeploymentContextBuilder) processConfig(moduleName string, ref *schema.Ref, config *schema.Config) (goConfigHandle, error) {
	nn, ok := b.nativeNames[ref]
	if !ok {
		return goConfigHandle{}, fmt.Errorf("missing native name for config %s.%s", moduleName, config.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goConfigHandle{}, err
	}

	ct, err := b.getGoSchemaType(config.Type)
	if err != nil {
		return goConfigHandle{}, fmt.Errorf("failed to get config type for %s.%s: %w", moduleName, config.Name, err)
	}
	return goConfigHandle{
		Name:       config.Name,
		Module:     moduleName,
		Type:       ct,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processSecret(moduleName string, ref *schema.Ref, secret *schema.Secret) (goSecretHandle, error) {
	nn, ok := b.nativeNames[ref]
	if !ok {
		return goSecretHandle{}, fmt.Errorf("missing native name for secret %s.%s", moduleName, secret.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goSecretHandle{}, err
	}

	st, err := b.getGoSchemaType(secret.Type)
	if err != nil {
		return goSecretHandle{}, fmt.Errorf("failed to get secret type for %s.%s: %w", moduleName, secret.Name, err)
	}
	return goSecretHandle{
		Name:       secret.Name,
		Module:     moduleName,
		Type:       st,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processDatabase(moduleName string, db *schema.Database) (goDBHandle, error) {
	nn, ok := b.nativeNames[db]
	if !ok {
		return goDBHandle{}, fmt.Errorf("missing native name for database %s.%s", moduleName, db.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goDBHandle{}, err
	}
	return goDBHandle{
		Name:       db.Name,
		Module:     moduleName,
		Type:       db.Type,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processTopic(moduleName string, ref *schema.Ref, topic *schema.Topic) (goTopicHandle, error) {
	nn, ok := b.nativeNames[ref]
	if !ok {
		return goTopicHandle{}, fmt.Errorf("missing native name for topic %s.%s", moduleName, topic.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goTopicHandle{}, err
	}

	et, err := b.getGoSchemaType(topic.Event)
	if err != nil {
		return goTopicHandle{}, fmt.Errorf("failed to get event type for topic %s.%s: %w", moduleName, topic.Name, err)
	}
	mapperQualifiedNames, ok := b.topicMapperNames[topic]
	if !ok {
		return goTopicHandle{}, fmt.Errorf("missing topic partition mapper name for topic %s.%s", moduleName, topic.Name)
	}
	mt, err := b.getNativeType(mapperQualifiedNames.Mapper)
	if err != nil {
		return goTopicHandle{}, fmt.Errorf("failed to get event type for topic partition mapper %s.%s: %w", moduleName, topic.Name, err)
	}
	var mapperAssociatedType optional.Option[nativeType]
	if a, ok := mapperQualifiedNames.AssociatedType.Get(); ok {
		at, err := b.getNativeType(a)
		if err != nil {
			return goTopicHandle{}, fmt.Errorf("failed to get associated type for topic partition mapper %s.%s: %w", moduleName, topic.Name, err)
		}
		mapperAssociatedType = optional.Some(at)
	}

	return goTopicHandle{
		Name:                 topic.Name,
		Module:               moduleName,
		EventType:            et,
		MapperType:           mt,
		MapperAssociatedType: mapperAssociatedType,
		nativeType:           nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) getGoVerb(nativeName string, verb *schema.Verb, resources ...verbResource) (goVerb, error) {
	nt, err := b.getNativeType(nativeName)
	if err != nil {
		return goVerb{}, err
	}
	req, err := b.getGoSchemaType(verb.Request)
	if err != nil {
		return goVerb{}, err
	}
	resp, err := b.getGoSchemaType(verb.Response)
	if err != nil {
		return goVerb{}, err
	}
	return goVerb{
		nativeType: nt,
		Request:    req,
		Response:   resp,
		Resources:  resources,
	}, nil
}

func (b *mainDeploymentContextBuilder) getGoSchemaType(typ schema.Type) (goSchemaType, error) {
	result := goSchemaType{
		TypeName:      genTypeWithNativeNames(nil, typ, b.nativeNames),
		LocalTypeName: genTypeWithNativeNames(b.mainModule, typ, b.nativeNames),
		children:      []goSchemaType{},
		nativeType:    optional.None[nativeType](),
	}

	nn, ok := b.nativeNames[typ]
	if ok {
		nt, err := b.getNativeType(nn)
		if err != nil {
			return goSchemaType{}, err
		}
		result.nativeType = optional.Some(nt)
	}

	switch t := typ.(type) {
	case *schema.Ref:
		// we add native types for all refs traversed from the main module. however, if this
		// ref was not directly traversed by the main module (e.g. the request of a verb in an external module, where
		// the main module needs a generated client for this verb), we can infer its native qualified name to get the
		// native type here.
		if !result.nativeType.Ok() {
			nt, err := b.getNativeType("ftl/" + t.Module + "." + t.Name)
			if err != nil {
				return goSchemaType{}, err
			}
			result.nativeType = optional.Some(nt)
		}
		if len(t.TypeParameters) > 0 {
			for _, tp := range t.TypeParameters {
				_r, err := b.getGoSchemaType(tp)
				if err != nil {
					return goSchemaType{}, err
				}
				result.children = append(result.children, _r)
			}
		}
	case *schema.Time:
		nt, err := b.getNativeType("time.Time")
		if err != nil {
			return goSchemaType{}, err
		}
		result.nativeType = optional.Some(nt)
	default:
	}

	return result, nil
}

func (b *mainDeploymentContextBuilder) getNativeType(qualifiedName string) (nativeType, error) {
	nt, err := nativeTypeFromQualifiedName(qualifiedName)
	if err != nil {
		return nativeType{}, err
	}
	// we already have an alias name for this import path
	if alias, ok := b.imports[nt.importPath]; ok {
		if alias != path.Base(nt.importPath) {
			nt.pkg = alias
			nt.importAlias = true
		}
		return nt, nil
	}
	b.imports = addImports(b.imports, nt)
	return nt, nil
}

var scaffoldFuncs = scaffolder.FuncMap{
	"comment": schema.EncodeComments,
	"type":    genType,
	"is": func(kind string, t schema.Node) bool {
		return stdreflect.Indirect(stdreflect.ValueOf(t)).Type().Name() == kind
	},
	"imports": func(m *schema.Module) map[string]string {
		return imports(m, true)
	},
	"value": func(v schema.Value) string {
		switch t := v.(type) {
		case *schema.StringValue:
			return fmt.Sprintf("%q", t.Value)
		case *schema.IntValue:
			return strconv.Itoa(t.Value)
		case *schema.TypeValue:
			return t.Value.String()
		}
		panic(fmt.Sprintf("unsupported value %T", v))
	},
	"enumInterfaceFunc": func(e schema.Enum) string {
		r := []rune(e.Name)
		for i, c := range r {
			if unicode.IsUpper(c) {
				r[i] = unicode.ToLower(c)
			} else {
				break
			}
		}
		return string(r)
	},
	"basicType": func(m *schema.Module, v schema.EnumVariant) bool {
		switch val := v.Value.(type) {
		case *schema.IntValue, *schema.StringValue:
			return false // This func should only return true for type enums
		case *schema.TypeValue:
			if _, ok := val.Value.(*schema.Ref); !ok {
				return true
			}
		}
		return false
	},
	"schemaType": schemaType,
	// A standalone enum variant is one that is purely an alias to a type and does not appear
	// elsewhere in the schema.
	"isStandaloneEnumVariant": func(v schema.EnumVariant) bool {
		tv, ok := v.Value.(*schema.TypeValue)
		if !ok {
			return false
		}
		if ref, ok := tv.Value.(*schema.Ref); ok {
			return ref.Name != v.Name
		}

		return false
	},
	"sumTypes": func(m *schema.Module) []*schema.Enum {
		out := []*schema.Enum{}
		for _, d := range m.Decls {
			switch d := d.(type) {
			// Type enums (i.e. sum types) are all the non-value enums
			case *schema.Enum:
				if !d.IsValueEnum() && d.IsExported() {
					out = append(out, d)
				}
			default:
			}
		}
		return out
	},
	"trimModuleQualifier": trimModuleQualifier,
	"typeAliasType": func(m *schema.Module, t *schema.TypeAlias) string {
		for _, md := range t.Metadata {
			md, ok := md.(*schema.MetadataTypeMap)
			if !ok || md.Runtime != "go" {
				continue
			}
			nt, err := nativeTypeFromQualifiedName(md.NativeName)
			if err != nil {
				return ""
			}
			return fmt.Sprintf("%s.%s", nt.pkg, nt.Name)
		}
		return genType(m, t.Type)
	},
	"getVerbClient": func(resource verbResource) *verbClient {
		if c, ok := resource.(verbClient); ok {
			return &c
		}
		return nil
	},
	"getDatabaseHandle": func(resource verbResource) *goDBHandle {
		if c, ok := resource.(goDBHandle); ok {
			return &c
		}
		return nil
	},
	"getTopicHandle": func(resource verbResource) *goTopicHandle {
		if c, ok := resource.(goTopicHandle); ok {
			return &c
		}
		return nil
	},
	"getConfigHandle": func(resource verbResource) *goConfigHandle {
		if c, ok := resource.(goConfigHandle); ok {
			return &c
		}
		return nil
	},
	"getSecretHandle": func(resource verbResource) *goSecretHandle {
		if c, ok := resource.(goSecretHandle); ok {
			return &c
		}
		return nil
	},
}

func trimModuleQualifier(moduleName string, str string) string {
	if strings.HasPrefix(str, moduleName+".") {
		return strings.TrimPrefix(str, moduleName+".")
	}
	return str
}

// returns the import path and the directory name for a type alias if there is an associated go library
func nativeTypeForWidenedType(t *schema.TypeAlias) (nt nativeType, ok bool) {
	for _, md := range t.Metadata {
		md, ok := md.(*schema.MetadataTypeMap)
		if !ok {
			continue
		}

		if md.Runtime == "go" {
			var err error
			goType, err := nativeTypeFromQualifiedName(md.NativeName)
			if err != nil {
				panic(err)
			}
			return goType, true
		}
	}
	return nt, false
}

func schemaType(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Int, *schema.Bool, *schema.String, *schema.Float, *schema.Unit, *schema.Any, *schema.Bytes, *schema.Time:
		return fmt.Sprintf("&%s{}", strings.TrimLeft(stdreflect.TypeOf(t).String(), "*"))
	case *schema.Ref:
		return fmt.Sprintf("&schema.Ref{Module: %q, Name: %q}", t.Module, t.Name)
	case *schema.Array:
		return fmt.Sprintf("&schema.Array{Element: %s}", schemaType(t.Element))
	case *schema.Map:
		return fmt.Sprintf("&schema.Map{Key: %s, Value: %s}", schemaType(t.Key), schemaType(t.Value))
	case *schema.Optional:
		return fmt.Sprintf("&schema.Optional{Type: %s}", schemaType(t.Type))
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}

func genType(module *schema.Module, t schema.Type) string {
	return genTypeWithNativeNames(module, t, nil)
}

// TODO: this is a hack because we don't currently qualify schema refs. Using native names for now to ensure
// even if the module is the same, we qualify the type with a package name when it's a subpackage.
func genTypeWithNativeNames(module *schema.Module, t schema.Type, nativeNames extract.NativeNames) string {
	switch t := t.(type) {
	case *schema.Ref:
		pkg := "ftl" + t.Module
		name := t.Name
		if nativeNames != nil {
			if nn, ok := nativeNames[t]; ok {
				nt, err := nativeTypeFromQualifiedName(nn)
				if err == nil {
					pkg = nt.pkg
					name = nt.Name
				}
			}
		}

		desc := ""
		if module != nil && pkg == "ftl"+module.Name {
			desc = name
		} else if t.Module == "" {
			desc = name
		} else {
			desc = pkg + "." + name
		}
		if len(t.TypeParameters) > 0 {
			desc += "["
			for i, tp := range t.TypeParameters {
				if i != 0 {
					desc += ", "
				}
				desc += genTypeWithNativeNames(module, tp, nativeNames)
			}
			desc += "]"
		}
		return desc

	case *schema.Float:
		return "float64"

	case *schema.Time:
		return "stdtime.Time"

	case *schema.Int, *schema.Bool, *schema.String:
		return strings.ToLower(t.String())

	case *schema.Array:
		return "[]" + genTypeWithNativeNames(module, t.Element, nativeNames)

	case *schema.Map:
		return "map[" + genTypeWithNativeNames(module, t.Key, nativeNames) + "]" + genType(module, t.Value)

	case *schema.Optional:
		return "ftl.Option[" + genTypeWithNativeNames(module, t.Type, nativeNames) + "]"

	case *schema.Unit:
		return "ftl.Unit"

	case *schema.Any:
		return "any"

	case *schema.Bytes:
		return "[]byte"
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}

// Update go.mod file to include the FTL version and return the Go version and any replace directives.
func updateGoModule(goModPath string, expectedModule string) (replacements []*modfile.Replace, goVersion string, err error) {
	goModFile, replacements, err := goModFileWithReplacements(goModPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update %s: %w", goModPath, err)
	}

	// Validate module name matches
	if !strings.HasPrefix(goModFile.Module.Mod.Path, "ftl/"+expectedModule) {
		return nil, "", fmt.Errorf("module name mismatch: expected 'ftl/%s' but got '%s' in go.mod",
			expectedModule, goModFile.Module.Mod.Path)
	}

	// Early return if we're not updating anything.
	if !ftl.IsRelease(ftl.Version) || !shouldUpdateVersion(goModFile) {
		return replacements, goModFile.Go.Version, nil
	}

	if err := goModFile.AddRequire("github.com/block/ftl", "v"+ftl.Version); err != nil {
		return nil, "", fmt.Errorf("failed to add github.com/block/ftl to %s: %w", goModPath, err)
	}

	// Atomically write the updated go.mod file.
	tmpFile, err := os.CreateTemp(filepath.Dir(goModPath), ".go.mod-")
	if err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	defer os.Remove(tmpFile.Name()) // Delete the temp file if we error.
	defer tmpFile.Close()
	goModBytes := modfile.Format(goModFile.Syntax)
	if _, err := tmpFile.Write(goModBytes); err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	if err := os.Rename(tmpFile.Name(), goModPath); err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	return replacements, goModFile.Go.Version, nil
}

func goModFileWithReplacements(goModPath string) (*modfile.File, []*modfile.Replace, error) {
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", goModPath, err)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s: %w", goModPath, err)
	}

	replacements := reflect.DeepCopy(goModFile.Replace)
	for i, r := range replacements {
		if strings.HasPrefix(r.New.Path, ".") {
			abs, err := filepath.Abs(filepath.Join(filepath.Dir(goModPath), r.New.Path))
			if err != nil {
				return nil, nil, err
			}
			replacements[i].New.Path = abs
		}
	}
	return goModFile, replacements, nil
}

func shouldUpdateVersion(goModfile *modfile.File) bool {
	for _, require := range goModfile.Require {
		if require.Mod.Path == "github.com/block/ftl" && require.Mod.Version == ftl.Version {
			return false
		}
	}
	return true
}

// returns the import path and directory name for a Go type
// package and directory names are the same (dir=bar, pkg=bar): "github.com/foo/bar.A" => "github.com/foo/bar", none
// package and directory names differ (dir=bar, pkg=baz): "github.com/foo/bar.baz.A" => "github.com/foo/bar", "baz"
func nativeTypeFromQualifiedName(qualifiedName string) (nativeType, error) {
	lastDotIndex := strings.LastIndex(qualifiedName, ".")
	if lastDotIndex == -1 {
		return nativeType{}, fmt.Errorf("invalid qualified type format %q", qualifiedName)
	}

	pkgPath := qualifiedName[:lastDotIndex]
	typeName := qualifiedName[lastDotIndex+1:]
	pkgName := path.Base(pkgPath)
	aliased := false

	if strings.LastIndex(pkgName, ".") != -1 {
		lastDotIndex = strings.LastIndex(pkgPath, ".")
		pkgName = pkgPath[lastDotIndex+1:]
		pkgPath = pkgPath[:lastDotIndex]
		aliased = true
	}

	if parts := strings.Split(qualifiedName, "/"); len(parts) > 0 && parts[0] == "ftl" {
		aliased = true
		pkgName = "ftl" + pkgName
	}

	return nativeType{
		Name:        typeName,
		pkg:         pkgName,
		importPath:  pkgPath,
		importAlias: aliased,
	}, nil
}

// imports returns a map of import paths to aliases for a module.
// - hardcoded for time ("stdtime")
// - prefixed with "ftl" for other modules (eg "ftlfoo")
// - addImports() is used to generate shortest unique aliases for external packages
func imports(m *schema.Module, aliasesMustBeExported bool) map[string]string {
	// find all imports
	imports := map[string]string{}
	// map from import path to the first dir we see
	extraImports := map[string]nativeType{}
	_ = schema.VisitExcludingMetadataChildren(m, func(n schema.Node, next func() error) error { //nolint:errcheck
		switch n := n.(type) {
		case *schema.Ref:
			if n.Module == "" || n.Module == m.Name {
				break
			}
			imports[path.Join("ftl", n.Module)] = "ftl" + n.Module
			for _, tp := range n.TypeParameters {
				if tpRef, ok := tp.(*schema.Ref); ok && tpRef.Module != "" && tpRef.Module != m.Name {
					imports[path.Join("ftl", tpRef.Module)] = "ftl" + tpRef.Module
				}
			}

		case *schema.Time:
			imports["time"] = "stdtime"

		case *schema.Optional, *schema.Unit:
			imports["github.com/block/ftl/go-runtime/ftl"] = "ftl"

		case *schema.Topic:
			if n.IsExported() {
				imports["github.com/block/ftl/go-runtime/ftl"] = "ftl"
			}

		case *schema.TypeAlias:
			if aliasesMustBeExported && !n.IsExported() {
				return next()
			}
			if nt, ok := nativeTypeForWidenedType(n); ok {
				if existing, ok := extraImports[nt.importPath]; !ok || !existing.importAlias {
					extraImports[nt.importPath] = nt
				}
			}
		default:
		}
		return next()
	})

	return addImports(imports, maps.Values(extraImports)...)
}

// addImports takes existing imports (mapping import path to pkg alias) and adds new imports by generating aliases
// aliases are generated for external types by finding the shortest unique alias that can be used without conflict:
func addImports(existingImports map[string]string, newTypes ...nativeType) map[string]string {
	imports := maps.Clone(existingImports)
	// maps import path to possible aliases, shortest to longest
	aliasesForImports := map[string][]string{}

	// maps possible aliases with the count of imports that could use the alias
	possibleImportAliases := map[string]int{}
	for _, alias := range imports {
		possibleImportAliases[alias]++
	}
	for _, nt := range newTypes {
		if _, ok := imports[nt.importPath]; ok {
			continue
		}

		importPath := nt.importPath
		pathComponents := strings.Split(importPath, "/")
		if nt.importAlias {
			pathComponents = append(pathComponents, nt.pkg)
		}

		var currentAlias string
		for i := range len(pathComponents) {
			runes := []rune(pathComponents[len(pathComponents)-1-i])
			for i, char := range runes {
				if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
					runes[i] = '_'
				}
			}
			if unicode.IsNumber(runes[0]) {
				newRunes := make([]rune, len(runes)+1)
				newRunes[0] = '_'
				copy(newRunes[1:], runes)
				runes = newRunes
			}
			foldedComponent := string(runes)
			if i == 0 {
				currentAlias = foldedComponent
			} else {
				currentAlias = foldedComponent + "_" + currentAlias
			}
			aliasesForImports[importPath] = append(aliasesForImports[importPath], currentAlias)
			possibleImportAliases[currentAlias]++
		}
	}
	for importPath, aliases := range aliasesForImports {
		found := false
		for _, alias := range aliases {
			if possibleImportAliases[alias] == 1 {
				imports[importPath] = alias
				found = true
				break
			}
		}
		if !found {
			// no possible alias that is unique, use the last one as no other type will choose the same
			imports[importPath] = aliases[len(aliases)-1]
		}
	}
	return imports
}
