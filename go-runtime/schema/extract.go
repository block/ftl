package schema

import (
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/tuple"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"

	"github.com/block/ftl-golang-tools/go/analysis"
	checker "github.com/block/ftl-golang-tools/go/analysis/checker"
	"github.com/block/ftl-golang-tools/go/analysis/passes/inspect"
	"github.com/block/ftl-golang-tools/go/packages"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/go-runtime/schema/call"
	"github.com/block/ftl/go-runtime/schema/common"
	"github.com/block/ftl/go-runtime/schema/config"
	"github.com/block/ftl/go-runtime/schema/data"
	"github.com/block/ftl/go-runtime/schema/enum"
	"github.com/block/ftl/go-runtime/schema/finalize"
	"github.com/block/ftl/go-runtime/schema/initialize"
	"github.com/block/ftl/go-runtime/schema/metadata"
	"github.com/block/ftl/go-runtime/schema/secret"
	"github.com/block/ftl/go-runtime/schema/topic"
	"github.com/block/ftl/go-runtime/schema/transitive"
	"github.com/block/ftl/go-runtime/schema/typealias"
	"github.com/block/ftl/go-runtime/schema/typeenum"
	"github.com/block/ftl/go-runtime/schema/typeenumvariant"
	"github.com/block/ftl/go-runtime/schema/valueenumvariant"
	"github.com/block/ftl/go-runtime/schema/verb"
)

// extractors contains all schema extractors that will run.
//
// It is a list of lists, where each list is a round of tasks dependent on the prior round's execution (e.g. an analyzer
// in extractors[1] will only execute once all analyzers in extractors[0] complete). Elements of the same list
// should be considered unordered and may run in parallel.
var extractors = [][]*analysis.Analyzer{
	{
		initialize.Analyzer,
		inspect.Analyzer,
	},
	{
		metadata.Extractor,
	},
	{
		// must run before typeenumvariant.Extractor; typeenum.Extractor determines all possible discriminator
		// interfaces and typeenumvariant.Extractor determines any types that implement these
		typeenum.Extractor,
	},
	{
		config.Extractor,
		data.Extractor,
		topic.Extractor,
		typealias.Extractor,
		typeenumvariant.Extractor,
		valueenumvariant.Extractor,
		secret.Extractor,
	},
	{
		call.Extractor,
		// must run after valueenumvariant.Extractor and typeenumvariant.Extractor;
		// visits a node and aggregates its enum variants if present
		enum.Extractor,
		verb.Extractor,
	},
	{
		transitive.Extractor,
	},
	{
		finalize.Analyzer,
	},
}

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Node]string

// Result contains the final schema extraction result.
type Result struct {
	// Module is the extracted module schema.
	Module *schema.Module
	// NativeNames maps schema nodes to their native Go names.
	NativeNames NativeNames
	// VerbResourceParamOrder contains the order of resource parameters for each verb.
	VerbResourceParamOrder map[*schema.Verb][]common.VerbResourceParam
	// TopicPartitionMapperNames maps topics to the native name of the partition mapper.
	TopicPartitionMapperNames map[*schema.Topic]finalize.TopicMapperQualifiedNames
	// Errors is a list of errors encountered during schema extraction.
	Errors []builderrors.Error
}

var orderedAnalyzers []*analysis.Analyzer

func init() {
	// observes dependencies as specified by tiered list ordering in Extractors and applies the dependency
	// requirements to the analyzers
	//
	// flattens Extractors (a list of lists) into a single list to provide as input for the checker
	var beforeIndex []*analysis.Analyzer
	for i, extractorRound := range extractors {
		for _, extractor := range extractorRound {
			extractor.RunDespiteErrors = true
			beforeIndex = dependenciesBeforeIndex(i)
			extractor.Requires = append(extractor.Requires, beforeIndex...)
			orderedAnalyzers = append(orderedAnalyzers, extractor)
		}
	}
}

// Extract statically parses Go FTL module source into a schema.Module
func Extract(moduleDir string, sch *schema.Schema) (Result, error) {
	pkgConfig := packages.Config{
		Dir:  moduleDir,
		Mode: packages.LoadTypes | packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}
	pkgs, err := packages.Load(&pkgConfig, "./...")
	if err != nil {
		return Result{}, fmt.Errorf("failed to load packages: %w", err)
	}
	graph, err := checker.Analyze(orderedAnalyzers, pkgs, &checker.Options{
		ReverseImportExecutionOrder: true,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to analyze module packages: %w", err)
	}

	var errors []builderrors.Error
	finalizeResults := []finalize.Result{}
	for act := range graph.All() {
		for _, d := range act.Diagnostics {
			errors = append(errors, builderrors.Error{
				Pos:   optional.Some(toErrorPos(act.Package.Fset.Position(d.Pos), act.Package.Fset.Position(d.End))),
				Msg:   d.Message,
				Level: common.DiagnosticCategory(d.Category).ToErrorLevel(),
			})
		}
		if act.Analyzer == finalize.Analyzer {
			fr, ok := act.Result.(finalize.Result)
			if !ok {
				return Result{}, fmt.Errorf("unexpected schema extraction result type: %T", act.Result)
			}
			finalizeResults = append(finalizeResults, fr)
		}
	}
	if len(finalizeResults) == 0 {
		return Result{}, fmt.Errorf("schema extraction finalizer result not found")
	}

	return combineAllPackageResults(sch, finalizeResults, errors)
}

type refResultType int

const (
	failed refResultType = iota
	widened
)

type refResult struct {
	typ    refResultType
	obj    types.Object
	fqName optional.Option[string]
}

// used to combine result data across passes (each pass analyzes one package within the module)
type combinedData struct {
	module *schema.Module
	errs   []builderrors.Error

	nativeNames            NativeNames
	functionCalls          map[schema.Position]finalize.FunctionCall
	verbs                  map[types.Object]*schema.Verb
	verbResourceParamOrder map[*schema.Verb][]common.VerbResourceParam
	refResults             map[schema.RefKey]refResult
	extractedDecls         map[schema.Decl]types.Object
	externalTypeAliases    sets.Set[*schema.TypeAlias]
	topicPartitionMaps     map[*schema.Topic]finalize.TopicMapperQualifiedNames
	// for detecting duplicates
	typeUniqueness   map[string]tuple.Pair[types.Object, schema.Position]
	globalUniqueness map[string]tuple.Pair[types.Object, schema.Position]
}

func newCombinedData(diagnostics []builderrors.Error) *combinedData {
	return &combinedData{
		errs:                   diagnostics,
		nativeNames:            make(NativeNames),
		functionCalls:          make(map[schema.Position]finalize.FunctionCall),
		verbs:                  make(map[types.Object]*schema.Verb),
		verbResourceParamOrder: make(map[*schema.Verb][]common.VerbResourceParam),
		refResults:             make(map[schema.RefKey]refResult),
		extractedDecls:         make(map[schema.Decl]types.Object),
		externalTypeAliases:    sets.NewSet[*schema.TypeAlias](),
		topicPartitionMaps:     make(map[*schema.Topic]finalize.TopicMapperQualifiedNames),
		typeUniqueness:         make(map[string]tuple.Pair[types.Object, schema.Position]),
		globalUniqueness:       make(map[string]tuple.Pair[types.Object, schema.Position]),
	}
}

func (cd *combinedData) error(err builderrors.Error) {
	cd.errs = append(cd.errs, err)
}

func (cd *combinedData) update(fr finalize.Result) {
	for decl, obj := range fr.Extracted {
		cd.validateDecl(decl, obj, false)
		cd.extractedDecls[decl] = obj
	}
	copyFailedRefs(cd.refResults, fr.Failed)
	maps.Copy(cd.nativeNames, fr.NativeNames)
	maps.Copy(cd.topicPartitionMaps, fr.TopicPartitionMaps)
	maps.Copy(cd.functionCalls, fr.FunctionCalls)
	maps.Copy(cd.verbResourceParamOrder, fr.VerbResourceParamOrder)
}

func (cd *combinedData) toResult() Result {
	cd.module.AddDecls(maps.Keys(cd.extractedDecls))
	cd.updateDeclVisibility()
	cd.propagateTypeErrors()
	cd.errorDirectVerbInvocations()
	builderrors.SortErrorsByPosition(cd.errs)
	return Result{
		Module:                    cd.module,
		NativeNames:               cd.nativeNames,
		VerbResourceParamOrder:    cd.verbResourceParamOrder,
		TopicPartitionMapperNames: cd.topicPartitionMaps,
		Errors:                    cd.errs,
	}
}

func (cd *combinedData) updateModule(fr finalize.Result) error {
	if cd.module == nil {
		cd.module = &schema.Module{Name: fr.ModuleName, Comments: fr.ModuleComments}
	} else {
		if cd.module.Name != fr.ModuleName {
			return fmt.Errorf("unexpected schema extraction result module name: %s", fr.ModuleName)
		}
		if len(cd.module.Comments) == 0 {
			cd.module.Comments = fr.ModuleComments
		}
	}
	return nil
}

func (cd *combinedData) validateDecl(decl schema.Decl, obj types.Object, suppressErrors bool) bool {
	valid := true
	typeKey := getTypeUniquenessKey(decl)
	if value, ok := cd.typeUniqueness[typeKey]; ok && value.A != obj {
		if !suppressErrors {
			cd.error(builderrors.Errorf(decl.Position().ToErrorPos(),
				"duplicate %s declaration for %q; already declared at %q", common.GetDeclTypeName(decl),
				cd.module.Name+"."+decl.GetName(), value.B))
		}
		valid = false
	} else if value, ok := cd.globalUniqueness[decl.GetName()]; ok && value.A != obj {
		if !suppressErrors {
			cd.error(builderrors.Errorf(decl.Position().ToErrorPos(),
				"schema declaration with name %q already exists for module %q; previously declared at %q",
				decl.GetName(), cd.module.Name, value.B))
		}
		valid = false
	} else {
		cd.typeUniqueness[typeKey] = tuple.Pair[types.Object, schema.Position]{A: obj, B: decl.Position()}
		cd.globalUniqueness[decl.GetName()] = tuple.Pair[types.Object, schema.Position]{A: obj, B: decl.Position()}
	}
	return valid
}

func (cd *combinedData) errorDirectVerbInvocations() {
	for pos, fnCall := range cd.functionCalls {
		if v, ok := cd.verbs[fnCall.Callee]; ok {
			cd.error(builderrors.Errorf(pos.ToErrorPos(),
				"direct verb calls are not allowed; use the provided %sClient instead. "+
					"See https://block.github.io/ftl/docs/reference/verbs/#calling-verbs",
				strcase.ToUpperCamel(v.Name)))
		}
	}
}

// updateDeclVisibility traverses the module schema via refs and updates visibility as needed.
func (cd *combinedData) updateDeclVisibility() {
	for _, d := range cd.module.Decls {
		if d.IsExported() {
			updateTransitiveVisibility(d, cd.module)
		}
	}
}

// propagateTypeErrors propagates type errors to referencing nodes. This improves error messaging for the LSP client by
// surfacing errors all the way up the schema chain.
func (cd *combinedData) propagateTypeErrors() {
	_ = schema.VisitWithParents(cd.module, nil, func(n schema.Node, ps []schema.Node, next func() error) error { //nolint:errcheck
		if len(ps) == 0 {
			return next()
		}
		p := ps[len(ps)-1]
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}

		result, ok := cd.refResults[ref.ToRefKey()]
		if !ok {
			return next()
		}

		switch result.typ {
		case failed:
			refNativeName := common.GetNativeName(result.obj)
			switch pt := p.(type) {
			case *schema.Verb:
				if pt.Request == n {
					cd.error(builderrors.Errorf(pt.Request.Position().ToErrorPos(),
						"unsupported request type %q", refNativeName))
				}
				if pt.Response == n {
					cd.error(builderrors.Errorf(pt.Response.Position().ToErrorPos(),
						"unsupported response type %q", refNativeName))
				}
			case *schema.Field:
				cd.error(builderrors.Errorf(pt.Position().ToErrorPos(), "unsupported type %q for "+
					"field %q", refNativeName, pt.Name))
			default:
				cd.error(builderrors.Errorf(pt.Position().ToErrorPos(), "unsupported type %q",
					refNativeName))
			}
		case widened:
			cd.error(builderrors.Warnf(n.Position().ToErrorPos(), "external type %q will be "+
				"widened to Any", result.fqName.MustGet()))
		}

		return next()
	})
}

func dependenciesBeforeIndex(idx int) []*analysis.Analyzer {
	var deps []*analysis.Analyzer
	for i := range idx {
		deps = append(deps, extractors[i]...)
	}
	return deps
}

func combineAllPackageResults(sch *schema.Schema, finalizeResults []finalize.Result, errors []builderrors.Error) (Result, error) {
	cd := newCombinedData(errors)

	for _, fr := range finalizeResults {
		if err := cd.updateModule(fr); err != nil {
			return Result{}, err
		}
		cd.update(fr)
	}

	for decl, obj := range cd.extractedDecls {
		moduleName := cd.module.Name
		switch d := decl.(type) {
		case *schema.TypeAlias:
			if len(d.Metadata) > 0 {
				fqName, err := goQualifiedNameForWidenedType(obj, d.Metadata)
				if err != nil {
					cd.error(builderrors.Error{
						Pos:   optional.Some(d.Position().ToErrorPos()),
						Msg:   err.Error(),
						Level: builderrors.ERROR})
				}
				cd.refResults[schema.RefKey{Module: moduleName, Name: d.Name}] = refResult{typ: widened, obj: obj,
					fqName: optional.Some(fqName)}
				cd.externalTypeAliases.Add(d)
				cd.nativeNames[d] = common.GetNativeName(obj)
			}
		case *schema.Verb:
			cd.verbs[obj] = d

		default:
		}
	}

	// add existing schema decls to the result, validating that there are no conflicts
	for _, generatedModule := range sch.Modules {
		if generatedModule.Name == cd.module.Name {
			for _, decl := range generatedModule.Decls {
				if !decl.IsGenerated() {
					continue
				}
				if cd.validateDecl(decl, nil, true) {
					cd.extractedDecls[decl] = nil
				} else {
					var genType string
					switch decl.(type) {
					case *schema.Verb:
						genType = "query verb"
					case *schema.Database:
						genType = "database"
					default:
						return Result{}, fmt.Errorf("%q is an unsupported generated type: %T", decl.GetName(), decl)
					}

					typeKey := getTypeUniquenessKey(decl)
					if value, ok := cd.typeUniqueness[typeKey]; ok {
						displayType := strings.ToLower(strings.Split(reflect.TypeOf(decl).String(), ".")[1])
						cd.error(builderrors.Errorf(value.B.ToErrorPos(), "declared %s %q conflicts with FTL-generated %s of the same name", displayType, decl.GetName(), genType))
					} else if value, ok := cd.globalUniqueness[decl.GetName()]; ok {
						cd.error(builderrors.Errorf(value.B.ToErrorPos(), "declaration %q conflicts with FTL-generated %s of the same name", decl, genType))
					}
				}
			}
		}
	}

	result := cd.toResult()
	if builderrors.ContainsTerminalError(result.Errors) {
		return result, nil
	}
	return result, result.Module.Validate() //nolint:wrapcheck
}

// updateTransitiveVisibility updates any decls that are transitively visible from d.
func updateTransitiveVisibility(d schema.Decl, module *schema.Module) {
	if !d.IsExported() {
		return
	}

	// exclude metadata children so we don't update callees to be exported if their callers are
	_ = schema.VisitExcludingMetadataChildren(d, func(n schema.Node, next func() error) error { //nolint:errcheck
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}

		resolved := module.Resolve(*ref)
		if resolved == nil || resolved.Symbol == nil {
			return next()
		}

		if decl, ok := resolved.Symbol.(schema.Decl); ok {
			switch t := decl.(type) {
			case *schema.Data:
				t.Export = true
			case *schema.Enum:
				t.Export = true
			case *schema.TypeAlias:
				t.Export = true
			case *schema.Topic:
				t.Export = true
			case *schema.Verb:
				t.Export = true
			case *schema.Database, *schema.Config, *schema.Secret:
			}
			updateTransitiveVisibility(decl, module)
		}
		return next()
	})
}

func copyFailedRefs(parsedRefs map[schema.RefKey]refResult, failedRefs map[schema.RefKey]types.Object) {
	for ref, obj := range failedRefs {
		parsedRefs[ref] = refResult{typ: failed, obj: obj}
	}
}

func goQualifiedNameForWidenedType(obj types.Object, metadata []schema.Metadata) (string, error) {
	var nativeName string
	for _, m := range metadata {
		if m, ok := m.(*schema.MetadataTypeMap); ok && m.Runtime == "go" {
			if nativeName != "" {
				return "", fmt.Errorf("multiple Go type mappings found for %q", common.GetNativeName(obj))
			}
			nativeName = m.NativeName
		}
	}
	if len(metadata) > 0 && nativeName == "" {
		return "", fmt.Errorf("missing Go native name in typemapped alias for %q",
			common.GetNativeName(obj))
	}
	return nativeName, nil
}

func toErrorPos(pos token.Position, end token.Position) builderrors.Position {
	return builderrors.Position{
		Filename:    pos.Filename,
		Line:        pos.Line,
		StartColumn: pos.Column,
		EndColumn:   end.Column,
	}
}

func getTypeUniquenessKey(decl schema.Decl) string {
	typename := common.GetDeclTypeName(decl)
	return fmt.Sprintf("%s-%s", typename, decl.GetName())
}
