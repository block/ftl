package topic

import (
	"go/ast"
	"go/types"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/go-runtime/schema/common"
)

// Extractor extracts topics.
var Extractor = common.NewResourceDeclExtractor[*schema.Topic]("topic", Extract, matchFunc)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.TypeSpec) optional.Option[*schema.Topic] {
	idxListExpr, ok := node.Type.(*ast.IndexListExpr)
	if !ok || len(idxListExpr.Indices) != 2 {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}

	eventType, ok := common.ExtractType(pass, idxListExpr.Indices[0]).Get()
	if !ok {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}

	// extract name
	name := strcase.ToLowerCamel(node.Name.Name)
	if !schema.ValidateName(name) {
		common.Errorf(pass, node, "topic names must be valid identifiers")
		return optional.None[*schema.Topic]()
	}

	// topic fact
	topic := &schema.Topic{
		Pos:   common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name:  name,
		Event: eventType,
	}
	partitions := 1
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		topic.Comments = md.Comments
		topic.Visibility = schema.Visibility(md.Visibility)

		for _, c := range md.Metadata {
			if pm, ok := c.(*schema.MetadataPartitions); ok {
				topic.Metadata = append(topic.Metadata, pm)
				partitions = pm.Partitions
			}
		}
	}

	// mapper fact
	mapperObj, ok := common.GetObjectForNode(pass.TypesInfo, idxListExpr.Indices[1]).Get()
	if !ok {
		common.Errorf(pass, node, "could not find type for topic partition mapper")
		return optional.None[*schema.Topic]()
	}
	if mapperObj.Type().String() == common.FtlSinglePartitionMapPath && partitions > 1 {
		common.Errorf(pass, node, "single partition map can not have more than one partition")
	}

	var associatedMapperObj optional.Option[types.Object]
	if mapper, ok := idxListExpr.Indices[1].(*ast.IndexExpr); ok {
		associatedObj, ok := common.GetObjectForNode(pass.TypesInfo, mapper.Index).Get()
		if !ok {
			common.Errorf(pass, node, "could not find associated type for topic partition mapper")
			return optional.None[*schema.Topic]()
		}
		associatedMapperObj = optional.Some(associatedObj)
	}
	common.MarkTopicMapper(pass, mapperObj, associatedMapperObj, topic)
	return optional.Some(topic)
}

func matchFunc(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	return common.GetVerbResourceType(pass, obj) == common.VerbResourceTypeTopicHandle
}
