// Code generated by FTL. DO NOT EDIT.
package mysql

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/server"
)

type InsertClient func(context.Context, InsertRequest) (InsertResponse, error)

type QueryClient func(context.Context) ([]string, error)

func init() {
	reflection.Register(
		reflection.Database[MyDbConfig]("testdb", server.InitMySQL),
		reflection.ProvideResourcesForVerb(
			Insert,
			server.DatabaseHandle[MyDbConfig]("mysql"),
		),
		reflection.ProvideResourcesForVerb(
			Query,
			server.DatabaseHandle[MyDbConfig]("mysql"),
		),
	)
}
