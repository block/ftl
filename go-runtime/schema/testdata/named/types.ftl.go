// Code generated by FTL. DO NOT EDIT.
package named

import (
	"context"
	"github.com/block/ftl/common/reflection"
)

type PingInternalUserClient func(context.Context, InternalUser) error

type PingUserClient func(context.Context, User) error

func init() {
	reflection.Register(

		reflection.ProvideResourcesForVerb(
			PingInternalUser,
		),

		reflection.ProvideResourcesForVerb(
			PingUser,
		),
	)
}
