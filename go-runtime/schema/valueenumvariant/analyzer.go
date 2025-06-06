package valueenumvariant

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl-golang-tools/go/analysis/passes/inspect"
	"github.com/block/ftl-golang-tools/go/ast/inspector"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/go-runtime/schema/common"
)

// Extractor extracts possible value enum variants.
//
// All named constants are marked as possible enum variants and subsequent extractors determine if they are part of an
// enum.
var Extractor = common.NewExtractor("valueenumvariant", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.GenDecl) //nolint:forcetypeassert
		if node.Tok != token.CONST {
			return
		}

		var typ ast.Expr
		for i, s := range node.Specs {
			v, ok := s.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// In an iota enum, only the first value has a type.
			// Hydrate this to subsequent values so we can associate them with the enum.
			if i == 0 && isIotaEnum(v) {
				typ = v.Type
			} else if v.Type == nil {
				v.Type = typ
			}
			extractEnumVariant(pass, v)
		}
	})
	return common.NewExtractorResult(pass), nil
}

func extractEnumVariant(pass *analysis.Pass, node *ast.ValueSpec) {
	_, ok := node.Type.(*ast.Ident)
	if !ok {
		return
	}
	c, ok := pass.TypesInfo.Defs[node.Names[0]].(*types.Const)
	if !ok {
		return
	}
	value, ok := extractValue(pass, c).Get()
	if !ok {
		return
	}

	obj, ok := common.GetObjectForNode(pass.TypesInfo, node).Get()
	if !ok {
		return
	}
	name := c.Id()
	nameComponents := strings.Split(name, ".")
	if !unicode.IsUpper(rune(nameComponents[len(nameComponents)-1][0])) {
		// First letter must be a capital
		common.Errorf(pass, node, "Enum value must be exported by making the first letter capitalized")
		return
	}
	name = strcase.ToUpperCamel(name)
	variant := &schema.EnumVariant{
		Pos:   common.GoPosToSchemaPos(pass, c.Pos()),
		Name:  name,
		Value: value,
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		variant.Comments = md.Comments
	}
	if typ, ok := common.GetObjectForNode(pass.TypesInfo, node.Type).Get(); ok {
		common.MarkMaybeValueEnumVariant(pass, obj, variant, typ)
	}
}

func extractValue(pass *analysis.Pass, cnode *types.Const) optional.Option[schema.Value] {
	if b, ok := cnode.Type().Underlying().(*types.Basic); ok {
		switch b.Kind() {
		case types.String:
			value, err := strconv.Unquote(cnode.Val().String())
			if err != nil {
				return optional.None[schema.Value]()
			}
			return optional.Some[schema.Value](&schema.StringValue{
				Pos:   common.GoPosToSchemaPos(pass, cnode.Pos()),
				Value: value,
			})

		case types.Int:
			value, err := strconv.ParseInt(cnode.Val().String(), 10, 64)
			if err != nil {
				return optional.None[schema.Value]()
			}
			return optional.Some[schema.Value](&schema.IntValue{
				Pos:   common.GoPosToSchemaPos(pass, cnode.Pos()),
				Value: int(value),
			})

		default:
			return optional.None[schema.Value]()
		}
	}
	return optional.None[schema.Value]()
}

func isIotaEnum(node ast.Node) bool {
	switch t := node.(type) {
	case *ast.ValueSpec:
		if len(t.Values) != 1 {
			return false
		}
		return isIotaEnum(t.Values[0])
	case *ast.Ident:
		return t.Name == "iota"
	case *ast.BinaryExpr:
		return isIotaEnum(t.X) || isIotaEnum(t.Y)
	default:
		return false
	}
}
