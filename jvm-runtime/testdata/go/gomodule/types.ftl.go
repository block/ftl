// Code generated by FTL. DO NOT EDIT.
package gomodule

import (
	"context"
	ftlbuiltin "ftl/builtin"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/decentralized-identity/web5-go/dids/did"
	stdtime "time"
)

type BoolVerbClient func(context.Context, bool) (bool, error)

type BytesVerbClient func(context.Context, []byte) ([]byte, error)

type EmptyVerbClient func(context.Context) error

type ErrorEmptyVerbClient func(context.Context) error

type ExternalTypeVerbClient func(context.Context, DID) (DID, error)

type FloatVerbClient func(context.Context, float64) (float64, error)

type IntVerbClient func(context.Context, int) (int, error)

type ObjectArrayVerbClient func(context.Context, []TestObject) ([]TestObject, error)

type ObjectMapVerbClient func(context.Context, map[string]TestObject) (map[string]TestObject, error)

type OptionalBoolVerbClient func(context.Context, ftl.Option[bool]) (ftl.Option[bool], error)

type OptionalBytesVerbClient func(context.Context, ftl.Option[[]byte]) (ftl.Option[[]byte], error)

type OptionalFloatVerbClient func(context.Context, ftl.Option[float64]) (ftl.Option[float64], error)

type OptionalIntVerbClient func(context.Context, ftl.Option[int]) (ftl.Option[int], error)

type OptionalStringArrayVerbClient func(context.Context, ftl.Option[[]string]) (ftl.Option[[]string], error)

type OptionalStringMapVerbClient func(context.Context, ftl.Option[map[string]string]) (ftl.Option[map[string]string], error)

type OptionalStringVerbClient func(context.Context, ftl.Option[string]) (ftl.Option[string], error)

type OptionalTestObjectOptionalFieldsVerbClient func(context.Context, ftl.Option[TestObjectOptionalFields]) (ftl.Option[TestObjectOptionalFields], error)

type OptionalTestObjectVerbClient func(context.Context, ftl.Option[TestObject]) (ftl.Option[TestObject], error)

type OptionalTimeVerbClient func(context.Context, ftl.Option[stdtime.Time]) (ftl.Option[stdtime.Time], error)

type ParameterizedObjectVerbClient func(context.Context, ParameterizedType[string]) (ParameterizedType[string], error)

type PrimitiveParamVerbClient func(context.Context, int) (string, error)

type PrimitiveResponseVerbClient func(context.Context, string) (int, error)

type SinkVerbClient func(context.Context, string) error

type SourceVerbClient func(context.Context) (string, error)

type StringArrayVerbClient func(context.Context, []string) ([]string, error)

type StringEnumVerbClient func(context.Context, ShapeWrapper) (ShapeWrapper, error)

type StringMapVerbClient func(context.Context, map[string]string) (map[string]string, error)

type StringVerbClient func(context.Context, string) (string, error)

type TestGenericTypeClient func(context.Context, ftlbuiltin.FailedEvent[TestObject]) (ftlbuiltin.FailedEvent[TestObject], error)

type TestObjectOptionalFieldsVerbClient func(context.Context, TestObjectOptionalFields) (TestObjectOptionalFields, error)

type TestObjectVerbClient func(context.Context, TestObject) (TestObject, error)

type TimeVerbClient func(context.Context, stdtime.Time) (stdtime.Time, error)

type TypeEnumVerbClient func(context.Context, AnimalWrapper) (AnimalWrapper, error)

type TypeWrapperEnumVerbClient func(context.Context, TypeEnumWrapper) (TypeEnumWrapper, error)

type ValueEnumVerbClient func(context.Context, ColorWrapper) (ColorWrapper, error)

func init() {
	reflection.Register(
		reflection.SumType[Animal](
			*new(Cat),
			*new(Dog),
		),
		reflection.SumType[TypeWrapperEnum](
			*new(Scalar),
			*new(StringList),
		),
		reflection.ExternalType(*new(did.DID)),

		reflection.ProvideResourcesForVerb(
			BoolVerb,
		),

		reflection.ProvideResourcesForVerb(
			BytesVerb,
		),

		reflection.ProvideResourcesForVerb(
			EmptyVerb,
		),

		reflection.ProvideResourcesForVerb(
			ErrorEmptyVerb,
		),

		reflection.ProvideResourcesForVerb(
			ExternalTypeVerb,
		),

		reflection.ProvideResourcesForVerb(
			FloatVerb,
		),

		reflection.ProvideResourcesForVerb(
			IntVerb,
		),

		reflection.ProvideResourcesForVerb(
			ObjectArrayVerb,
		),

		reflection.ProvideResourcesForVerb(
			ObjectMapVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalBoolVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalBytesVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalFloatVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalIntVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalStringArrayVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalStringMapVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalStringVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalTestObjectOptionalFieldsVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalTestObjectVerb,
		),

		reflection.ProvideResourcesForVerb(
			OptionalTimeVerb,
		),

		reflection.ProvideResourcesForVerb(
			ParameterizedObjectVerb,
		),

		reflection.ProvideResourcesForVerb(
			PrimitiveParamVerb,
		),

		reflection.ProvideResourcesForVerb(
			PrimitiveResponseVerb,
		),

		reflection.ProvideResourcesForVerb(
			SinkVerb,
		),

		reflection.ProvideResourcesForVerb(
			SourceVerb,
		),

		reflection.ProvideResourcesForVerb(
			StringArrayVerb,
		),

		reflection.ProvideResourcesForVerb(
			StringEnumVerb,
		),

		reflection.ProvideResourcesForVerb(
			StringMapVerb,
		),

		reflection.ProvideResourcesForVerb(
			StringVerb,
		),

		reflection.ProvideResourcesForVerb(
			TestGenericType,
		),

		reflection.ProvideResourcesForVerb(
			TestObjectOptionalFieldsVerb,
		),

		reflection.ProvideResourcesForVerb(
			TestObjectVerb,
		),

		reflection.ProvideResourcesForVerb(
			TimeVerb,
		),

		reflection.ProvideResourcesForVerb(
			TypeEnumVerb,
		),

		reflection.ProvideResourcesForVerb(
			TypeWrapperEnumVerb,
		),

		reflection.ProvideResourcesForVerb(
			ValueEnumVerb,
		),
	)
}
