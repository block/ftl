// Code generated by go2proto. DO NOT EDIT.

package testdata

import "fmt"
import "encoding"
import destpb "github.com/block/ftl/cmd/go2proto/testdata/testdatapb"
import "google.golang.org/protobuf/types/known/timestamppb"
import "google.golang.org/protobuf/types/known/durationpb"
import "github.com/alecthomas/errors"
import "github.com/alecthomas/types/optional"
import "github.com/alecthomas/types/result"

import "time"

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

func sliceMapR[T any, U any](values []T, f func(T) result.Result[U]) result.Result[[]U] {
	if len(values) == 0 {
		return result.Ok[[]U](nil)
	}
	out := make([]U, len(values))
	for i, v := range values {
		r := f(v)
		if r.Err() != nil {
			return result.Err[[]U](r.Err())
		}
		out[i], _ = r.Get()
	}
	return result.Ok[[]U](out)
}

func mapValues[K comparable, V, U any](m map[K]V, f func(V) U) map[K]U {
	out := make(map[K]U, len(m))
	for k, v := range m {
		out[k] = f(v)
	}
	return out
}

func mapValuesR[K comparable, V, U any](m map[K]V, f func(V) result.Result[U]) result.Result[map[K]U] {
	out := make(map[K]U, len(m))
	for k, v := range m {
		r := f(v)
		if r.Err() != nil {
			return result.Err[map[K]U](r.Err())
		}
		val, _ := r.Get()
		out[k] = val
	}
	return result.Ok[map[K]U](out)
}

func orZero[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func orZeroR[T any](v result.Result[*T]) result.Result[T] {
	if v.Err() != nil {
		return result.Err[T](v.Err())
	}
	r, _ := v.Get()
	return result.Ok[T](orZero(r))
}

func ptr[T any](o T) *T {
	return &o
}

func ptrR[T any](o result.Result[T]) result.Result[*T] {
	if o.Err() != nil {
		return result.Err[*T](o.Err())
	}
	r, _ := o.Get()
	return result.Ok[*T](ptr(r))
}

func fromPtr[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func fromPtrR[T any](v result.Result[*T]) result.Result[T] {
	if v.Err() != nil {
		return result.Err[T](v.Err())
	}
	r, _ := v.Get()
	return result.Ok[T](fromPtr(r))
}

func optionalR[T any](r result.Result[*T]) result.Result[optional.Option[T]] {
	if r.Err() != nil {
		return result.Err[optional.Option[T]](r.Err())
	}
	v, _ := r.Get()
	return result.Ok[optional.Option[T]](optional.Ptr(v))
}

func optionalRPtr[T any](r result.Result[*T]) result.Result[optional.Option[*T]] {
	if r.Err() != nil {
		return result.Err[optional.Option[*T]](r.Err())
	}
	v, _ := r.Get()
	return result.Ok[optional.Option[*T]](optional.Zero(v))
}

func setNil[T, O any](v *T, o *O) *T {
	if o == nil {
		return nil
	}
	return v
}

func setNilR[T, O any](v result.Result[*T], o *O) result.Result[*T] {
	if v.Err() != nil {
		return v
	}
	r, _ := v.Get()
	return result.Ok[*T](setNil(r, o))
}

type binaryUnmarshallable[T any] interface {
	*T
	encoding.BinaryUnmarshaler
}

type textUnmarshallable[T any] interface {
	*T
	encoding.TextUnmarshaler
}

func unmarshallBinary[T any, TPtr binaryUnmarshallable[T]](v []byte, f TPtr) result.Result[*T] {
	var to T
	toptr := (TPtr)(&to)

	err := toptr.UnmarshalBinary(v)
	if err != nil {
		return result.Err[*T](err)
	}
	return result.Ok[*T](&to)
}

func unmarshallText[T any, TPtr textUnmarshallable[T]](v []byte, f TPtr) result.Result[*T] {
	var to T
	toptr := (TPtr)(&to)

	err := toptr.UnmarshalText(v)
	if err != nil {
		return result.Err[*T](err)
	}
	return result.Ok[*T](&to)
}

func (x Enum) ToProto() destpb.Enum {
	return destpb.Enum(x)
}

func EnumFromProto(v destpb.Enum) (Enum, error) {
	// TODO: Check if the value is valid.
	return Enum(v), nil
}

func (x *Message) ToProto() *destpb.Message {
	if x == nil {
		return nil
	}
	return &destpb.Message{
		Time:           timestamppb.New(x.Time),
		OptTime:        timestamppb.New(x.OptTime),
		Duration:       durationpb.New(x.Duration),
		Invalid:        orZero(ptr(bool(x.Invalid))),
		Nested:         x.Nested.ToProto(),
		RepeatedNested: sliceMap(x.RepeatedNested, func(v Nested) *destpb.Nested { return v.ToProto() }),
	}
}

func MessageFromProto(v *destpb.Message) (out *Message, err error) {
	if v == nil {
		return nil, nil
	}

	out = &Message{}
	if out.Time, err = orZeroR(result.From(setNil(ptr(v.Time.AsTime()), v.Time), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "Time")
	}
	if out.OptTime, err = orZeroR(result.From(setNil(ptr(v.OptTime.AsTime()), v.OptTime), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "OptTime")
	}
	if out.Duration, err = orZeroR(result.From(setNil(ptr(v.Duration.AsDuration()), v.Duration), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "Duration")
	}
	if out.Invalid, err = orZeroR(result.From(ptr(bool(v.Invalid)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "Invalid")
	}
	if out.Nested, err = orZeroR(result.From(NestedFromProto(v.Nested))).Result(); err != nil {
		return nil, errors.Wrap(err, "Nested")
	}
	if out.RepeatedNested, err = sliceMapR(v.RepeatedNested, func(v *destpb.Nested) result.Result[Nested] { return orZeroR(result.From(NestedFromProto(v))) }).Result(); err != nil {
		return nil, errors.Wrap(err, "RepeatedNested")
	}
	if err := out.Validate(); err != nil {
		return nil, err
	}
	return out, nil
}

func (x *Nested) ToProto() *destpb.Nested {
	if x == nil {
		return nil
	}
	return &destpb.Nested{
		Nested: orZero(ptr(string(x.Nested))),
	}
}

func NestedFromProto(v *destpb.Nested) (out *Nested, err error) {
	if v == nil {
		return nil, nil
	}

	out = &Nested{}
	if out.Nested, err = orZeroR(result.From(ptr(string(v.Nested)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "Nested")
	}
	return out, nil
}

func (x *Root) ToProto() *destpb.Root {
	if x == nil {
		return nil
	}
	return &destpb.Root{
		Int:                orZero(ptr(int64(x.Int))),
		String_:            orZero(ptr(string(x.String))),
		MessagePtr:         x.MessagePtr.ToProto(),
		Enum:               orZero(ptr(x.Enum.ToProto())),
		SumType:            SumTypeToProto(x.SumType),
		OptionalInt:        ptr(int64(x.OptionalInt)),
		OptionalIntPtr:     setNil(ptr(int64(orZero(x.OptionalIntPtr))), x.OptionalIntPtr),
		OptionalMsg:        x.OptionalMsg.ToProto(),
		RepeatedInt:        sliceMap(x.RepeatedInt, func(v int) int64 { return orZero(ptr(int64(v))) }),
		RepeatedMsg:        sliceMap(x.RepeatedMsg, func(v *Message) *destpb.Message { return v.ToProto() }),
		Url:                orZero(ptr(protoMust(x.URL.MarshalBinary()))),
		OptionalWrapper:    setNil(ptr(string(orZero(x.OptionalWrapper.Ptr()))), x.OptionalWrapper.Ptr()),
		ExternalRoot:       orZero(ptr(string(protoMust(x.ExternalRoot.MarshalText())))),
		Key:                orZero(ptr(string(protoMust(x.Key.MarshalText())))),
		OptionalTime:       setNil(timestamppb.New(orZero(x.OptionalTime.Ptr())), x.OptionalTime.Ptr()),
		OptionalMessage:    x.OptionalMessage.Ptr().ToProto(),
		OptionalMessagePtr: x.OptionalMessagePtr.Default(nil).ToProto(),
		Map:                mapValues(x.Map, func(x time.Time) *timestamppb.Timestamp { return timestamppb.New(x) }),
	}
}

func RootFromProto(v *destpb.Root) (out *Root, err error) {
	if v == nil {
		return nil, nil
	}

	out = &Root{}
	if out.Int, err = orZeroR(result.From(ptr(int(v.Int)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "Int")
	}
	if out.String, err = orZeroR(result.From(ptr(string(v.String_)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "String")
	}
	if out.MessagePtr, err = result.From(MessageFromProto(v.MessagePtr)).Result(); err != nil {
		return nil, errors.Wrap(err, "MessagePtr")
	}
	if out.Enum, err = orZeroR(ptrR(result.From(EnumFromProto(v.Enum)))).Result(); err != nil {
		return nil, errors.Wrap(err, "Enum")
	}
	if out.SumType, err = orZeroR(ptrR(result.From(SumTypeFromProto(v.SumType)))).Result(); err != nil {
		return nil, errors.Wrap(err, "SumType")
	}
	if out.OptionalInt, err = orZeroR(result.From(setNil(ptr(int(orZero(v.OptionalInt))), v.OptionalInt), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalInt")
	}
	if out.OptionalIntPtr, err = result.From(setNil(ptr(int(orZero(v.OptionalIntPtr))), v.OptionalIntPtr), nil).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalIntPtr")
	}
	if out.OptionalMsg, err = result.From(MessageFromProto(v.OptionalMsg)).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalMsg")
	}
	if out.RepeatedInt, err = sliceMapR(v.RepeatedInt, func(v int64) result.Result[int] { return orZeroR(result.From(ptr(int(v)), nil)) }).Result(); err != nil {
		return nil, errors.Wrap(err, "RepeatedInt")
	}
	if out.RepeatedMsg, err = sliceMapR(v.RepeatedMsg, func(v *destpb.Message) result.Result[*Message] { return result.From(MessageFromProto(v)) }).Result(); err != nil {
		return nil, errors.Wrap(err, "RepeatedMsg")
	}
	if out.URL, err = unmarshallBinary(v.Url, out.URL).Result(); err != nil {
		return nil, errors.Wrap(err, "URL")
	}
	if out.OptionalWrapper, err = optionalR(result.From(setNil(ptr(string(orZero(v.OptionalWrapper))), v.OptionalWrapper), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalWrapper")
	}
	if out.ExternalRoot, err = orZeroR(unmarshallText([]byte(v.ExternalRoot), &out.ExternalRoot)).Result(); err != nil {
		return nil, errors.Wrap(err, "ExternalRoot")
	}
	if out.Key, err = orZeroR(unmarshallText([]byte(v.Key), &out.Key)).Result(); err != nil {
		return nil, errors.Wrap(err, "Key")
	}
	if out.OptionalTime, err = optionalR(result.From(setNil(ptr(v.OptionalTime.AsTime()), v.OptionalTime), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalTime")
	}
	if out.OptionalMessage, err = optionalR(result.From(MessageFromProto(v.OptionalMessage))).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalMessage")
	}
	if out.OptionalMessagePtr, err = optionalRPtr(result.From(MessageFromProto(v.OptionalMessagePtr))).Result(); err != nil {
		return nil, errors.Wrap(err, "OptionalMessagePtr")
	}
	if out.Map, err = mapValuesR(v.Map, func(v *timestamppb.Timestamp) result.Result[time.Time] {
		return orZeroR(result.From(setNil(ptr(v.AsTime()), v), nil))
	}).Result(); err != nil {
		return nil, errors.Wrap(err, "Map")
	}
	return out, nil
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

func SubSumTypeFromProto(v *destpb.SubSumType) (SubSumType, error) {
	if v == nil {
		return nil, nil
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
		A: orZero(ptr(string(x.A))),
	}
}

func SubSumTypeAFromProto(v *destpb.SubSumTypeA) (out *SubSumTypeA, err error) {
	if v == nil {
		return nil, nil
	}

	out = &SubSumTypeA{}
	if out.A, err = orZeroR(result.From(ptr(string(v.A)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "A")
	}
	return out, nil
}

func (x *SubSumTypeB) ToProto() *destpb.SubSumTypeB {
	if x == nil {
		return nil
	}
	return &destpb.SubSumTypeB{
		A: orZero(ptr(string(x.A))),
	}
}

func SubSumTypeBFromProto(v *destpb.SubSumTypeB) (out *SubSumTypeB, err error) {
	if v == nil {
		return nil, nil
	}

	out = &SubSumTypeB{}
	if out.A, err = orZeroR(result.From(ptr(string(v.A)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "A")
	}
	return out, nil
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
			Value: &destpb.SumType_SumTypeA{value.ToProto()},
		}
	case *SumTypeB:
		return &destpb.SumType{
			Value: &destpb.SumType_SumTypeB{value.ToProto()},
		}
	case *SumTypeC:
		return &destpb.SumType{
			Value: &destpb.SumType_SumTypeC{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func SumTypeFromProto(v *destpb.SumType) (SumType, error) {
	if v == nil {
		return nil, nil
	}
	switch v.Value.(type) {
	case *destpb.SumType_SubSumTypeA:
		return SubSumTypeAFromProto(v.GetSubSumTypeA())
	case *destpb.SumType_SubSumTypeB:
		return SubSumTypeBFromProto(v.GetSubSumTypeB())
	case *destpb.SumType_SumTypeA:
		return SumTypeAFromProto(v.GetSumTypeA())
	case *destpb.SumType_SumTypeB:
		return SumTypeBFromProto(v.GetSumTypeB())
	case *destpb.SumType_SumTypeC:
		return SumTypeCFromProto(v.GetSumTypeC())
	default:
		panic(fmt.Sprintf("unknown variant: %T", v.Value))
	}
}

func (x *SumTypeA) ToProto() *destpb.SumTypeA {
	if x == nil {
		return nil
	}
	return &destpb.SumTypeA{
		A: orZero(ptr(string(x.A))),
	}
}

func SumTypeAFromProto(v *destpb.SumTypeA) (out *SumTypeA, err error) {
	if v == nil {
		return nil, nil
	}

	out = &SumTypeA{}
	if out.A, err = orZeroR(result.From(ptr(string(v.A)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "A")
	}
	return out, nil
}

func (x *SumTypeB) ToProto() *destpb.SumTypeB {
	if x == nil {
		return nil
	}
	return &destpb.SumTypeB{
		B: orZero(ptr(int64(x.B))),
	}
}

func SumTypeBFromProto(v *destpb.SumTypeB) (out *SumTypeB, err error) {
	if v == nil {
		return nil, nil
	}

	out = &SumTypeB{}
	if out.B, err = orZeroR(result.From(ptr(int(v.B)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "B")
	}
	return out, nil
}

func (x *SumTypeC) ToProto() *destpb.SumTypeC {
	if x == nil {
		return nil
	}
	return &destpb.SumTypeC{
		C: orZero(ptr(float64(x.C))),
	}
}

func SumTypeCFromProto(v *destpb.SumTypeC) (out *SumTypeC, err error) {
	if v == nil {
		return nil, nil
	}

	out = &SumTypeC{}
	if out.C, err = orZeroR(result.From(ptr(float64(v.C)), nil)).Result(); err != nil {
		return nil, errors.Wrap(err, "C")
	}
	return out, nil
}
