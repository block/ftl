package metadata

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl-golang-tools/go/analysis/passes/inspect"
	"github.com/block/ftl-golang-tools/go/ast/inspector"
	ireflect "github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/go-runtime/schema/common"
)

// Extractor extracts metadata to the module schema.
var Extractor = common.NewExtractor("metadata", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.TypeSpec)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.FuncDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		var doc *ast.CommentGroup
		switch n := n.(type) {
		case *ast.TypeSpec:
			doc = n.Doc
		case *ast.ValueSpec:
			doc = n.Doc
		case *ast.GenDecl:
			doc = n.Doc
			if len(n.Specs) == 0 {
				return
			}
			if ts, ok := n.Specs[0].(*ast.TypeSpec); len(n.Specs) > 0 && ok {
				if doc == nil {
					doc = ts.Doc
				}
			}
		case *ast.FuncDecl:
			doc = n.Doc
		}
		obj, ok := common.GetObjectForNode(pass.TypesInfo, n).Get()
		if !ok {
			return
		}
		if mdFact, ok := extractMetadata(pass, n, doc, obj).Get(); ok {
			common.MarkMetadata(pass, obj, mdFact)
		}
	})
	return common.NewExtractorResult(pass), nil
}

func extractMetadata(pass *analysis.Pass, node ast.Node, doc *ast.CommentGroup, obj types.Object) optional.Option[*common.ExtractedMetadata] {
	if doc == nil {
		return optional.None[*common.ExtractedMetadata]()
	}
	directives := common.ParseDirectives(pass, node, doc)
	found := sets.NewSet[string]()
	exported := isExported(directives)
	var declType schema.Decl
	var metadata []schema.Metadata
	for _, dir := range directives {
		var newSchType schema.Decl
		if found.Contains(dir.GetTypeName()) && !canRepeatDirective(dir) {
			common.Errorf(pass, node, `expected exactly one "ftl:%s" directive but found multiple`,
				dir.GetTypeName())
			continue
		}
		found.Add(dir.GetTypeName())

		if !isAnnotatingValidGoNode(dir, node) {
			if _, ok := node.(*ast.FuncDecl); ok {
				common.NoEndColumnErrorf(pass, dir.GetPosition(), "unexpected directive \"ftl:%s\" attached "+
					"for verb, did you mean to use '//ftl:verb export' instead?", dir.GetTypeName())
				continue
			}

			common.NoEndColumnErrorf(pass, dir.GetPosition(), "unexpected directive \"ftl:%s\"",
				dir.GetTypeName())
			continue
		}

		switch dt := dir.(type) {
		case *common.DirectiveIngress:
			newSchType = &schema.Verb{}
			typ := dt.Type
			if typ == "" {
				typ = "http"
			}
			metadata = append(metadata, &schema.MetadataIngress{
				Pos:    common.GoPosToSchemaPos(pass.Fset, dt.GetPosition()),
				Type:   typ,
				Method: dt.Method,
				Path:   dt.Path,
			})
		case *common.DirectiveCronJob:
			newSchType = &schema.Verb{}
			if exported {
				common.NoEndColumnErrorf(pass, dt.GetPosition(), "ftl:cron cannot be attached to exported verbs")
				continue
			}
			metadata = append(metadata, &schema.MetadataCronJob{
				Pos:  common.GoPosToSchemaPos(pass.Fset, dt.Pos),
				Cron: dt.Cron.String(),
			})
		case *common.DirectiveRetry:
			pos := common.GoPosToSchemaPos(pass.Fset, dt.Pos)
			var catch *schema.Ref
			if dt.Catch != nil {
				catch = &schema.Ref{
					Module: dt.Catch.Module,
					Name:   dt.Catch.Name,
					Pos:    posFromPosWithinDirective(dt.Catch.Pos, pos),
				}
			}
			metadata = append(metadata, &schema.MetadataRetry{
				Pos:        pos,
				Count:      dt.Count,
				MinBackoff: dt.MinBackoff,
				MaxBackoff: dt.MaxBackoff,
				Catch:      catch,
			})
		case *common.DirectiveTopic:
			newSchType = &schema.Topic{}
			metadata = append(metadata, &schema.MetadataPartitions{
				Pos:        common.GoPosToSchemaPos(pass.Fset, dt.Pos),
				Partitions: dt.Partitions,
			})
		case *common.DirectiveSubscriber:
			newSchType = &schema.Verb{}
			pos := common.GoPosToSchemaPos(pass.Fset, dt.Pos)
			metadata = append(metadata, &schema.MetadataSubscriber{
				Pos: pos,
				Topic: &schema.Ref{
					Module: dt.Topic.Module,
					Name:   dt.Topic.Name,
					Pos:    posFromPosWithinDirective(dt.Topic.Pos, pos),
				},
				FromOffset: *dt.FromOffset,
				DeadLetter: dt.DeadLetter,
			})
		case *common.DirectiveTypeMap:
			newSchType = &schema.TypeAlias{}
			metadata = append(metadata, &schema.MetadataTypeMap{
				Pos:        common.GoPosToSchemaPos(pass.Fset, dt.GetPosition()),
				Runtime:    dt.Runtime,
				NativeName: dt.NativeName,
			})
		case *common.DirectiveEncoding:
			metadata = append(metadata, &schema.MetadataEncoding{
				Pos:     common.GoPosToSchemaPos(pass.Fset, dt.GetPosition()),
				Lenient: dt.Lenient,
			})
		case *common.DirectiveVerb:
			newSchType = &schema.Verb{}
		case *common.DirectiveData:
			newSchType = &schema.Data{}
		case *common.DirectiveEnum:
			requireOnlyDirective(pass, node, directives, dt.GetTypeName())
			newSchType = &schema.Enum{}
		case *common.DirectiveExport:
			requireOnlyDirective(pass, node, directives, dt.GetTypeName())
		case *common.DirectiveTypeAlias:
			newSchType = &schema.TypeAlias{}
		case *common.DirectiveDatabase:
			common.MarkDatabaseConfig(pass, obj, dt.Name, dt.Engine)
		}
		declType = updateDeclType(pass, node.Pos(), declType, newSchType)
	}

	md := &common.ExtractedMetadata{
		Type:       declType,
		Metadata:   metadata,
		IsExported: exported,
		Comments:   common.ExtractComments(doc),
	}
	validateMetadata(pass, node, md)
	return optional.Some(md)
}

func posFromPosWithinDirective(pos schema.Position, parentPos schema.Position) schema.Position {
	out := ireflect.DeepCopy(parentPos)
	out.Column += pos.Column - 1
	fmt.Printf("pos: %v\ntopicPos: %v\nfinal: %v\n", pos, parentPos, out)
	return out
}

func validateMetadata(pass *analysis.Pass, node ast.Node, extracted *common.ExtractedMetadata) {
	if _, ok := extracted.Type.(*schema.Verb); !ok {
		return
	}

	isIngress := false
	customEncoding := false
	for _, md := range extracted.Metadata {
		if _, ok := md.(*schema.MetadataIngress); ok {
			isIngress = true
		}

		if _, ok := md.(*schema.MetadataEncoding); ok {
			customEncoding = true
		}
	}

	if customEncoding && !isIngress {
		common.Errorf(pass, node, "custom encoding options can only be specified on ingress verbs")
	}
}

func requireOnlyDirective(pass *analysis.Pass, node ast.Node, directives []common.Directive, typeName string) {
	if len(directives) > 1 {
		common.Errorf(pass, node, "only one directive expected when directive \"ftl:%s\" is present, "+
			"found multiple", typeName)
	}
}

func updateDeclType(pass *analysis.Pass, pos token.Pos, a schema.Decl, b schema.Decl) schema.Decl {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		common.NoEndColumnErrorf(pass, pos, "schema declaration contains conflicting directives")
	}
	return b
}

func isExported(directives []common.Directive) bool {
	for _, d := range directives {
		if exportable, ok := d.(common.Exportable); ok {
			return exportable.IsExported()
		}
	}
	return false
}

func isAnnotatingValidGoNode(dir common.Directive, node ast.Node) bool {
	for _, n := range dir.MustAnnotate() {
		if reflect.TypeOf(n) == reflect.TypeOf(node) {
			return true
		}
	}
	return false
}

func canRepeatDirective(dir common.Directive) bool {
	_, ok := dir.(*common.DirectiveTypeMap)
	return ok
}

// TODO: fix - this doesn't work for member functions.
//
// func getDuplicate(pass *analysis.Pass, name string, newMd *common.ExtractedMetadata) optional.Option[types.Object] {
// 	for obj, md := range common.GetAllFactsOfType[*common.ExtractedMetadata](pass) {
// 		if reflect.TypeOf(md.Type) == reflect.TypeOf(newMd.Type) && obj.Ref() == name {
// 			return optional.Some(obj)
// 		}
// 	}
// 	return optional.None[types.Object]()
// }
