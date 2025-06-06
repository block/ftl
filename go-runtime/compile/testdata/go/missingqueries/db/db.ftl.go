// Code generated by FTL. DO NOT EDIT.
package db

import (
    "context"
    "github.com/alecthomas/types/tuple"
    "github.com/block/ftl/common/reflection"
    "github.com/block/ftl/go-runtime/server"
    stdtime "time"
)

type InsertPriceQuery struct {
  	Code string
  	Price string
  	Time stdtime.Time
  	Currency string
}
	
type InsertPriceClient func(context.Context, InsertPriceQuery) error

func init() {
	reflection.Register(
		server.QuerySink[InsertPriceQuery]("missingqueries", "insertPrice", reflection.CommandTypeExec, "prices", "mysql", "INSERT INTO prices (code, price, time, currency) VALUES (?, ?, ?, ?)", []string{"Code","Price","Time","Currency"}, []tuple.Pair[string,string]{}),
	)
}