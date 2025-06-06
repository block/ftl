package compile

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	stdreflect "reflect"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/alecthomas/errors"
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
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/common/strcase"
	extract "github.com/block/ftl/go-runtime/schema"
	"github.com/block/ftl/go-runtime/schema/common"
	"github.com/block/ftl/go-runtime/schema/finalize"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

var (
	ftlTypesFilename = "types.ftl.go"
	// Searches for paths beginning with ../../../
	// Matches the previous character as well to avoid matching the middle deeper backtrack paths, and because
	// regex package does not support negative lookbehind assertions
	pathFromMainToModuleRegex = regexp.MustCompile(`(^|[^/])\.\./\.\./\.\./`)
)

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
	QueriesCtx         queriesFileContext
}

func (c *mainDeploymentContext) withImports(mainModuleImport string) {
	c.MainCtx.Imports = c.generateMainImports()
	c.TypesCtx.Imports = c.generateTypesImports(mainModuleImport)
	c.QueriesCtx.Imports = c.generateQueryImports()
}

func (c *mainDeploymentContext) generateMainImports() []string {
	imports := sets.NewSet[string]()
	imports.Add(`"context"`)
	imports.Add(`"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"`)
	imports.Add(`"github.com/block/ftl/common/plugin"`)
	imports.Add(`"github.com/block/ftl/go-runtime/server"`)

	for _, v := range c.Verbs {
		imports.Append(verbImports(v, true)...)
	}
	out := imports.ToSlice()
	slices.Sort(out)
	return out
}

func (c *mainDeploymentContext) generateQueryImports() []string {
	imports := sets.NewSet[string]()
	imports.Add(`"context"`)
	imports.Add(`"github.com/block/ftl/go-runtime/server"`)
	imports.Add(`"github.com/block/ftl/common/reflection"`)
	imports.Add(`"github.com/alecthomas/types/tuple"`)
	for _, d := range c.QueriesCtx.Data {
		for _, f := range d.Fields {
			imports.Append(schemaTypeImports(f.Type, true)...)
		}
	}
	for _, v := range c.QueriesCtx.Verbs {
		imports.Append(schemaTypeImports(v.Request, false)...)
		imports.Append(schemaTypeImports(v.Response, false)...)
	}
	result := imports.ToSlice()
	slices.Sort(result)
	return result
}

func (c *mainDeploymentContext) generateTypesImports(mainModuleImport string) []string {
	imports := sets.NewSet[string]()
	if len(c.TypesCtx.SumTypes) > 0 || len(c.TypesCtx.ExternalTypes) > 0 {
		imports.Add(`"github.com/block/ftl/common/reflection"`)
	}
	hasVerbs := false
	for _, i := range c.Verbs {
		if !i.IsQuery {
			hasVerbs = true
		}
	}
	if hasVerbs {
		imports.Add(`"context"`)
	}
	if len(c.Databases) > 0 {
		imports.Add(`"github.com/block/ftl/go-runtime/server"`)
		imports.Add(`"github.com/block/ftl/go-runtime/ftl"`)
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
		if v.IsQuery {
			continue
		}

		imports.Add(`"github.com/block/ftl/common/reflection"`)
		if len(v.Resources) > 0 {
			imports.Add(`"github.com/block/ftl/go-runtime/server"`)
		}
		imports.Append(verbImports(v, false)...)
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

func typeImports(t goSchemaType, importUnit bool) []string {
	imports := sets.NewSet[string]()
	if nt, ok := t.nativeType.Get(); ok {
		imports.Add(nt.importStatement())
	}
	imports.Append(schemaTypeImports(t.schemaType, importUnit)...)
	for _, c := range t.children {
		imports.Append(typeImports(c, importUnit)...)
	}
	return imports.ToSlice()
}

func schemaTypeImports(t schema.Type, importUnit bool) []string {
	imports := sets.NewSet[string]()
	if usesType(t, &schema.Optional{}) {
		imports.Add(`"github.com/block/ftl/go-runtime/ftl"`)
	}
	if usesType(t, &schema.Time{}) {
		imports.Add(`stdtime "time"`)
	}
	if usesType(t, &schema.Unit{}) && importUnit {
		imports.Add(`"github.com/block/ftl/go-runtime/ftl"`)
	}
	return imports.ToSlice()
}

func verbImports(v goVerb, main bool) []string {
	imports := sets.NewSet[string]()

	if main {
		imports.Add("_ " + strconv.Quote(v.importPath))
	} else {
		imports.Add(v.importStatement())
		imports.Add(`"github.com/block/ftl/common/reflection"`)
	}

	imports.Append(typeImports(v.Request, false)...)
	imports.Append(typeImports(v.Response, false)...)

	if nt, ok := v.Request.nativeType.Get(); ok && v.Request.TypeName != "ftl.Unit" {
		imports.Add(nt.importStatement())
	}
	if nt, ok := v.Response.nativeType.Get(); ok && v.Response.TypeName != "ftl.Unit" {
		imports.Add(nt.importStatement())
	}
	for _, r := range v.Request.children {
		imports.Append(typeImports(r, true)...)
	}
	for _, r := range v.Response.children {
		imports.Append(typeImports(r, true)...)
	}

	if !main {
		for _, r := range v.Resources {
			switch r := r.(type) {
			case verbClient:
				imports.Add(`"github.com/block/ftl/go-runtime/server"`)
				imports.Append(verbImports(r.goVerb, false)...)
			case goTopicHandle:
				imports.Add(r.MapperType.importStatement())
			}
		}
	}
	return imports.ToSlice()
}

type mainFileContext struct {
	Imports []string

	ProjectName string
}

type typesFileContext struct {
	Imports       []string
	MainModulePkg string

	SumTypes      []goSumType
	ExternalTypes []goExternalType
}

type queriesFileContext struct {
	Module  *schema.Module
	Verbs   []queryVerb
	Data    []*schema.Data
	Imports []string
}

type queryVerb struct {
	*schema.Verb
	CommandType    string
	RawSQL         string
	ParamFields    string
	ColToFieldName string
	DBName         string
	DBType         string
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
	Name                  string
	Request               goSchemaType
	Response              goSchemaType
	Resources             []verbResource
	IsQuery               bool
	TransactionDatasource optional.Option[string]

	nativeType
}

func (g goVerb) IsTransaction() bool {
	return g.TransactionDatasource.Ok()
}

func (g goVerb) TransactionDatasourceName() string {
	return g.TransactionDatasource.MustGet()
}

type goSchemaType struct {
	TypeName      string
	LocalTypeName string
	children      []goSchemaType
	schemaType    schema.Type

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

type verbEgress struct {
	Target string
}

func (v verbEgress) resource() {}

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
		filepath.Join(config.SQLRootDir, moduleconfig.DBFilename),
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

func buildErrorFromError(err error) builderrors.Error {
	var buildErr builderrors.Error
	if errors.As(err, &buildErr) {
		return buildErr
	}
	return builderrors.Error{
		Type:  builderrors.FTL,
		Msg:   err.Error(),
		Pos:   optional.None[builderrors.Position](),
		Level: builderrors.ERROR,
	}
}

// IsEmpty returns true if the OngoingState is in its initial/reset state
func (s *OngoingState) IsEmpty() bool {
	return s == nil || (s.imports == nil && stdreflect.DeepEqual(s.moduleCtx, mainDeploymentContext{}))
}

// Build the given module.
func Build(ctx context.Context, projectConfig projectconfig.Config, stubsRoot string, config moduleconfig.AbsModuleConfig,
	sch *schema.Schema, deps, buildEnv []string, filesTransaction watch.ModifyFilesTransaction, ongoingState *OngoingState,
	devMode bool) (moduleSch optional.Option[*schema.Module], invalidateDeps bool, buildErrors []builderrors.Error) {
	logger := log.FromContext(ctx)

	// Only clean the build directory if the ongoing state is empty (first build or reset)
	if ongoingState.IsEmpty() {
		buildDir := buildDir(config.Dir)
		if err := os.RemoveAll(buildDir); err != nil {
			return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "failed to clean build directory"))}
		}
	}

	if err := filesTransaction.Begin(); err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "could not start a file transaction"))}
	}
	defer func() {
		if terr := filesTransaction.End(); terr != nil {
			buildErrors = append(buildErrors, buildErrorFromError(errors.Wrap(terr, "failed to end file transaction")))
		}
		if _, hasErrs := islices.Find(buildErrors, func(berr builderrors.Error) bool { //nolint:errcheck
			return berr.Level == builderrors.ERROR
		}); hasErrs {
			// If we failed, reset the state to ensure we don't skip steps on the next build.
			// Example: If `go mod tidy` fails due to a network failure, we need to try again next time, even if nothing else has changed.
			ongoingState.reset()
		}
	}()

	// Check dependencies
	newDeps, imports, err := extractDependenciesAndImports(config)
	if err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "could not extract dependencies"))}
	}
	importsChanged := ongoingState.checkIfImportsChanged(imports)
	if !slices.Equal(islices.Sort(newDeps), islices.Sort(deps)) {
		// dependencies have changed
		return moduleSch, true, nil
	}

	replacements, goModVersion, err := updateGoModule(filepath.Join(config.Dir, "go.mod"), config.Module, optional.Some(filesTransaction))
	if err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(err)}
	}

	goVersion := runtime.Version()[2:]
	if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion))}
	}

	funcs := maps.Clone(scaffoldFuncs)

	mainDir := filepath.Join(buildDir(config.Dir), "go", "main")

	err = os.MkdirAll(buildDir(config.Dir), 0750)
	if err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "failed to create build directory"))}
	}

	var sharedModulesPaths []string
	for _, mod := range sch.InternalModules() {
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
		return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "failed to scaffold zip"))}
	}

	// In parallel, extract schema and optimistically compile.
	// These two steps take the longest, and only sometimes depend on each other.
	// After both have completed, we will scaffold out the build template and only use the optimistic compile
	// if the extracted schema has not caused any changes.
	extractResultChan := make(chan result.Result[extract.Result], 1)
	go func() {
		logger.Debugf("Extracting schema")
		extractResultChan <- result.From(extract.Extract(projectConfig.Root(), config.Dir, sch))
	}()
	optimisticHashesChan := make(chan watch.FileHashes, 1)
	optimisticCompileChan := make(chan []builderrors.Error, 1)
	go func() {
		hashes, err := fileHashesForOptimisticCompilation(config)
		if err != nil {
			optimisticHashesChan <- watch.FileHashes{}
			optimisticCompileChan <- nil
			return
		}
		optimisticHashesChan <- hashes

		logger.Debugf("Optimistically compiling")
		optimisticCompileChan <- compile(ctx, mainDir, buildEnv, devMode)
	}()

	// wait for schema extraction to complete
	extractResult, err := (<-extractResultChan).Result()
	if err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "could not extract schema"))}
	}
	// We do not fail yet if result has terminal errors. These errors may be due to missing templated files (queries.ftl.go).
	// Instead we scaffold if needed and then re-extract the schema to see if the errors are resolved.
	logger.Debugf("Generating main package")
	projectName := projectConfig.Name
	mctx, err := buildMainDeploymentContext(sch, extractResult, goModVersion, projectName, sharedModulesPaths, replacements)
	if err != nil {
		// Combine with compiler errors as they are likely the cause of why we would not build mctx.
		buildErrs := <-optimisticCompileChan
		buildErrs = append(buildErrs, buildErrorFromError(err))
		return moduleSch, false, buildErrs
	}
	mainModuleCtxChanged := ongoingState.checkIfMainDeploymentContextChanged(mctx)
	if err := scaffoldBuildTemplateAndTidy(ctx, config, mainDir, importsChanged, mainModuleCtxChanged, mctx, funcs, filesTransaction); err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(err)}
	}

	if mainModuleCtxChanged && builderrors.ContainsTerminalError(extractResult.Errors) {
		// We may have terminal errors that are resolved by scaffolding queries.
		logger.Debugf("Re-extracting schema after scaffolding")
		extractResult, err = extract.Extract(projectConfig.Root(), config.Dir, sch)
		if err != nil {
			return moduleSch, false, []builderrors.Error{buildErrorFromError(errors.Wrap(err, "could not extract schema"))}
		}
	}
	if builderrors.ContainsTerminalError(extractResult.Errors) {
		// Only bail if schema errors contain elements at level ERROR.
		// If errors are only at levels below ERROR (e.g. INFO, WARN), the schema can still be used.
		return moduleSch, false, extractResult.Errors
	}

	logger.Debugf("Writing launch script")
	if err := writeLaunchScript(buildDir(config.Dir)); err != nil {
		return moduleSch, false, []builderrors.Error{buildErrorFromError(err)}
	}

	// Compare main package hashes to when we optimistically compiled
	if originalHashes := (<-optimisticHashesChan); len(originalHashes) > 0 {
		currentHashes, err := fileHashesForOptimisticCompilation(config)
		if err == nil {
			changes := watch.CompareFileHashes(originalHashes, currentHashes)
			// Wait for optimistic compile to complete if there has been no changes
			if len(changes) == 0 && (<-optimisticCompileChan) == nil {
				logger.Debugf("Accepting optimistic compilation")
				return optional.Some(extractResult.Module), false, extractResult.Errors
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
	buildErrors = compile(ctx, mainDir, buildEnv, devMode)
	buildErrors = append(buildErrors, extractResult.Errors...)
	return optional.Some(extractResult.Module), false, buildErrors
}

func fileHashesForOptimisticCompilation(config moduleconfig.AbsModuleConfig) (watch.FileHashes, error) {
	args := []string{filepath.Join(buildDirName, "go", "main", "*"), "go.mod", "go.tidy", ftlTypesFilename}
	// Include every file that may change while scaffolding the build template or tidying.
	if config.SQLRootDir == "" {
		relativeSQLRootDir, err := filepath.Rel(config.Dir, config.SQLRootDir)
		if err != nil {
			return nil, errors.Wrap(err, "could not calculate relative SQL root dir")
		}
		args = append(args, filepath.Join(relativeSQLRootDir, moduleconfig.DBFilename))
	}
	hashes, err := watch.ComputeFileHashes(config.Dir, false, args)
	if err != nil {
		return nil, errors.Wrap(err, "could not calculate hashes for optimistic compilation")
	}
	if _, ok := hashes[filepath.Join(config.Dir, buildDirName, "go", "main", "main.go")]; !ok {
		return nil, errors.Errorf("main package not scaffolded yet")
	}
	return hashes, nil
}

func compile(ctx context.Context, mainDir string, buildEnv []string, devMode bool) []builderrors.Error {
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
		return buildErrsFromCompilerErr(err.Error())
	}
	return nil
}

// buildErrsFromCompilerErr converts a compiler error to a list of builderrors by trying to parse
// the error text and extract multiple errors with file, line and column information
func buildErrsFromCompilerErr(input string) []builderrors.Error {
	errs := []builderrors.Error{}
	lines := strings.Split(input, "\n")
	for i := 0; i < len(lines); {
		if i > 0 && strings.HasPrefix(lines[i], "\t") {
			lines[i-1] = lines[i-1] + ": " + strings.TrimSpace(lines[i])
			lines = append(lines[:i], lines[i+1:]...)
		} else {
			lines[i] = strings.TrimSpace(lines[i])
			if len(lines[i]) == 0 {
				lines = append(lines[:i], lines[i+1:]...)
				continue
			}
			i++
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			continue
		}
		// ../../../example.go:11:59: undefined: lorem ipsum
		line = pathFromMainToModuleRegex.ReplaceAllString(line, "$1")
		components := islices.Map(strings.SplitN(line, ":", 4), strings.TrimSpace)
		if len(components) != 4 {
			errs = append(errs, builderrors.Error{
				Type:  builderrors.COMPILER,
				Msg:   line,
				Level: builderrors.ERROR,
			})
			continue
		}

		path := components[0]
		hasPos := true
		line, err := strconv.Atoi(components[1])
		if err != nil {
			hasPos = false
		}
		startCol, err := strconv.Atoi(components[2])
		if err != nil {
			hasPos = false
		}
		pos := optional.None[builderrors.Position]()
		if hasPos {
			pos = optional.Some(builderrors.Position{
				Filename:    path,
				Line:        line,
				StartColumn: startCol,
				EndColumn:   startCol,
			})
		}
		errs = append(errs, builderrors.Error{
			Type:  builderrors.COMPILER,
			Msg:   components[3],
			Pos:   pos,
			Level: builderrors.ERROR,
		})
	}
	return errs
}

func scaffoldBuildTemplateAndTidy(ctx context.Context, config moduleconfig.AbsModuleConfig, mainDir string, importsChanged,
	mainModuleCtxChanged bool, mctx mainDeploymentContext, funcs scaffolder.FuncMap, filesTransaction watch.ModifyFilesTransaction) error {
	logger := log.FromContext(ctx)
	if mainModuleCtxChanged {
		if err := internal.ScaffoldZip(buildTemplateFiles(), config.Dir, mctx, scaffolder.Exclude("^go.mod$"),
			scaffolder.Functions(funcs)); err != nil {
			return errors.Wrap(err, "failed to scaffold build template")
		}
		if len(mctx.QueriesCtx.Verbs) > 0 {
			if err := internal.ScaffoldZip(queriesTemplateFiles(), config.SQLRootDir, mctx, scaffolder.Exclude("^go.mod$"),
				scaffolder.Functions(funcs)); err != nil {
				return errors.Wrap(err, "failed to scaffold queries template")
			}
			if err := filesTransaction.ModifiedFiles(filepath.Join(config.SQLRootDir, moduleconfig.DBFilename)); err != nil {
				return errors.Wrapf(err, "failed to mark %s as modified", moduleconfig.DBFilename)
			}
		} else if _, err := os.Stat(filepath.Join(config.SQLRootDir, moduleconfig.DBFilename)); err == nil && len(mctx.QueriesCtx.Verbs) == 0 && len(mctx.QueriesCtx.Data) == 0 {
			if err := os.Remove(filepath.Join(config.SQLRootDir, moduleconfig.DBFilename)); err != nil {
				return errors.Wrapf(err, "failed to delete %s", moduleconfig.DBFilename)
			}
			if err := filesTransaction.ModifiedFiles(filepath.Join(config.SQLRootDir, moduleconfig.DBFilename)); err != nil {
				return errors.Wrapf(err, "failed to mark %s as deleted", moduleconfig.DBFilename)
			}
		}
		if err := filesTransaction.ModifiedFiles(filepath.Join(config.Dir, ftlTypesFilename)); err != nil {
			return errors.Wrapf(err, "failed to mark %s as modified", ftlTypesFilename)
		}
	} else {
		logger.Debugf("Skipped scaffolding build template")
	}
	logger.Debugf("Tidying go.mod files")
	wg, wgctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		updatedFiles := []string{}
		if !importsChanged {
			log.FromContext(ctx).Debugf("Skipped go mod tidy (module dir)")
			// Even if imports didn't change, we might need to format if scaffolding happened.
		} else {
			if err := exec.Command(wgctx, log.Debug, config.Dir, "go", "mod", "tidy").RunStderrError(wgctx); err != nil {
				return errors.Wrapf(err, "%s: failed to tidy go.mod", config.Dir)
			}
			// Mark files as modified if tidy or format might have run
			// This is slightly broader but ensures transaction catches changes.
			updatedFiles = append(updatedFiles, filepath.Join(config.Dir, "go.mod"), filepath.Join(config.Dir, "go.sum"))
		}

		// Always run go fmt on potentially scaffolded files if they exist.
		typesFilePath := filepath.Join(config.Dir, ftlTypesFilename)
		queriesFilePath := filepath.Join(config.SQLRootDir, moduleconfig.DBFilename)

		if _, err := os.Stat(typesFilePath); err == nil {
			log.FromContext(ctx).Debugf("Formatting scaffolded files: %v", typesFilePath)
			if err := exec.Command(wgctx, log.Debug, config.Dir, "go", "fmt", ftlTypesFilename).RunStderrError(wgctx); err != nil {
				return errors.Wrapf(err, "%s: failed to format module dir files: %v", config.Dir, ftlTypesFilename)
			}
			updatedFiles = append(updatedFiles, typesFilePath)
		}
		if _, err := os.Stat(queriesFilePath); err == nil {
			log.FromContext(ctx).Debugf("Formatting scaffolded files: %v", queriesFilePath)
			if err := exec.Command(wgctx, log.Debug, config.Dir, "go", "fmt", queriesFilePath).RunStderrError(wgctx); err != nil {
				return errors.Wrapf(err, "%s: failed to format module dir files: %v", config.Dir, queriesFilePath)
			}
			updatedFiles = append(updatedFiles, queriesFilePath)
		}

		if len(updatedFiles) > 0 {
			if err := filesTransaction.ModifiedFiles(updatedFiles...); err != nil {
				return errors.Wrap(err, "could not mark files as modified after tidying/formatting module package")
			}
		}

		return nil
	})
	wg.Go(func() error {
		if !mainModuleCtxChanged {
			log.FromContext(ctx).Debugf("Skipped go mod tidy (build dir)")
			return nil
		}
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "mod", "tidy").RunStderrError(wgctx); err != nil {
			return errors.Wrapf(err, "%s: failed to tidy go.mod", mainDir)
		}
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "fmt", "./...").RunStderrError(wgctx); err != nil {
			return errors.Wrapf(err, "%s: failed to format main dir", mainDir)
		}
		if err := filesTransaction.ModifiedFiles(filepath.Join(mainDir, "go.mod"), filepath.Join(mainDir, "go.sum")); err != nil {
			return errors.Wrap(err, "could not files as modified after tidying main package")
		}
		return nil
	})
	return errors.WithStack(wg.Wait()) //nolint:wrapcheck
}

type mainDeploymentContextBuilder struct {
	sch                     *schema.Schema
	mainModule              *schema.Module
	nativeNames             extract.NativeNames
	topicMapperNames        map[*schema.Topic]finalize.TopicMapperQualifiedNames
	verbResourceParamOrders map[*schema.Verb][]common.VerbResourceParam
	imports                 map[string]string
	visited                 sets.Set[string]
}

func buildMainDeploymentContext(
	sch *schema.Schema,
	result extract.Result,
	goModVersion,
	projectName string,
	sharedModulesPaths []string,
	replacements []*modfile.Replace,
) (mainDeploymentContext, error) {
	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}
	realm := &schema.Realm{
		Name:    projectName,
		Modules: append(sch.InternalModules(), result.Module),
	}
	combinedSch := &schema.Schema{Realms: []*schema.Realm{realm}}
	builder := &mainDeploymentContextBuilder{
		sch:                     combinedSch,
		mainModule:              result.Module,
		nativeNames:             result.NativeNames,
		topicMapperNames:        result.TopicPartitionMapperNames,
		verbResourceParamOrders: result.VerbResourceParamOrder,
		imports:                 imports(result.Module, false),
		visited:                 sets.NewSet[string](),
	}
	return errors.WithStack2(builder.build(goModVersion, ftlVersion, projectName, sharedModulesPaths, replacements))
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
			ProjectName: projectName,
		},
		TypesCtx: typesFileContext{
			SumTypes:      []goSumType{},
			ExternalTypes: []goExternalType{},
		},
		QueriesCtx: queriesFileContext{
			Module: b.mainModule,
			Verbs:  []queryVerb{},
			Data:   []*schema.Data{},
		},
	}

	err := b.visit(ctx, b.mainModule, b.mainModule, []schema.Node{})
	if err != nil {
		return mainDeploymentContext{}, errors.WithStack(err)
	}

	slices.SortFunc(ctx.TypesCtx.SumTypes, func(a, b goSumType) int {
		return strings.Compare(a.TypeName(), b.TypeName())
	})

	ctx.TypesCtx.MainModulePkg = b.mainModule.Name
	mainModuleImport := fmt.Sprintf("ftl/%s", b.mainModule.Name)
	if alias, ok := b.imports[mainModuleImport]; ok {
		mainModuleImport = fmt.Sprintf("%s %q", alias, mainModuleImport)
		ctx.TypesCtx.MainModulePkg = alias
	}

	slices.SortFunc(ctx.QueriesCtx.Verbs, func(a, b queryVerb) int {
		return strings.Compare(a.Verb.Name, b.Verb.Name)
	})
	slices.SortFunc(ctx.QueriesCtx.Data, func(a, b *schema.Data) int {
		return strings.Compare(a.Name, b.Name)
	})

	ctx.withImports(mainModuleImport)
	return *ctx, nil
}

func writeLaunchScript(buildDir string) error {
	err := os.WriteFile(filepath.Join(buildDir, "launch"), []byte(`#!/bin/bash
	if [ -n "$FTL_DEBUG_PORT" ] && command -v dlv &> /dev/null ; then
	    dlv --listen=localhost:$FTL_DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --allow-non-terminal-interactive exec --continue ./main
	else
	 	./main
	fi
	`), 0770) // #nosec
	if err != nil {
		return errors.Wrap(err, "failed to write launch script")
	}
	return nil
}

func (b *mainDeploymentContextBuilder) visit(
	ctx *mainDeploymentContext,
	module *schema.Module,
	node schema.Node,
	parents []schema.Node,
) error {
	err := schema.VisitWithParents(node, parents, func(node schema.Node, parents []schema.Node, next func() error) error {
		switch n := node.(type) {
		case *schema.Verb:
			if _, isQuery := n.GetQuery(); isQuery {
				refName := module.Name + "." + n.Name
				if b.visited.Contains(refName) {
					return errors.WithStack(next())
				}
				b.visited.Add(refName)
				isLocal := b.visitingMainModule(module.Name)
				if isLocal {
					verbs, data, err := b.getQueryDecls(n)
					if err != nil {
						return errors.WithStack(err)
					}
					ctx.QueriesCtx.Verbs = append(ctx.QueriesCtx.Verbs, verbs...)
					ctx.QueriesCtx.Data = append(ctx.QueriesCtx.Data, data...)
					verb, err := b.getGoVerb(fmt.Sprintf("ftl/%s/db.%s", b.mainModule.Name, n.Name), n)
					if err != nil {
						return errors.WithStack(err)
					}
					ctx.Verbs = append(ctx.Verbs, verb)
				}
			}
		case *schema.Ref:
			maybeResolved, maybeModule := b.sch.ResolveWithModule(n)
			resolved, ok := maybeResolved.Get()
			if !ok {
				return errors.WithStack(next())
			}
			m, ok := maybeModule.Get()
			if !ok {
				return errors.WithStack(next())
			}
			// Check for infinite loop
			for i, p := range parents {
				if p == resolved {
					named := slices.Collect(islices.FilterVariants[schema.Named](parents[i : len(parents)-1]))
					return errors.Errorf("cyclic references are not allowed: %s%s", strings.Join(islices.Map(named,
						func(n schema.Named) string {
							return module.Name + "." + n.GetName() + " refers to "
						}), ""), n.String())
				}
			}
			err := b.visit(ctx, m, resolved, parents)
			if err != nil {
				return errors.WithStack(err) //nolint:wrapcheck
			}
			return errors.WithStack(next())
		default:
		}

		maybeGoType, _, err := b.getGoType(module, node)
		if err != nil {
			return errors.WithStack(err)
		}
		gotype, ok := maybeGoType.Get()
		if !ok {
			return errors.WithStack(next())
		}
		if b.visited.Contains(gotype.getNativeType().TypeName()) {
			return errors.WithStack(next())
		}
		b.visited.Add(gotype.getNativeType().TypeName())

		switch n := gotype.(type) {
		case goVerb:
			ctx.Verbs = append(ctx.Verbs, n)
		case goSumType:
			ctx.TypesCtx.SumTypes = append(ctx.TypesCtx.SumTypes, n)
		case goExternalType:
			ctx.TypesCtx.ExternalTypes = append(ctx.TypesCtx.ExternalTypes, n)
		case goDBHandle:
			ctx.Databases = append(ctx.Databases, n)
		}
		return errors.WithStack(next())
	})
	if err != nil {
		return errors.WithStack(err) //nolint:wrapcheck
	}
	return nil
}

func (b *mainDeploymentContextBuilder) getQueryDecls(node schema.Node) ([]queryVerb, []*schema.Data, error) {
	var verbs []queryVerb
	var data []*schema.Data
	err := schema.Visit(node, func(node schema.Node, next func() error) error {
		switch n := node.(type) {
		case *schema.Verb:
			verb, err := b.toQueryVerb(n)
			if err != nil {
				return errors.WithStack(err)
			}
			verbs = append(verbs, verb)
		case *schema.Data:
			refName := b.mainModule.Name + "." + n.Name
			if b.visited.Contains(refName) {
				return errors.WithStack(next())
			}
			b.visited.Add(refName)
			data = append(data, n)
		case *schema.Ref:
			maybeResolved, _ := b.sch.ResolveWithModule(n)
			resolved, ok := maybeResolved.Get()
			if !ok {
				return errors.WithStack(next())
			}
			nestedVerbs, nestedData, err := b.getQueryDecls(resolved)
			if err != nil {
				return errors.WithStack(err)
			}
			verbs = append(verbs, nestedVerbs...)
			data = append(data, nestedData...)
		default:

		}
		return errors.WithStack(next())
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get query decls")
	}
	return verbs, data, nil
}

func (b *mainDeploymentContextBuilder) toQueryVerb(verb *schema.Verb) (queryVerb, error) {
	request := b.getQueryRequestResponseData(verb.Request)
	response := b.getQueryRequestResponseData(verb.Response)

	var dbRef *schema.Ref
	for _, md := range verb.Metadata {
		if db, ok := md.(*schema.MetadataDatabases); ok {
			dbRef = db.Uses[0]
		}
	}
	if dbRef == nil || dbRef.Name == "" {
		return queryVerb{}, errors.Errorf("missing database call for query verb %s", verb.Name)
	}

	decl, ok := b.sch.Resolve(dbRef).Get()
	if !ok {
		return queryVerb{}, errors.Errorf("could not resolve database %s used by query verb %s", dbRef.String(), verb.Name)
	}
	db, ok := decl.(*schema.Database)
	if !ok {
		return queryVerb{}, errors.Errorf("declaration %s referenced by query verb %s is not a database", dbRef.String(), verb.Name)
	}

	var params []string
	if request != nil {
		for _, field := range request.Fields {
			if _, ok := islices.FindVariant[*schema.MetadataSQLColumn](field.Metadata); ok {
				// casing field name with the same mechanism as the generated code
				params = append(params, strings.Title(field.Name))
			}
		}
	}

	var pairs []string
	if response != nil {
		for _, field := range response.Fields {
			if md, ok := islices.FindVariant[*schema.MetadataSQLColumn](field.Metadata); ok {
				// casing field name with the same mechanism as the generated code
				pairs = append(pairs, fmt.Sprintf("tuple.PairOf(%q, %q)", md.Name, strings.Title(field.Name)))
			}
		}
	}

	sqlQuery, _ := verb.GetQuery()
	return queryVerb{
		Verb:           verb,
		CommandType:    strcase.ToUpperCamel(strings.TrimPrefix(sqlQuery.Command, ":")),
		RawSQL:         sqlQuery.Query,
		ParamFields:    fmt.Sprintf("[]string{%s}", strings.Join(islices.Map(params, strconv.Quote), ",")),
		ColToFieldName: fmt.Sprintf("[]tuple.Pair[string,string]{%s}", strings.Join(pairs, ",")),
		DBName:         db.Name,
		DBType:         db.Type,
	}, nil
}

func (b *mainDeploymentContextBuilder) getQueryRequestResponseData(reqResp schema.Type) *schema.Data {
	switch r := reqResp.(type) {
	case *schema.Ref:
		resolved, ok := b.sch.Resolve(r).Get()
		if !ok {
			return nil
		}
		return resolved.(*schema.Data) //nolint:forcetypeassert
	case *schema.Array:
		return b.getQueryRequestResponseData(r.Element)
	default:
		return nil
	}
}

func (b *mainDeploymentContextBuilder) getGoType(module *schema.Module, node schema.Node) (gotype optional.Option[goType], isLocal bool, err error) {
	isLocal = b.visitingMainModule(module.Name)
	switch n := node.(type) {
	case *schema.Verb:
		if !isLocal || n.IsGenerated() {
			return optional.None[goType](), false, nil
		}
		goverb, err := b.processVerb(n)
		if err != nil {
			return optional.None[goType](), isLocal, errors.WithStack(err)
		}
		return optional.Some[goType](goverb), isLocal, nil

	case *schema.Enum:
		if n.IsValueEnum() {
			return optional.None[goType](), isLocal, nil
		}
		st, err := b.processSumType(module, n)
		if err != nil {
			return optional.None[goType](), isLocal, errors.WithStack(err)
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
		db, err := b.processDatabase(module.Name, n)
		if err != nil {
			return optional.None[goType](), isLocal, errors.WithStack(err)
		}
		return optional.Some[goType](db), isLocal, nil

	default:
	}
	return optional.None[goType](), isLocal, nil
}

func (b *mainDeploymentContextBuilder) visitingMainModule(moduleName string) bool {
	return moduleName == b.mainModule.Name
}

func (b *mainDeploymentContextBuilder) processSumType(module *schema.Module, enum *schema.Enum) (out goSumType, err error) {
	defer func() {
		err = errors.WithStack(wrapErrWithPos(err, enum.Pos, ""))
	}()
	moduleName := module.Name
	var nt nativeType
	if !b.visitingMainModule(moduleName) {
		nt, err = nativeTypeFromQualifiedName("ftl/" + moduleName + "." + enum.Name)
	} else if nn, ok := b.nativeNames[enum]; ok {
		nt, err = b.getNativeType(nn)
	} else {
		return goSumType{}, errors.Errorf("missing native name for enum %s", enum.Name)
	}
	if err != nil {
		return goSumType{}, errors.WithStack(err)
	}

	variants := make([]goSumTypeVariant, 0, len(enum.Variants))
	for _, v := range enum.Variants {
		var vnt nativeType
		if !b.visitingMainModule(moduleName) {
			vnt, err = nativeTypeFromQualifiedName("ftl/" + moduleName + "." + v.Name)
		} else if nn, ok := b.nativeNames[v]; ok {
			vnt, err = b.getNativeType(nn)
		} else {
			return goSumType{}, errors.WithStack(wrapErrWithPos(errors.Errorf("missing native name for enum variant %s", enum.Name), v.Pos, ""))
		}
		if err != nil {
			return goSumType{}, errors.WithStack(wrapErrWithPos(err, v.Pos, ""))
		}

		typ, err := b.getGoSchemaType(v.Value.(*schema.TypeValue).Value)
		if err != nil {
			return goSumType{}, errors.WithStack(wrapErrWithPos(err, v.Pos, ""))
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
		return goVerb{}, errors.Errorf("missing verb resource param order for %s", verb.Name)
	}
	for _, m := range verbResourceParams {
		resource, err := b.getVerbResource(verb, m)
		if err != nil {
			return goVerb{}, errors.WithStack(err)
		}
		resources = append(resources, resource)
	}
	nativeName, ok := b.nativeNames[verb]
	if !ok {
		return goVerb{}, errors.Errorf("missing native name for verb %s", verb.Name)
	}
	return errors.WithStack2(b.getGoVerb(nativeName, verb, resources...))
}

func (b *mainDeploymentContextBuilder) getVerbResource(verb *schema.Verb, param common.VerbResourceParam) (verbResource, error) {
	if param.EgressTarget != "" {
		return verbEgress{Target: param.EgressTarget}, nil
	}
	ref := param.Ref
	resolved, ok := b.sch.Resolve(ref).Get()
	if !ok {
		return nil, errors.Errorf("failed to resolve %s resource, used by %s.%s", ref,
			b.mainModule.Name, verb.Name)
	}

	switch param.Type.(type) {
	case *schema.MetadataCalls:
		callee, ok := resolved.(*schema.Verb)
		if !ok {
			return verbClient{}, errors.Errorf("%s.%s uses %s client, but %s is not a verb",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		calleeNativeName, ok := b.nativeNames[ref]
		if !ok {
			return verbClient{}, errors.Errorf("missing native name for verb client %s", ref)
		}
		calleeverb, err := b.getGoVerb(calleeNativeName, callee)
		if err != nil {
			return verbClient{}, errors.WithStack(err)
		}
		return verbClient{
			calleeverb,
		}, nil
	case *schema.MetadataDatabases:
		db, ok := resolved.(*schema.Database)
		if !ok {
			return goDBHandle{}, errors.Errorf("%s.%s uses %s database handle, but %s is not a database",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return errors.WithStack2(b.processDatabase(ref.Module, db))
	case *schema.MetadataPublisher:
		topic, ok := resolved.(*schema.Topic)
		if !ok {
			return goTopicHandle{}, errors.Errorf("%s.%s uses %s topic handle, but %s is not a topic",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return errors.WithStack2(b.processTopic(ref.Module, ref, topic))
	case *schema.MetadataConfig:
		cfg, ok := resolved.(*schema.Config)
		if !ok {
			return goConfigHandle{}, errors.Errorf("%s.%s uses %s config handle, but %s is not a config",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return errors.WithStack2(b.processConfig(ref.Module, ref, cfg))
	case *schema.MetadataSecrets:
		secret, ok := resolved.(*schema.Secret)
		if !ok {
			return goSecretHandle{}, errors.Errorf("%s.%s uses %s secret handle, but %s is not a secret",
				b.mainModule.Name, verb.Name, ref, ref)
		}
		return errors.WithStack2(b.processSecret(ref.Module, ref, secret))

	default:
		return nil, errors.Errorf("unsupported resource type for verb %q", verb.Name)
	}
}

func (b *mainDeploymentContextBuilder) processConfig(moduleName string, ref *schema.Ref, config *schema.Config) (out goConfigHandle, err error) {
	defer func() {
		err = errors.WithStack(wrapErrWithPos(err, config.Pos, ""))
	}()
	nn, ok := b.nativeNames[ref]
	if !ok {
		return goConfigHandle{}, errors.Errorf("missing native name for config %s.%s", moduleName, config.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goConfigHandle{}, errors.WithStack(err)
	}

	ct, err := b.getGoSchemaType(config.Type)
	if err != nil {
		return goConfigHandle{}, errors.Wrapf(err, "failed to get config type for %s.%s", moduleName, config.Name)
	}
	return goConfigHandle{
		Name:       config.Name,
		Module:     moduleName,
		Type:       ct,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processSecret(moduleName string, ref *schema.Ref, secret *schema.Secret) (out goSecretHandle, err error) {
	defer func() {
		err = errors.WithStack(wrapErrWithPos(err, secret.Pos, ""))
	}()
	nn, ok := b.nativeNames[ref]
	if !ok {
		return goSecretHandle{}, errors.Errorf("missing native name for secret %s.%s", moduleName, secret.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goSecretHandle{}, errors.WithStack(err)
	}

	st, err := b.getGoSchemaType(secret.Type)
	if err != nil {
		return goSecretHandle{}, errors.Wrapf(err, "failed to get secret type for %s.%s", moduleName, secret.Name)
	}
	return goSecretHandle{
		Name:       secret.Name,
		Module:     moduleName,
		Type:       st,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processDatabase(moduleName string, db *schema.Database) (goDBHandle, error) {
	nt, err := b.getNativeType(fmt.Sprintf("ftl/%s.%s", b.mainModule.Name, fmt.Sprintf("%sConfig", strings.Title(db.Name))))
	if err != nil {
		return goDBHandle{}, errors.WithStack(err)
	}
	return goDBHandle{
		Name:       db.Name,
		Module:     moduleName,
		Type:       db.Type,
		nativeType: nt,
	}, nil
}

func (b *mainDeploymentContextBuilder) processTopic(moduleName string, ref *schema.Ref, topic *schema.Topic) (out goTopicHandle, err error) {
	defer func() {
		err = errors.WithStack(wrapErrWithPos(err, topic.Pos, ""))
	}()
	nn, ok := b.nativeNames[ref]
	if !ok {
		return goTopicHandle{}, errors.Errorf("missing native name for topic %s.%s", moduleName, topic.Name)
	}

	nt, err := b.getNativeType(nn)
	if err != nil {
		return goTopicHandle{}, errors.WithStack(err)
	}

	et, err := b.getGoSchemaType(topic.Event)
	if err != nil {
		return goTopicHandle{}, errors.Wrapf(err, "failed to get event type for topic %s.%s", moduleName, topic.Name)
	}
	mapperQualifiedNames, ok := b.topicMapperNames[topic]
	if !ok {
		return goTopicHandle{}, errors.Errorf("missing topic partition mapper name for topic %s.%s", moduleName, topic.Name)
	}
	mt, err := b.getNativeType(mapperQualifiedNames.Mapper)
	if err != nil {
		return goTopicHandle{}, errors.Wrapf(err, "failed to get event type for topic partition mapper %s.%s", moduleName, topic.Name)
	}
	var mapperAssociatedType optional.Option[nativeType]
	if a, ok := mapperQualifiedNames.AssociatedType.Get(); ok {
		at, err := b.getNativeType(a)
		if err != nil {
			return goTopicHandle{}, errors.Wrapf(err, "failed to get associated type for topic partition mapper %s.%s", moduleName, topic.Name)
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

func (b *mainDeploymentContextBuilder) getGoVerb(nativeName string, verb *schema.Verb, resources ...verbResource) (out goVerb, err error) {
	nt, err := b.getNativeType(nativeName)
	if err != nil {
		return goVerb{}, errors.WithStack(wrapErrWithPos(err, verb.Pos, ""))
	}
	req, err := b.getGoSchemaType(verb.Request)
	if err != nil {
		return goVerb{}, errors.WithStack(wrapErrWithPos(err, verb.Pos, "could not parse request type"))
	}
	resp, err := b.getGoSchemaType(verb.Response)
	if err != nil {
		return goVerb{}, errors.WithStack(wrapErrWithPos(err, verb.Pos, "could not parse response type"))
	}
	txn := optional.None[string]()
	if verb.IsTransaction() {
		dbSet := verb.ResolveDatabaseUses(b.sch, b.mainModule.Name)
		dbs := dbSet.ToSlice()
		if len(dbs) == 0 {
			return goVerb{}, errors.Errorf("transaction verbs must access a datasource; %s.%s does not access any",
				b.mainModule.Name, verb.Name)
		}
		if len(dbs) > 1 {
			return goVerb{}, errors.Errorf("transaction verbs can only access a single datasource; %s.%s accesses %d: %v",
				b.mainModule.Name, verb.Name, len(dbs), dbs)
		}
		txn = optional.Some(dbs[0].Name)
	}
	return goVerb{
		Name:                  verb.Name,
		nativeType:            nt,
		Request:               req,
		Response:              resp,
		Resources:             resources,
		TransactionDatasource: txn,
		IsQuery:               verb.IsQuery(),
	}, nil
}

func (b *mainDeploymentContextBuilder) getGoSchemaType(typ schema.Type) (out goSchemaType, err error) {
	defer func() {
		if typ != nil {
			err = errors.WithStack(wrapErrWithPos(err, typ.Position(), ""))
		}
	}()
	result := goSchemaType{
		children:   []goSchemaType{},
		nativeType: optional.None[nativeType](),
		schemaType: typ,
	}

	nn, ok := b.nativeNames[typ]
	if ok {
		nt, err := b.getNativeType(nn)
		if err != nil {
			return goSchemaType{}, errors.WithStack(err)
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
			fallbackImportPath := fmt.Sprintf("ftl/%s.%s", t.Module, t.Name)
			if decl := b.mainModule.Resolve(*t); decl != nil && decl.Symbol != nil {
				if s, ok := decl.Symbol.(*schema.Data); ok {
					if s.IsGenerated() {
						fallbackImportPath = fmt.Sprintf("ftl/%s/db.%s", t.Module, s.Name)
					}
				}
			}

			nt, err := b.getNativeType(fallbackImportPath)
			if err != nil {
				return goSchemaType{}, errors.WithStack(err)
			}
			result.nativeType = optional.Some(nt)
			b.nativeNames[typ] = fallbackImportPath
		}
		if len(t.TypeParameters) > 0 {
			for _, tp := range t.TypeParameters {
				_r, err := b.getGoSchemaType(tp)
				if err != nil {
					return goSchemaType{}, errors.WithStack(err)
				}
				result.children = append(result.children, _r)
			}
		}
	case *schema.Time:
		nt, err := b.getNativeType("time.Time")
		if err != nil {
			return goSchemaType{}, errors.WithStack(err)
		}
		result.nativeType = optional.Some(nt)
	case *schema.Array:
		e, err := b.getGoSchemaType(t.Element)
		if err != nil {
			return goSchemaType{}, errors.WithStack(err)
		}
		result.children = append(result.children, e)
	case *schema.Map:
		k, err := b.getGoSchemaType(t.Key)
		if err != nil {
			return goSchemaType{}, errors.WithStack(err)
		}
		result.children = append(result.children, k)

		v, err := b.getGoSchemaType(t.Value)
		if err != nil {
			return goSchemaType{}, errors.WithStack(err)
		}
		result.children = append(result.children, v)
	case *schema.Optional:
		e, err := b.getGoSchemaType(t.Type)
		if err != nil {
			return goSchemaType{}, errors.WithStack(err)
		}
		result.children = append(result.children, e)
	default:
	}

	result.TypeName, err = genTypeWithNativeNames(nil, typ, b.nativeNames)
	if err != nil {
		return goSchemaType{}, errors.WithStack(err)
	}
	result.LocalTypeName, err = genTypeWithNativeNames(b.mainModule, typ, b.nativeNames)
	if err != nil {
		return goSchemaType{}, errors.WithStack(err)
	}

	return result, nil
}

func (b *mainDeploymentContextBuilder) getNativeType(qualifiedName string) (nativeType, error) {
	nt, err := nativeTypeFromQualifiedName(qualifiedName)
	if err != nil {
		return nativeType{}, errors.WithStack(err)
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
				if !d.IsValueEnum() && d.GetVisibility().Exported() {
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
	"getEgressTarget": func(resource verbResource) *verbEgress {
		if c, ok := resource.(verbEgress); ok {
			return &c
		}
		return nil
	},
	"getRawSQLQuery": func(verb *schema.Verb) string {
		query, _ := verb.GetQuery()
		return query.Query
	},
	"queryResponseType": func(s *schema.Module, typ schema.Type) string {
		if arr, ok := typ.(*schema.Array); ok {
			return genType(s, arr.Element)
		}
		return genType(s, typ)
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

func genType(module *schema.Module, t schema.Type) string {
	typ, err := genTypeWithNativeNames(module, t, nil)
	if err != nil {
		panic(err)
	}
	return typ
}

// TODO: this is a hack because we don't currently qualify schema refs. Using native names for now to ensure
// even if the module is the same, we qualify the type with a package name when it's a subpackage.
func genTypeWithNativeNames(module *schema.Module, t schema.Type, nativeNames extract.NativeNames) (string, error) {
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
				tpDesc, err := genTypeWithNativeNames(module, tp, nativeNames)
				if err != nil {
					return "", errors.WithStack(err)
				}
				desc += tpDesc
			}
			desc += "]"
		}
		return desc, nil

	case *schema.Float:
		return "float64", nil

	case *schema.Time:
		return "stdtime.Time", nil

	case *schema.Int, *schema.Bool, *schema.String:
		return strings.ToLower(t.String()), nil

	case *schema.Array:
		element, err := genTypeWithNativeNames(module, t.Element, nativeNames)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return "[]" + element, nil

	case *schema.Map:
		key, err := genTypeWithNativeNames(module, t.Key, nativeNames)
		if err != nil {
			return "", errors.WithStack(err)
		}
		value, err := genTypeWithNativeNames(module, t.Value, nativeNames)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return "map[" + key + "]" + value, nil

	case *schema.Optional:
		typ, err := genTypeWithNativeNames(module, t.Type, nativeNames)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return "ftl.Option[" + typ + "]", nil

	case *schema.Unit:
		return "ftl.Unit", nil

	case *schema.Any:
		return "any", nil

	case *schema.Bytes:
		return "[]byte", nil

	case *schema.Data, *schema.Enum, *schema.TypeAlias:
		panic(fmt.Sprintf("unsupported type %T, this should never be reached", t))
	}
	return "", errors.Errorf("unsupported type %T", t)
}

// Update go.mod file to include the FTL version and return the Go version and any replace directives.
func updateGoModule(goModPath string, expectedModule string, transaction optional.Option[watch.ModifyFilesTransaction]) (replacements []*modfile.Replace, goVersion string, err error) {
	goModFile, replacements, err := goModFileWithReplacements(goModPath)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to update %s", goModPath)
	}

	// Validate module name matches
	if !strings.HasPrefix(goModFile.Module.Mod.Path, "ftl/"+expectedModule) {
		return nil, "", errors.Errorf("module name mismatch: expected 'ftl/%s' but got '%s' in go.mod",
			expectedModule, goModFile.Module.Mod.Path)
	}

	// Early return if we're not updating anything.
	if !ftl.IsRelease(ftl.Version) || !shouldUpdateVersion(goModFile) {
		return replacements, goModFile.Go.Version, nil
	}
	if tx, ok := transaction.Get(); ok {
		err := tx.ModifiedFiles(goModPath)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to add go.mod to transactio")
		}
	}

	if err := goModFile.AddRequire("github.com/block/ftl", "v"+ftl.Version); err != nil {
		return nil, "", errors.Wrapf(err, "failed to add github.com/block/ftl to %s", goModPath)
	}

	// Atomically write the updated go.mod file.
	tmpFile, err := os.CreateTemp(filepath.Dir(goModPath), ".go.mod-")
	if err != nil {
		return nil, "", errors.Wrapf(err, "update %s", goModPath)
	}
	defer os.Remove(tmpFile.Name()) // Delete the temp file if we error.
	defer tmpFile.Close()
	goModBytes := modfile.Format(goModFile.Syntax)
	if _, err := tmpFile.Write(goModBytes); err != nil {
		return nil, "", errors.Wrapf(err, "update %s", goModPath)
	}
	if err := os.Rename(tmpFile.Name(), goModPath); err != nil {
		return nil, "", errors.Wrapf(err, "update %s", goModPath)
	}
	return replacements, goModFile.Go.Version, nil
}

func goModFileWithReplacements(goModPath string) (*modfile.File, []*modfile.Replace, error) {
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to read %s", goModPath)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse %s", goModPath)
	}

	replacements := reflect.DeepCopy(goModFile.Replace)
	for i, r := range replacements {
		if strings.HasPrefix(r.New.Path, ".") {
			abs, err := filepath.Abs(filepath.Join(filepath.Dir(goModPath), r.New.Path))
			if err != nil {
				return nil, nil, errors.WithStack(err)
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

// returns the import path and the directory name for a Go type
// package and directory names are the same (dir=bar, pkg=bar): "github.com/foo/bar.A" => "github.com/foo/bar", none
// package and directory names differ (dir=bar, pkg=baz): "github.com/foo/bar.baz.A" => "github.com/foo/bar", "baz"
func nativeTypeFromQualifiedName(qualifiedName string) (nativeType, error) {
	lastDotIndex := strings.LastIndex(qualifiedName, ".")
	if lastDotIndex == -1 {
		return nativeType{}, errors.Errorf("invalid qualified type format %q", qualifiedName)
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

func usesType(actual schema.Type, expected schema.Type) bool {
	if stdreflect.TypeOf(actual) == stdreflect.TypeOf(expected) {
		return true
	}
	switch t := actual.(type) {
	case *schema.Optional:
		return usesType(t.Type, expected)
	case *schema.Array:
		return usesType(t.Element, expected)
	default:

	}
	return false
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
			if n.GetVisibility().Exported() {
				imports["github.com/block/ftl/go-runtime/ftl"] = "ftl"
			}

		case *schema.TypeAlias:
			if aliasesMustBeExported && !n.GetVisibility().Exported() {
				return errors.WithStack(next())
			}
			if nt, ok := nativeTypeForWidenedType(n); ok {
				if existing, ok := extraImports[nt.importPath]; !ok || !existing.importAlias {
					extraImports[nt.importPath] = nt
				}
			}
		default:
		}
		return errors.WithStack(next())
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

func wrapErrWithPos(err error, pos schema.Position, format string) error {
	if err == nil {
		return nil
	}
	var zero = schema.Position{}
	if pos == zero {
		return errors.WithStack(err)
	}
	return errors.WithStack(builderrors.Wrapf(err, pos.ToErrorPos(), format))
}
