// Code generated by FTL. DO NOT EDIT.
package mysql

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/server"
)

type InsertClient func(context.Context, InsertRequest) (InsertResponse, error)

type QueryClient func(context.Context) (map[string]string, error)

//ftl:database mysql testdb
type TestdbConfig struct{}

type TestdbHandle = ftl.DatabaseHandle[TestdbConfig]

func init() {
	reflection.Register(
		reflection.Database[TestdbConfig]("testdb", server.InitMySQL),

		reflection.ProvideResourcesForVerb(
			Insert,
			server.SinkClient[CreateRequestClient, string](),
		),

		reflection.ProvideResourcesForVerb(
			Query,
			server.SourceClient[GetRequestDataClient, []string](),
		),
	)
}
