// Code generated by FTL. DO NOT EDIT.
package postgres

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/server"
)

type InsertClient func(context.Context, InsertRequest) (InsertResponse, error)

type QueryClient func(context.Context) ([]string, error)

//ftl:database postgres testdb
type TestdbConfig struct{}

type TestdbHandle = ftl.DatabaseHandle[TestdbConfig]

func init() {
	reflection.Register(
		reflection.Database[TestdbConfig]("testdb", server.InitPostgres),

		reflection.ProvideResourcesForVerb(
			Insert,
			server.DatabaseHandle[TestdbConfig]("testdb", "postgres"),
		),

		reflection.ProvideResourcesForVerb(
			Query,
			server.DatabaseHandle[TestdbConfig]("testdb", "postgres"),
		),
	)
}
