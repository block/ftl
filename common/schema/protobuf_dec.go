package schema

import (
	"fmt"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

func PosFromProto(pos *schemapb.Position) Position {
	if pos == nil {
		return Position{}
	}
	return Position{
		Line:     int(pos.Line),
		Column:   int(pos.Column),
		Filename: pos.Filename,
	}
}

func declListToSchema(s []*schemapb.Decl) []Decl {
	var out []Decl
	for _, n := range s {
		switch n := n.Value.(type) {
		case *schemapb.Decl_Verb:
			out = append(out, VerbFromProto(n.Verb))
		case *schemapb.Decl_Data:
			out = append(out, DataFromProto(n.Data))
		case *schemapb.Decl_Database:
			out = append(out, DatabaseFromProto(n.Database))
		case *schemapb.Decl_Enum:
			out = append(out, EnumFromProto(n.Enum))
		case *schemapb.Decl_TypeAlias:
			out = append(out, TypeAliasFromProto(n.TypeAlias))
		case *schemapb.Decl_Config:
			out = append(out, ConfigFromProto(n.Config))
		case *schemapb.Decl_Secret:
			out = append(out, SecretFromProto(n.Secret))
		case *schemapb.Decl_Topic:
			out = append(out, TopicFromProto(n.Topic))
		}
	}
	return out
}

func valueToSchema(v *schemapb.Value) Value {
	switch s := v.Value.(type) {
	case *schemapb.Value_IntValue:
		return &IntValue{
			Pos:   PosFromProto(s.IntValue.Pos),
			Value: int(s.IntValue.Value),
		}
	case *schemapb.Value_StringValue:
		return &StringValue{
			Pos:   PosFromProto(s.StringValue.Pos),
			Value: s.StringValue.GetValue(),
		}
	case *schemapb.Value_TypeValue:
		return &TypeValue{
			Pos:   PosFromProto(s.TypeValue.Pos),
			Value: TypeFromProto(s.TypeValue.Value),
		}
	}
	panic(fmt.Sprintf("unhandled schema value: %T", v.Value))
}

func metadataListToSchema(s []*schemapb.Metadata) []Metadata {
	var out []Metadata
	for _, n := range s {
		out = append(out, metadataToSchema(n))
	}
	return out
}

func metadataToSchema(s *schemapb.Metadata) Metadata {
	switch s := s.Value.(type) {
	case *schemapb.Metadata_Calls:
		return &MetadataCalls{
			Pos:   PosFromProto(s.Calls.Pos),
			Calls: refListToSchema(s.Calls.Calls),
		}
	case *schemapb.Metadata_Config:
		return &MetadataConfig{
			Pos:    PosFromProto(s.Config.Pos),
			Config: refListToSchema(s.Config.Config),
		}
	case *schemapb.Metadata_Databases:
		return &MetadataDatabases{
			Pos:   PosFromProto(s.Databases.Pos),
			Calls: refListToSchema(s.Databases.Calls),
		}

	case *schemapb.Metadata_Ingress:
		return &MetadataIngress{
			Pos:    PosFromProto(s.Ingress.Pos),
			Type:   s.Ingress.Type,
			Method: s.Ingress.Method,
			Path:   ingressPathComponentListToSchema(s.Ingress.Path),
		}

	case *schemapb.Metadata_CronJob:
		return &MetadataCronJob{
			Pos:  PosFromProto(s.CronJob.Pos),
			Cron: s.CronJob.Cron,
		}

	case *schemapb.Metadata_Alias:
		return &MetadataAlias{
			Pos:   PosFromProto(s.Alias.Pos),
			Kind:  AliasKind(s.Alias.Kind),
			Alias: s.Alias.Alias,
		}

	case *schemapb.Metadata_Retry:
		var count *int
		if s.Retry.Count != nil {
			countValue := int(*s.Retry.Count)
			count = &countValue
		}
		var catch *Ref
		if s.Retry.Catch != nil {
			catch = RefFromProto(s.Retry.Catch)
		}
		return &MetadataRetry{
			Pos:        PosFromProto(s.Retry.Pos),
			Count:      count,
			MinBackoff: s.Retry.MinBackoff,
			MaxBackoff: s.Retry.MaxBackoff,
			Catch:      catch,
		}

	case *schemapb.Metadata_Secrets:
		return &MetadataSecrets{
			Pos:     PosFromProto(s.Secrets.Pos),
			Secrets: refListToSchema(s.Secrets.Secrets),
		}

	case *schemapb.Metadata_Subscriber:
		return &MetadataSubscriber{
			Pos:        PosFromProto(s.Subscriber.Pos),
			Topic:      RefFromProto(s.Subscriber.Topic),
			FromOffset: FromOffset(s.Subscriber.FromOffset),
			DeadLetter: s.Subscriber.DeadLetter,
		}

	case *schemapb.Metadata_TypeMap:
		return &MetadataTypeMap{
			Pos:        PosFromProto(s.TypeMap.Pos),
			Runtime:    s.TypeMap.Runtime,
			NativeName: s.TypeMap.NativeName,
		}

	case *schemapb.Metadata_Encoding:
		return &MetadataEncoding{
			Pos:     PosFromProto(s.Encoding.Pos),
			Lenient: s.Encoding.Lenient,
		}

	case *schemapb.Metadata_Publisher:
		return &MetadataPublisher{
			Pos:    PosFromProto(s.Publisher.Pos),
			Topics: refListToSchema(s.Publisher.Topics),
		}

	case *schemapb.Metadata_SqlMigration:
		return &MetadataSQLMigration{
			Pos:    PosFromProto(s.SqlMigration.Pos),
			Digest: s.SqlMigration.Digest,
		}

	case *schemapb.Metadata_Artefact:
		return &MetadataArtefact{
			Pos:        PosFromProto(s.Artefact.Pos),
			Executable: s.Artefact.Executable,
			Path:       s.Artefact.Path,
			Digest:     s.Artefact.Digest,
		}
	case *schemapb.Metadata_Partitions:
		return &MetadataPartitions{
			Pos:        PosFromProto(s.Partitions.Pos),
			Partitions: int(s.Partitions.Partitions),
		}
	case *schemapb.Metadata_SqlQuery:
		return &MetadataSQLQuery{
			Pos:   PosFromProto(s.SqlQuery.Pos),
			Query: s.SqlQuery.Query,
		}
	case *schemapb.Metadata_DbColumn:
		return &MetadataDBColumn{
			Pos:   PosFromProto(s.DbColumn.Pos),
			Table: s.DbColumn.Table,
			Name:  s.DbColumn.Name,
		}

	default:
		panic(fmt.Sprintf("unhandled metadata type: %T", s))
	}
}
