// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: async_queries.sql

package sql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/block/ftl/backend/controller/encryption/api"
	"github.com/block/ftl/backend/controller/sql/sqltypes"
	"github.com/block/ftl/internal/schema"
	"github.com/alecthomas/types/optional"
)

const asyncCallQueueDepth = `-- name: AsyncCallQueueDepth :one
SELECT count(*)
FROM async_calls
WHERE state = 'pending' AND scheduled_at <= (NOW() AT TIME ZONE 'utc')
`

func (q *Queries) AsyncCallQueueDepth(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, asyncCallQueueDepth)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createAsyncCall = `-- name: CreateAsyncCall :one
INSERT INTO async_calls (
  scheduled_at,
  verb,
  origin,
  request,
  remaining_attempts,
  backoff,
  max_backoff,
  catch_verb,
  parent_request_key,
  trace_context
)
VALUES (
  $1::TIMESTAMPTZ,
  $2,
  $3,
  $4,
  $5,
  $6::interval,
  $7::interval,
  $8,
  $9,
  $10::jsonb
)
RETURNING id
`

type CreateAsyncCallParams struct {
	ScheduledAt       time.Time
	Verb              schema.RefKey
	Origin            string
	Request           api.EncryptedAsyncColumn
	RemainingAttempts int32
	Backoff           sqltypes.Duration
	MaxBackoff        sqltypes.Duration
	CatchVerb         optional.Option[schema.RefKey]
	ParentRequestKey  optional.Option[string]
	TraceContext      json.RawMessage
}

func (q *Queries) CreateAsyncCall(ctx context.Context, arg CreateAsyncCallParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createAsyncCall,
		arg.ScheduledAt,
		arg.Verb,
		arg.Origin,
		arg.Request,
		arg.RemainingAttempts,
		arg.Backoff,
		arg.MaxBackoff,
		arg.CatchVerb,
		arg.ParentRequestKey,
		arg.TraceContext,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}
