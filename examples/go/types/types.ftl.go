// Code generated by FTL. DO NOT EDIT.
package types

import (
	"github.com/block/ftl/common/reflection"
)

func init() {
	reflection.Register(
		reflection.SumType[SumType](
			*new(SumTypeOne),
			*new(SumTypeTwo),
		),
	)
}
