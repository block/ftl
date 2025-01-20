package query

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/tuple"
	"google.golang.org/protobuf/types/known/timestamppb"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

// SQLNullable is an interface that allows for setting a value to null/none. Optional types implement this interface.
type SQLNullable interface {
	// SetNull sets the value to null/none
	SetNull()
	// SetValue sets the underlying value
	SetValue(value reflect.Value) error
	// GetType returns the type of the value in the Option
	GetType() reflect.Type
}

func One[Req, Resp any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) (resp Resp, err error) {
	client, err := newClient(dbName)
	if err != nil {
		return resp, err
	}
	respType := reflect.TypeFor[Resp]()
	results, err := client.performQuery(ctx, respType, querypb.CommandType_COMMAND_TYPE_ONE, rawSQL, params, colFieldNames)
	if err != nil {
		return resp, err
	}
	if len(results) == 0 {
		return resp, fmt.Errorf("no results found")
	}
	return results[0].(Resp), nil //nolint:forcetypeassert
}

func Many[Req, Resp any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]Resp, error) {
	client, err := newClient(dbName)
	if err != nil {
		return nil, err
	}
	respType := reflect.TypeFor[Resp]()
	results, err := client.performQuery(ctx, respType, querypb.CommandType_COMMAND_TYPE_MANY, rawSQL, params, colFieldNames)
	if err != nil {
		return nil, err
	}
	typed := make([]Resp, len(results))
	for i, r := range results {
		typed[i] = r.(Resp) //nolint:forcetypeassert
	}
	return typed, nil
}

func Exec[Req any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) error {
	client, err := newClient(dbName)
	if err != nil {
		return err
	}
	_, err = client.performQuery(ctx, nil, querypb.CommandType_COMMAND_TYPE_EXEC, rawSQL, params, colFieldNames)
	return err
}

type client struct {
	queryconnect.QueryServiceClient
}

func newClient(dbName string) (*client, error) {
	address := os.Getenv(strings.ToUpper("FTL_QUERY_ADDRESS_" + dbName))
	if address == "" {
		return nil, fmt.Errorf("query address for %s not found", dbName)
	}
	return &client{QueryServiceClient: rpc.Dial(queryconnect.NewQueryServiceClient, address, log.Error)}, nil
}

// performQuery performs a SQL query and returns the results.
//
// note: accepts colFieldNames as []tuple.Pair[string, string] rather than map[string]string to preserve column order
func (c *client) performQuery(ctx context.Context, respType reflect.Type, commandType querypb.CommandType, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]any, error) {
	sqlParams := make([]*querypb.SQLValue, 0, len(params))
	for _, param := range params {
		sqlValue, err := convertParamToSQLValue(param)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter %v of type %T: %w", param, param, err)
		}
		sqlParams = append(sqlParams, sqlValue)
	}

	colToFieldName := map[string]string{}
	for _, pair := range colFieldNames {
		colToFieldName[pair.A] = pair.B
	}

	resultCols := make([]string, 0, len(colFieldNames))
	for _, pair := range colFieldNames {
		resultCols = append(resultCols, pair.A)
	}

	stream, err := c.QueryServiceClient.ExecuteQuery(ctx, connect.NewRequest(&querypb.ExecuteQueryRequest{
		RawSql:        rawSQL,
		CommandType:   commandType,
		Parameters:    sqlParams,
		ResultColumns: resultCols,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var results []any
	for stream.Receive() {
		resp := stream.Msg()
		switch r := resp.Result.(type) {
		case *querypb.ExecuteQueryResponse_ExecResult:
			return nil, nil
		case *querypb.ExecuteQueryResponse_RowResults:
			if len(r.RowResults.Rows) == 0 {
				continue
			}

			respValue := reflect.New(respType).Elem()
			for col, field := range colToFieldName {
				if sqlVal, ok := r.RowResults.Rows[col]; ok {
					fieldValue := respValue.FieldByName(field)
					if !fieldValue.IsValid() {
						continue
					}
					if err := setFieldFromSQLValue(fieldValue, sqlVal); err != nil {
						return nil, fmt.Errorf("failed to set field %s from column %s: value type %T = %v: %w",
							field, col, sqlVal.Value, sqlVal.Value, err)
					}
				}
			}
			results = append(results, respValue.Interface())
		}
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("failed to receive query results: %w", err)
	}
	return results, nil
}

func setFieldFromSQLValue(field reflect.Value, sqlValue *querypb.SQLValue) error {
	if !field.CanSet() {
		return fmt.Errorf("field is not settable")
	}

	var nullable SQLNullable
	var ok bool
	if field.Kind() == reflect.Ptr {
		nullable, ok = field.Interface().(SQLNullable)
	} else if field.CanAddr() {
		nullable, ok = field.Addr().Interface().(SQLNullable)
	}

	// handle null values
	if sqlValue == nil || sqlValue.Value == nil {
		if ok {
			nullable.SetNull()
			return nil
		}
		return nil
	}
	if _, isNull := sqlValue.Value.(*querypb.SQLValue_NullValue); isNull {
		if ok {
			nullable.SetNull()
			return nil
		}
		return nil
	}

	// handle non-null values
	if ok {
		var newValue reflect.Value
		if field.Kind() == reflect.Ptr {
			newValue = reflect.New(field.Type().Elem())
			if err := setNonNullValue(newValue.Elem(), sqlValue); err != nil {
				return err
			}
		} else {
			newValue = reflect.New(nullable.GetType()).Elem()
			if err := setNonNullValue(newValue, sqlValue); err != nil {
				return err
			}
			return nil
		}
		err := nullable.SetValue(newValue)
		if err != nil {
			return fmt.Errorf("failed to set value: %w", err)
		}
		return nil
	}
	return setNonNullValue(field, sqlValue)
}

func setNonNullValue(field reflect.Value, sqlValue *querypb.SQLValue) error {
	switch v := sqlValue.Value.(type) {
	case *querypb.SQLValue_StringValue:
		if field.Kind() != reflect.String {
			return fmt.Errorf("cannot convert string value %q to field type %s", v.StringValue, field.Type())
		}
		field.SetString(v.StringValue)

	case *querypb.SQLValue_IntValue:
		switch field.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32:
			field.SetInt(v.IntValue)
		case reflect.Float64, reflect.Float32:
			field.SetFloat(float64(v.IntValue))
		case reflect.Bool:
			field.SetBool(v.IntValue != 0)
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				field.Set(reflect.ValueOf(time.Unix(v.IntValue, 0)))
			} else {
				return fmt.Errorf("cannot convert int value %d to field type %s", v.IntValue, field.Type())
			}
		default:
			return fmt.Errorf("cannot convert int value %d to field type %s", v.IntValue, field.Type())
		}

	case *querypb.SQLValue_FloatValue:
		switch field.Kind() {
		case reflect.Float64, reflect.Float32:
			field.SetFloat(v.FloatValue)
		case reflect.Int, reflect.Int64, reflect.Int32:
			field.SetInt(int64(v.FloatValue))
		case reflect.Bool:
			field.SetBool(v.FloatValue != 0)
		default:
			return fmt.Errorf("cannot convert float value %f to field type %s", v.FloatValue, field.Type())
		}

	case *querypb.SQLValue_BoolValue:
		if field.Kind() != reflect.Bool {
			return fmt.Errorf("cannot convert bool value %v to field type %s", v.BoolValue, field.Type())
		}
		field.SetBool(v.BoolValue)

	case *querypb.SQLValue_BytesValue:
		str := string(v.BytesValue)
		switch field.Kind() {
		case reflect.String:
			field.SetString(str)
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.Uint8 {
				field.SetBytes(v.BytesValue)
			} else {
				return fmt.Errorf("cannot convert bytes value to field type %s", field.Type())
			}
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				for _, format := range []string{
					"2006-01-02 15:04:05",
					"2006-01-02 15:04:05.000000",
					time.RFC3339,
					time.RFC3339Nano,
				} {
					if t, err := time.Parse(format, str); err == nil {
						field.Set(reflect.ValueOf(t))
						return nil
					}
				}
				return fmt.Errorf("cannot parse time string %q", str)
			}
			return fmt.Errorf("cannot convert bytes value to field type %s", field.Type())
		default:
			return fmt.Errorf("cannot convert bytes value to field type %s", field.Type())
		}

	case *querypb.SQLValue_TimestampValue:
		if field.Type() != reflect.TypeOf(time.Time{}) {
			return fmt.Errorf("cannot convert timestamp value %v to field type %s", v.TimestampValue.AsTime(), field.Type())
		}
		field.Set(reflect.ValueOf(v.TimestampValue.AsTime()))

	default:
		return fmt.Errorf("unsupported SQLValue type: %T", v)
	}
	return nil
}

func convertParamToSQLValue(v any) (*querypb.SQLValue, error) {
	if v == nil {
		return &querypb.SQLValue{Value: &querypb.SQLValue_NullValue{NullValue: true}}, nil
	}

	switch val := v.(type) {
	case string:
		return &querypb.SQLValue{Value: &querypb.SQLValue_StringValue{StringValue: val}}, nil
	case int, int32, int64:
		return &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: reflect.ValueOf(val).Int()}}, nil
	case float32, float64:
		return &querypb.SQLValue{Value: &querypb.SQLValue_FloatValue{FloatValue: reflect.ValueOf(val).Float()}}, nil
	case bool:
		return &querypb.SQLValue{Value: &querypb.SQLValue_BoolValue{BoolValue: val}}, nil
	case []byte:
		return &querypb.SQLValue{Value: &querypb.SQLValue_BytesValue{BytesValue: val}}, nil
	case time.Time:
		return &querypb.SQLValue{Value: &querypb.SQLValue_TimestampValue{TimestampValue: timestamppb.New(val)}}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
