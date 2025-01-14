// Code generated by go2proto. DO NOT EDIT.

package testdata

import "fmt"
import destpb "github.com/block/ftl/cmd/go2proto/testdata/testdatapb"
import "google.golang.org/protobuf/proto"
import "google.golang.org/protobuf/types/known/timestamppb"
import "google.golang.org/protobuf/types/known/durationpb"

import "github.com/block/ftl/internal/key"
import "net/url"

var _ fmt.Stringer
var _ = timestamppb.Timestamp{}
var _ = durationpb.Duration{}

// protoSlice converts a slice of values to a slice of protobuf values.
func protoSlice[P any, T interface{ ToProto() P }](values []T) []P {
	out := make([]P, len(values))
	for i, v := range values {
		out[i] = v.ToProto()
	}
	return out
}

// protoSlicef converts a slice of values to a slice of protobuf values using a mapping function.
func protoSlicef[P, T any](values []T, f func(T) P) []P {
	out := make([]P, len(values))
	for i, v := range values {
		out[i] = f(v)
	}
	return out
}

func protoMust[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func sliceMap[T any, U any](values []T, f func(T) U) []U {
	out := make([]U, len(values))
	for i, v := range values {
		out[i] = f(v)
	}
	return out
}

func orZero[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func ptr[T any, O any](v *O, o T) *T {
	if v == nil {
		return nil
	}
	return &o
}

func fromPtr[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func (x Enum) ToProto() destpb.Enum {
	return destpb.Enum(x)
}

func EnumFromProto(v destpb.Enum) Enum {
	return Enum(v)
}

func (x *Message) ToProto() *destpb.Message {
	if x == nil {
		return nil
	}
	return &destpb.Message{
		Time:     timestamppb.New(x.Time),
		Duration: durationpb.New(x.Duration),
		Nested:   x.Nested.ToProto(),
	}
}

func MessageFromProto(v *destpb.Message) *Message {
	if v == nil {
		return nil
	}

	return &Message{
		Time:     v.Time.AsTime(),
		Duration: v.Duration.AsDuration(),
		Nested:   fromPtr(NestedFromProto(v.Nested)),
	}
}

func (x *Nested) ToProto() *destpb.Nested {
	if x == nil {
		return nil
	}
	return &destpb.Nested{
		Nested: string(x.Nested),
	}
}

func NestedFromProto(v *destpb.Nested) *Nested {
	if v == nil {
		return nil
	}

	return &Nested{
		Nested: string(v.Nested),
	}
}

func (x *Root) ToProto() *destpb.Root {
	if x == nil {
		return nil
	}
	return &destpb.Root{
		Int:            int64(x.Int),
		String_:        string(x.String),
		MessagePtr:     x.MessagePtr.ToProto(),
		Enum:           x.Enum.ToProto(),
		SumType:        SumTypeToProto(x.SumType),
		OptionalInt:    proto.Int64(int64(x.OptionalInt)),
		OptionalIntPtr: proto.Int64(int64(*x.OptionalIntPtr)),
		OptionalMsg:    x.OptionalMsg.ToProto(),
		RepeatedInt:    protoSlicef(x.RepeatedInt, func(v int) int64 { return int64(v) }),
		RepeatedMsg:    protoSlice[*destpb.Message](x.RepeatedMsg),
		Url:            protoMust(x.URL.MarshalBinary()),
		Key:            string(protoMust(x.Key.MarshalText())),
	}
}

func RootFromProto(v *destpb.Root) *Root {
	if v == nil {
		return nil
	}
	f12 := &url.URL{}
	f12.UnmarshalBinary(v.Url)
	f13 := &key.Deployment{}
	f13.UnmarshalText([]byte(v.Key))

	return &Root{
		Int:            int(v.Int),
		String:         string(v.String_),
		MessagePtr:     MessageFromProto(v.MessagePtr),
		Enum:           EnumFromProto(v.Enum),
		SumType:        SumTypeFromProto(v.SumType),
		OptionalInt:    int(orZero(v.OptionalInt)),
		OptionalIntPtr: ptr(v.OptionalIntPtr, int(orZero(v.OptionalIntPtr))),
		OptionalMsg:    MessageFromProto(v.OptionalMsg),
		RepeatedInt:    sliceMap(v.RepeatedInt, func(v int64) int { return int(v) }),
		RepeatedMsg:    sliceMap(v.RepeatedMsg, MessageFromProto),
		URL:            f12,
		Key:            fromPtr(f13),
	}
}

// SubSumTypeToProto converts a SubSumType sum type to a protobuf message.
func SubSumTypeToProto(value SubSumType) *destpb.SubSumType {
	switch value := value.(type) {
	case nil:
		return nil
	case *SubSumTypeA:
		return &destpb.SubSumType{
			Value: &destpb.SubSumType_A{value.ToProto()},
		}
	case *SubSumTypeB:
		return &destpb.SubSumType{
			Value: &destpb.SubSumType_B{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func SubSumTypeFromProto(v *destpb.SubSumType) SubSumType {
	if v == nil {
		return nil
	}
	switch v.Value.(type) {
	case *destpb.SubSumType_A:
		return SubSumTypeAFromProto(v.GetA())
	case *destpb.SubSumType_B:
		return SubSumTypeBFromProto(v.GetB())
	default:
		panic(fmt.Sprintf("unknown variant: %T", v.Value))
	}
}

func (x *SubSumTypeA) ToProto() *destpb.SubSumTypeA {
	if x == nil {
		return nil
	}
	return &destpb.SubSumTypeA{
		A: string(x.A),
	}
}

func SubSumTypeAFromProto(v *destpb.SubSumTypeA) *SubSumTypeA {
	if v == nil {
		return nil
	}

	return &SubSumTypeA{
		A: string(v.A),
	}
}

func (x *SubSumTypeB) ToProto() *destpb.SubSumTypeB {
	if x == nil {
		return nil
	}
	return &destpb.SubSumTypeB{
		A: string(x.A),
	}
}

func SubSumTypeBFromProto(v *destpb.SubSumTypeB) *SubSumTypeB {
	if v == nil {
		return nil
	}

	return &SubSumTypeB{
		A: string(v.A),
	}
}

// SumTypeToProto converts a SumType sum type to a protobuf message.
func SumTypeToProto(value SumType) *destpb.SumType {
	switch value := value.(type) {
	case nil:
		return nil
	case *SubSumTypeA:
		return &destpb.SumType{
			Value: &destpb.SumType_SubSumTypeA{value.ToProto()},
		}
	case *SubSumTypeB:
		return &destpb.SumType{
			Value: &destpb.SumType_SubSumTypeB{value.ToProto()},
		}
	case *SumTypeA:
		return &destpb.SumType{
			Value: &destpb.SumType_A{value.ToProto()},
		}
	case *SumTypeB:
		return &destpb.SumType{
			Value: &destpb.SumType_B{value.ToProto()},
		}
	case *SumTypeC:
		return &destpb.SumType{
			Value: &destpb.SumType_C{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func SumTypeFromProto(v *destpb.SumType) SumType {
	if v == nil {
		return nil
	}
	switch v.Value.(type) {
	case *destpb.SumType_SubSumTypeA:
		return SubSumTypeAFromProto(v.GetSubSumTypeA())
	case *destpb.SumType_SubSumTypeB:
		return SubSumTypeBFromProto(v.GetSubSumTypeB())
	case *destpb.SumType_A:
		return SumTypeAFromProto(v.GetA())
	case *destpb.SumType_B:
		return SumTypeBFromProto(v.GetB())
	case *destpb.SumType_C:
		return SumTypeCFromProto(v.GetC())
	default:
		panic(fmt.Sprintf("unknown variant: %T", v.Value))
	}
}

func (x *SumTypeA) ToProto() *destpb.SumTypeA {
	if x == nil {
		return nil
	}
	return &destpb.SumTypeA{
		A: string(x.A),
	}
}

func SumTypeAFromProto(v *destpb.SumTypeA) *SumTypeA {
	if v == nil {
		return nil
	}

	return &SumTypeA{
		A: string(v.A),
	}
}

func (x *SumTypeB) ToProto() *destpb.SumTypeB {
	if x == nil {
		return nil
	}
	return &destpb.SumTypeB{
		B: int64(x.B),
	}
}

func SumTypeBFromProto(v *destpb.SumTypeB) *SumTypeB {
	if v == nil {
		return nil
	}

	return &SumTypeB{
		B: int(v.B),
	}
}

func (x *SumTypeC) ToProto() *destpb.SumTypeC {
	if x == nil {
		return nil
	}
	return &destpb.SumTypeC{
		C: float64(x.C),
	}
}

func SumTypeCFromProto(v *destpb.SumTypeC) *SumTypeC {
	if v == nil {
		return nil
	}

	return &SumTypeC{
		C: float64(v.C),
	}
}
