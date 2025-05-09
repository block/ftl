// Code generated by FTL. DO NOT EDIT.
package other

import (
	"context"
	"github.com/block/ftl/common/reflection"
	lib "github.com/block/ftl/go-runtime/schema/testdata"
)

type EchoClient func(context.Context, EchoRequest) (EchoResponse, error)

func init() {
	reflection.Register(
		reflection.SumType[SecondTypeEnum](
			*new(A),
			*new(B),
		),
		reflection.SumType[TypeEnum](
			*new(MyBool),
			*new(MyBytes),
			*new(MyFloat),
			*new(MyInt),
			*new(MyList),
			*new(MyMap),
			*new(MyOption),
			*new(MyString),
			*new(MyStruct),
			*new(MyTime),
			*new(MyUnit),
		),
		reflection.ExternalType(*new(lib.NonFTLType)),
		reflection.ExternalType(*new(lib.AnotherNonFTLType)),
		reflection.ProvideResourcesForVerb(
			Echo,
		),
	)
}
