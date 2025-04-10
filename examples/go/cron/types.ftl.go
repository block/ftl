// Code generated by FTL. DO NOT EDIT.
package cron

import (
	"context"
	"github.com/block/ftl/common/reflection"
)

type HourlyClient func(context.Context) error

type ThirtySecondsClient func(context.Context) error

func init() {
	reflection.Register(

		reflection.ProvideResourcesForVerb(
			Hourly,
		),

		reflection.ProvideResourcesForVerb(
			ThirtySeconds,
		),
	)
}
