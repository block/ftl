package server

import (
	"context"
	"reflect"
	"strings"

	"github.com/alecthomas/types/tuple"
	"github.com/block/ftl/backend/runner/query"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
)

func Query[Req, Resp any](
	module string,
	verbName string,
	commandType reflection.CommandType,
	dbName string,
	rawSQL string,
	fields []string,
	colToFieldName []tuple.Pair[string, string],
) reflection.Registree {
	return reflection.Query(module, verbName, getQueryFunc[Req, Resp](commandType, dbName, rawSQL, getQueryParamValues(fields), colToFieldName))
}

func QueryEmpty(
	module string,
	verbName string,
	commandType reflection.CommandType,
	dbName string,
	rawSQL string,
	fields []string,
	colToFieldName []tuple.Pair[string, string],
) reflection.Registree {
	return reflection.Query(module, verbName, getQueryFunc[ftl.Unit, ftl.Unit](commandType, dbName, rawSQL, getQueryParamValues(fields), colToFieldName))
}

func QuerySource[Resp any](
	module string,
	verbName string,
	commandType reflection.CommandType,
	dbName string,
	rawSQL string,
	fields []string,
	colToFieldName []tuple.Pair[string, string],
) reflection.Registree {
	return reflection.Query(module, verbName, getQueryFunc[ftl.Unit, Resp](commandType, dbName, rawSQL, getQueryParamValues(fields), colToFieldName))
}

func QuerySink[Req any](
	module string,
	verbName string,
	commandType reflection.CommandType,
	dbName string,
	rawSQL string,
	fields []string,
	colToFieldName []tuple.Pair[string, string],
) reflection.Registree {
	return reflection.Query(module, verbName, getQueryFunc[Req, ftl.Unit](commandType, dbName, rawSQL, getQueryParamValues(fields), colToFieldName))
}

func getQueryParamValues(fields []string) func(req any) []any {
	return func(req any) []any {
		reqValue := reflect.ValueOf(req)
		if reqValue.Kind() == reflect.Ptr {
			reqValue = reqValue.Elem()
		}
		params := make([]any, 0, len(fields))
		for _, field := range fields {
			fieldValue := reqValue.FieldByName(field)
			if !fieldValue.IsValid() {
				continue
			}
			value := fieldValue.Interface()

			// handle ftl.Option
			if fieldType := fieldValue.Type(); fieldType.Kind() == reflect.Struct && fieldType.PkgPath() == "github.com/block/ftl/go-runtime/ftl" && strings.HasPrefix(fieldType.Name(), "Option") {
				getMethod := fieldValue.MethodByName("Get")
				results := getMethod.Call(nil)
				if !results[1].Bool() {
					value = nil
				} else {
					value = results[0].Interface()
				}
			}

			params = append(params, value)
		}
		return params
	}
}

func getQueryFunc[Req, Resp any](
	commandType reflection.CommandType,
	dbName string,
	rawSQL string,
	paramsFn func(req any) []any,
	colToFieldName []tuple.Pair[string, string],
) any {
	var fn any
	switch commandType {
	case reflection.CommandTypeExec:
		if reflect.TypeFor[Req]() != reflect.TypeFor[ftl.Unit]() {
			fn = func(ctx context.Context, req Req) error {
				params := paramsFn(req)
				return query.Exec[Req](ctx, dbName, rawSQL, params, colToFieldName)
			}
		} else {
			fn = func(ctx context.Context) error {
				return query.Exec[Req](ctx, dbName, rawSQL, []any{}, colToFieldName)
			}
		}
	case reflection.CommandTypeOne:
		if reflect.TypeFor[Req]() != reflect.TypeFor[ftl.Unit]() {
			fn = func(ctx context.Context, req Req) (Resp, error) {
				return query.One[Req, Resp](ctx, dbName, rawSQL, paramsFn(req), colToFieldName)
			}
		} else {
			fn = func(ctx context.Context) (Resp, error) {
				return query.One[Req, Resp](ctx, dbName, rawSQL, []any{}, colToFieldName)
			}
		}
	case reflection.CommandTypeMany:
		if reflect.TypeFor[Req]() != reflect.TypeFor[ftl.Unit]() {
			fn = func(ctx context.Context, req Req) ([]Resp, error) {
				return query.Many[Req, Resp](ctx, dbName, rawSQL, paramsFn(req), colToFieldName)
			}
		} else {
			fn = func(ctx context.Context) ([]Resp, error) {
				return query.Many[Req, Resp](ctx, dbName, rawSQL, []any{}, colToFieldName)
			}
		}
	}
	return fn
}
