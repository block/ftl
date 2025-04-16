package mysql

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/go-runtime/ftl/ftltest"
)

func TestDatabase(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WithCallsAllowedWithinModule(),
		ftltest.WithSQLVerbsEnabled(),
	)

	_, err := ftltest.Call[InsertClient, InsertRequest, InsertResponse](ctx, InsertRequest{Data: "unit test 1"})
	assert.NoError(t, err)
	list, err := queryDBDirectly(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "unit test 1", list[0])

	_, err = ftltest.Call[InsertClient, InsertRequest, InsertResponse](ctx, InsertRequest{Data: "unit test 2"})
	assert.NoError(t, err)
	list, err = queryDBDirectly(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, "unit test 2", list[1])
	results, err := ftltest.CallSource[QueryClient, []string](ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "unit test 1", results[0])
	assert.Equal(t, "unit test 2", results[1])
}

func queryDBDirectly(ctx context.Context) ([]string, error) {
	db, err := ftltest.GetDatabaseHandle[TestdbConfig]()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	rows, err := db.Get(ctx).Query("SELECT data FROM requests ORDER BY created_at;")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	list := []string{}
	for rows.Next() {
		var data string
		err := rows.Scan(&data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		list = append(list, data)
	}
	return list, nil
}
