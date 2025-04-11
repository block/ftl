// Code generated by FTL. DO NOT EDIT.
package mysql

import (
	"context"
	"github.com/alecthomas/types/tuple"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/server"
)

type Author struct {
	Id        int
	Bio       ftl.Option[string]
	BirthYear ftl.Option[int]
	Hometown  ftl.Option[string]
}
type GetAuthorInfoRow struct {
	Bio      ftl.Option[string]
	Hometown ftl.Option[string]
}
type GetManyAuthorsInfoRow struct {
	Bio      ftl.Option[string]
	Hometown ftl.Option[string]
}

type CreateRequestClient func(context.Context, ftl.Option[string]) error

type GetAllAuthorsClient func(context.Context) ([]Author, error)

type GetAuthorByIdClient func(context.Context, int) (Author, error)

type GetAuthorInfoClient func(context.Context, int) (GetAuthorInfoRow, error)

type GetManyAuthorsInfoClient func(context.Context, int) ([]GetManyAuthorsInfoRow, error)

type GetRequestDataClient func(context.Context) ([]ftl.Option[string], error)

func init() {
	reflection.Register(
		server.QuerySink[ftl.Option[string]]("mysql", "createRequest", reflection.CommandTypeExec, "testdb", "mysql", "INSERT INTO requests (data) VALUES (?)", []string{}, []tuple.Pair[string, string]{}),
		server.QuerySource[Author]("mysql", "getAllAuthors", reflection.CommandTypeMany, "testdb", "mysql", "SELECT id, bio, birth_year, hometown FROM authors", []string{}, []tuple.Pair[string, string]{tuple.PairOf("id", "Id"), tuple.PairOf("bio", "Bio"), tuple.PairOf("birth_year", "BirthYear"), tuple.PairOf("hometown", "Hometown")}),
		server.Query[int, Author]("mysql", "getAuthorById", reflection.CommandTypeOne, "testdb", "mysql", "SELECT id, bio, birth_year, hometown FROM authors WHERE id = ?", []string{}, []tuple.Pair[string, string]{tuple.PairOf("id", "Id"), tuple.PairOf("bio", "Bio"), tuple.PairOf("birth_year", "BirthYear"), tuple.PairOf("hometown", "Hometown")}),
		server.Query[int, GetAuthorInfoRow]("mysql", "getAuthorInfo", reflection.CommandTypeOne, "testdb", "mysql", "SELECT bio, hometown FROM authors WHERE id = ?", []string{}, []tuple.Pair[string, string]{tuple.PairOf("bio", "Bio"), tuple.PairOf("hometown", "Hometown")}),
		server.Query[int, GetManyAuthorsInfoRow]("mysql", "getManyAuthorsInfo", reflection.CommandTypeMany, "testdb", "mysql", "SELECT bio, hometown FROM authors WHERE id IN (?)", []string{}, []tuple.Pair[string, string]{tuple.PairOf("bio", "Bio"), tuple.PairOf("hometown", "Hometown")}),
		server.QuerySource[ftl.Option[string]]("mysql", "getRequestData", reflection.CommandTypeMany, "testdb", "mysql", "SELECT data FROM requests", []string{}, []tuple.Pair[string, string]{}),
	)
}
