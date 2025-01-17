// Code generated by FTL. DO NOT EDIT.
package parent

import (
	"context"
	ftlchild "ftl/parent/child"
	"github.com/block/ftl/go-runtime/ftl/reflection"
)

type ChildVerbClient func(context.Context) (ftlchild.Resp, error)

type VerbClient func(context.Context) (ftlchild.ChildStruct, error)

func init() {
	reflection.Register(
		reflection.SumType[ftlchild.ChildTypeEnum](
			*new(ftlchild.List),
			*new(ftlchild.Scalar),
		),
		reflection.ProvideResourcesForVerb(
			ftlchild.ChildVerb,
		),
		reflection.ProvideResourcesForVerb(
			Verb,
		),
	)
}
