package verb

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/go-runtime/schema/common"
	"github.com/block/ftl/go-runtime/schema/initialize"
)

type resource struct {
	ref *schema.Ref
	typ common.VerbResourceType
}

func (r resource) toMetadataType() (schema.Metadata, error) {
	switch r.typ {
	case common.VerbResourceTypeVerbClient:
		return &schema.MetadataCalls{}, nil
	case common.VerbResourceTypeDatabaseHandle:
		return &schema.MetadataDatabases{}, nil
	case common.VerbResourceTypeTopicHandle:
		return &schema.MetadataPublisher{}, nil
	case common.VerbResourceTypeConfig:
		return &schema.MetadataConfig{}, nil
	case common.VerbResourceTypeSecret:
		return &schema.MetadataSecrets{}, nil
	case common.VerbResourceTypeEgress:
		return &schema.MetadataEgress{}, nil
	default:
		return nil, errors.Errorf("unsupported resource type")
	}
}

// Extractor extracts verbs to the module schema.
var Extractor = common.NewDeclExtractor[*schema.Verb, *ast.FuncDecl]("verb", Extract)

func Extract(pass *analysis.Pass, node *ast.FuncDecl, obj types.Object) optional.Option[*schema.Verb] {
	verb := &schema.Verb{
		Pos:  common.GoPosToSchemaPos(pass, node.Pos()),
		Name: strcase.ToLowerCamel(node.Name.Name),
	}

	loaded := pass.ResultOf[initialize.Analyzer].(initialize.Result) //nolint:forcetypeassert

	hasRequest := false
	var orderedResourceParams []common.VerbResourceParam
	if !common.ApplyMetadata[*schema.Verb](pass, obj, func(md *common.ExtractedMetadata) {
		_, isTransaction := slices.FindVariant[*schema.MetadataTransaction](md.Metadata)
		verb.Comments = md.Comments
		verb.Visibility = schema.Visibility(md.Visibility)
		verb.Metadata = md.Metadata
		for idx, param := range node.Type.Params.List {
			var err error
			var r *resource
			egress := ""
			params := common.GetFactForObject[*common.EgressParams](pass, obj)
			if p, ok := params.Get(); ok {
				egress = p.Targets[param.Names[0].Name]
				if egress != "" {
					r = &resource{
						typ: common.VerbResourceTypeEgress,
					}
				}
			}
			if r == nil {
				r, err = resolveResource(pass, param.Type)
				if err != nil {
					common.Wrapf(pass, param, err, "")
					continue
				}
			}

			// if this parameter can't be resolved to a resource, it must either be the context or request parameter:
			// Verb(context.Context, <request>, <resource1>, <resource2>, ...)
			if r == nil || r.typ == common.VerbResourceTypeNone {
				if idx > 1 {
					common.Errorf(pass, param, "unsupported verb parameter type in verb %s at parameter %d; verbs must have the "+
						"signature func(Context, Request?, Resources...)", verb.Name, idx)
					continue
				}
				if idx == 1 {
					hasRequest = true
				}
				continue
			}

			paramObj, ok := common.GetObjectForNode(pass.TypesInfo, param.Type).Get()
			if !ok {
				common.Errorf(pass, param, "unsupported verb parameter type")
				continue
			}
			if r.ref != nil {
				common.MarkIncludeNativeName(pass, paramObj, r.ref)
			}
			switch r.typ {
			case common.VerbResourceTypeVerbClient:
				if currentModule, err := common.FtlModuleFromGoPackage(pass.Pkg.Path()); err == nil && isTransaction && r.ref.Module != currentModule {
					common.Errorf(pass, param, "could not inject external verb %q because calling external verbs from within a transaction is not allowed", r.ref)
					continue
				}
				verb.AddCall(r.ref)
			case common.VerbResourceTypeDatabaseHandle:
				verb.AddDatabase(r.ref)
			case common.VerbResourceTypeTopicHandle:
				if currentModule, err := common.FtlModuleFromGoPackage(pass.Pkg.Path()); err == nil && r.ref.Module != currentModule {
					common.Errorf(pass, param, "could not inject external topic %q because publishing directly to external topics is not allowed", r.ref)
					continue
				}
				verb.AddTopicPublish(r.ref)
			case common.VerbResourceTypeConfig:
				verb.AddConfig(r.ref)
			case common.VerbResourceTypeSecret:
				verb.AddSecret(r.ref)
			case common.VerbResourceTypeEgress:
				verb.AddEgress(egress)
			default:
				common.Errorf(pass, param, "unsupported verb parameter type in verb %s at parameter %d; verbs must have the "+
					"signature func(Context, Request?, Resources...)", verb.Name, idx)
			}
			rt, err := r.toMetadataType()
			if err != nil {
				common.Wrapf(pass, param, err, "")
			}
			orderedResourceParams = append(orderedResourceParams, common.VerbResourceParam{
				Ref:          r.ref,
				Type:         rt,
				EgressTarget: egress,
			})
		}
	}) {
		return optional.None[*schema.Verb]()
	}

	fnt := obj.(*types.Func)             //nolint:forcetypeassert
	sig := fnt.Type().(*types.Signature) //nolint:forcetypeassert
	if sig.Recv() != nil {
		common.Errorf(pass, node, "ftl:verb cannot be a method")
		return optional.None[*schema.Verb]()
	}

	reqt, respt := checkSignature(pass, loaded, node, sig, hasRequest)
	req := optional.Some[schema.Type](&schema.Unit{})
	if reqt.Ok() {
		req = common.ExtractType(pass, node.Type.Params.List[1].Type)
	}

	resp := optional.Some[schema.Type](&schema.Unit{})
	if respt.Ok() {
		resp = common.ExtractType(pass, node.Type.Results.List[0].Type)
	}

	params := sig.Params()
	results := sig.Results()
	reqV, ok := req.Get()
	if !ok {
		common.Errorf(pass, node.Type.Params.List[1], "unsupported request type %q", params.At(1).Type())
	}
	resV, ok := resp.Get()
	if !ok {
		common.Errorf(pass, node.Type.Results.List[0], "unsupported response type %q", results.At(0).Type())
	}
	verb.Request = reqV
	verb.Response = resV

	common.MarkVerbResourceParamOrder(pass, obj, orderedResourceParams)
	return optional.Some(verb)
}

func resolveResource(pass *analysis.Pass, typ ast.Expr) (*resource, error) {
	obj, hasObj := common.GetObjectForNode(pass.TypesInfo, typ).Get()
	rType := common.GetVerbResourceType(pass, obj)
	var ref *schema.Ref
	switch rType {
	case common.VerbResourceTypeNone:
		return nil, nil
	case common.VerbResourceTypeVerbClient:
		calleeRef, ok := common.ExtractSimpleRefWithCasing(pass, typ, strcase.ToLowerCamel).Get()
		if !ok {
			return nil, errors.Errorf("unsupported verb parameter type")
		}
		calleeRef.Name = strings.TrimSuffix(calleeRef.Name, "Client")
		ref = calleeRef
	case common.VerbResourceTypeDatabaseHandle:
		if !hasObj {
			return nil, errors.Errorf("unsupported verb parameter type; expected generated database handle")
		}
		var dbObj types.Object
		if alias, ok := obj.Type().(*types.Alias); ok {
			named, ok := alias.Rhs().(*types.Named)
			if !ok {
				return nil, errors.Errorf("unsupported verb parameter type; expected generated database handle")
			}

			typeArgs := named.TypeArgs()
			if typeArgs.Len() == 0 {
				return nil, errors.Errorf("unsupported verb parameter type; expected generated database handle")
			}
			configType := typeArgs.At(0)
			named, ok = configType.(*types.Named)
			if !ok {
				return nil, errors.Errorf("unsupported verb parameter type; expected generated database handle")
			}
			dbObj = named.Obj()
		} else {
			idxExpr, ok := typ.(*ast.IndexExpr)
			if !ok {
				return nil, errors.Errorf("unsupported verb parameter type; expected generated database handle")
			}
			ident, ok := idxExpr.Index.(*ast.Ident)
			if !ok {
				return nil, errors.Errorf("unsupported verb parameter type; expected generated database handle")
			}
			dbObj, ok = common.GetObjectForNode(pass.TypesInfo, ident).Get()
			if !ok {
				return nil, errors.Errorf("unsupported verb parameter type")
			}
		}
		dbName, ok := common.GetFactForObject[*common.DatabaseConfig](pass, dbObj).Get()
		if !ok {
			return nil, errors.Errorf("no database name found for verb parameter")
		}
		module, err := common.FtlModuleFromGoPackage(dbObj.Pkg().Path())
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve module for type")
		}
		ref = &schema.Ref{
			Module: module,
			Name:   dbName.Name,
		}
	case common.VerbResourceTypeTopicHandle, common.VerbResourceTypeSecret, common.VerbResourceTypeConfig:
		var ok bool
		if ref, ok = common.ExtractSimpleRefWithCasing(pass, typ, strcase.ToLowerCamel).Get(); !ok {
			return nil, errors.Errorf("unsupported verb parameter type; expected ftl.TopicHandle[Event, PartitionMapper]")
		}
	}
	if ref == nil {
		return nil, errors.Errorf("unsupported verb parameter type")
	}
	return &resource{ref: ref, typ: rType}, nil
}

func checkSignature(
	pass *analysis.Pass,
	loaded initialize.Result,
	node *ast.FuncDecl,
	sig *types.Signature,
	hasRequest bool,
) (req, resp optional.Option[*types.Var]) {
	if node.Name.Name == "" {
		common.Errorf(pass, node, "verb function must be named")
		return optional.None[*types.Var](), optional.None[*types.Var]()
	}
	if !unicode.IsUpper(rune(node.Name.Name[0])) {
		common.Errorf(pass, node, "verb function name must be exported (i.e. %s instead of %s)", strcase.ToUpperCamel(node.Name.Name), node.Name.Name)
		return optional.None[*types.Var](), optional.None[*types.Var]()
	}

	params := sig.Params()
	results := sig.Results()
	if params.Len() == 0 {
		common.Errorf(pass, node, "first parameter must be context.Context")
	} else if !loaded.IsContextType(params.At(0).Type()) {
		common.TokenErrorf(pass, params.At(0).Pos(), params.At(0).Name(), "first parameter must be of type context.Context but is %s", params.At(0).Type())
	}

	if params.Len() >= 2 {
		if params.At(1).Type().String() == common.FtlUnitTypePath {
			common.TokenErrorf(pass, params.At(1).Pos(), params.At(1).Name(), "second parameter must not be ftl.Unit (verbs without a request type should omit the request parameter)")
		}

		if hasRequest {
			req = optional.Some(params.At(1))
		}
	}

	if results.Len() > 2 {
		common.Errorf(pass, node, "must have at most two results (<type>, error)")
	}
	if results.Len() == 0 {
		common.Errorf(pass, node, "must at least return an error")
	} else if !loaded.IsStdlibErrorType(results.At(results.Len() - 1).Type()) {
		common.TokenErrorf(pass, results.At(results.Len()-1).Pos(), results.At(results.Len()-1).Name(), "must return an error but is %q", results.At(0).Type())
	}
	if results.Len() == 2 {
		if results.At(1).Type().String() == common.FtlUnitTypePath {
			common.TokenErrorf(pass, results.At(1).Pos(), results.At(1).Name(), "second result must not be ftl.Unit")
		}
		resp = optional.Some(results.At(0))
	}
	return req, resp
}
