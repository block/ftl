// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package sql

import (
	"context"
	"encoding/json"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
	"github.com/sqlc-dev/pqtype"
)

const beginConsumingTopicEvent = `-- name: BeginConsumingTopicEvent :exec
WITH event AS (
    SELECT id, created_at, key, topic_id, payload, caller, request_key, trace_context
    FROM topic_events
    WHERE "key" = $2::topic_event_key
)
UPDATE topic_subscriptions
SET state = 'executing',
    cursor = (SELECT id FROM event)
WHERE key = $1::subscription_key
`

func (q *Queries) BeginConsumingTopicEvent(ctx context.Context, subscription model.SubscriptionKey, event model.TopicEventKey) error {
	_, err := q.db.ExecContext(ctx, beginConsumingTopicEvent, subscription, event)
	return err
}

const completeEventForSubscription = `-- name: CompleteEventForSubscription :exec
WITH module AS (
    SELECT id
    FROM modules
    WHERE name = $2::TEXT
)
UPDATE topic_subscriptions
SET state = 'idle'
WHERE name = $1::TEXT
  AND module_id = (SELECT id FROM module)
`

func (q *Queries) CompleteEventForSubscription(ctx context.Context, name string, module string) error {
	_, err := q.db.ExecContext(ctx, completeEventForSubscription, name, module)
	return err
}

const deleteSubscribers = `-- name: DeleteSubscribers :many
DELETE FROM topic_subscribers
WHERE deployment_id IN (
    SELECT deployments.id
    FROM deployments
    WHERE deployments.key = $1::deployment_key
)
RETURNING topic_subscribers.key
`

func (q *Queries) DeleteSubscribers(ctx context.Context, deployment model.DeploymentKey) ([]model.SubscriberKey, error) {
	rows, err := q.db.QueryContext(ctx, deleteSubscribers, deployment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.SubscriberKey
	for rows.Next() {
		var key model.SubscriberKey
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		items = append(items, key)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteSubscriptions = `-- name: DeleteSubscriptions :many
DELETE FROM topic_subscriptions
WHERE deployment_id IN (
    SELECT deployments.id
    FROM deployments
    WHERE deployments.key = $1::deployment_key
)
RETURNING topic_subscriptions.key
`

func (q *Queries) DeleteSubscriptions(ctx context.Context, deployment model.DeploymentKey) ([]model.SubscriptionKey, error) {
	rows, err := q.db.QueryContext(ctx, deleteSubscriptions, deployment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.SubscriptionKey
	for rows.Next() {
		var key model.SubscriptionKey
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		items = append(items, key)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNextEventForSubscription = `-- name: GetNextEventForSubscription :one
WITH cursor AS (
    SELECT
        created_at,
        id
    FROM topic_events
    WHERE "key" = $3::topic_event_key
)
SELECT events."key" as event,
       events.payload,
       events.created_at,
       events.caller,
       events.request_key,
       events.trace_context,
       NOW() - events.created_at >= $1::interval AS ready
FROM topics
         LEFT JOIN topic_events as events ON events.topic_id = topics.id
WHERE topics.key = $2::topic_key
  AND (events.created_at, events.id) > (SELECT COALESCE(MAX(cursor.created_at), '1900-01-01'), COALESCE(MAX(cursor.id), 0) FROM cursor)
ORDER BY events.created_at, events.id
LIMIT 1
`

type GetNextEventForSubscriptionRow struct {
	Event        optional.Option[model.TopicEventKey]
	Payload      api.OptionalEncryptedAsyncColumn
	CreatedAt    sqltypes.OptionalTime
	Caller       optional.Option[string]
	RequestKey   optional.Option[string]
	TraceContext pqtype.NullRawMessage
	Ready        bool
}

func (q *Queries) GetNextEventForSubscription(ctx context.Context, consumptionDelay sqltypes.Duration, topic model.TopicKey, cursor optional.Option[model.TopicEventKey]) (GetNextEventForSubscriptionRow, error) {
	row := q.db.QueryRowContext(ctx, getNextEventForSubscription, consumptionDelay, topic, cursor)
	var i GetNextEventForSubscriptionRow
	err := row.Scan(
		&i.Event,
		&i.Payload,
		&i.CreatedAt,
		&i.Caller,
		&i.RequestKey,
		&i.TraceContext,
		&i.Ready,
	)
	return i, err
}

const getRandomSubscriber = `-- name: GetRandomSubscriber :one
SELECT
    subscribers.sink as sink,
    subscribers.retry_attempts as retry_attempts,
    subscribers.backoff as backoff,
    subscribers.max_backoff as max_backoff,
    subscribers.catch_verb as catch_verb
FROM topic_subscribers as subscribers
         JOIN topic_subscriptions ON subscribers.topic_subscriptions_id = topic_subscriptions.id
WHERE topic_subscriptions.key = $1::subscription_key
ORDER BY RANDOM()
LIMIT 1
`

type GetRandomSubscriberRow struct {
	Sink          schema.RefKey
	RetryAttempts int32
	Backoff       sqltypes.Duration
	MaxBackoff    sqltypes.Duration
	CatchVerb     optional.Option[schema.RefKey]
}

func (q *Queries) GetRandomSubscriber(ctx context.Context, key model.SubscriptionKey) (GetRandomSubscriberRow, error) {
	row := q.db.QueryRowContext(ctx, getRandomSubscriber, key)
	var i GetRandomSubscriberRow
	err := row.Scan(
		&i.Sink,
		&i.RetryAttempts,
		&i.Backoff,
		&i.MaxBackoff,
		&i.CatchVerb,
	)
	return i, err
}

const getSubscription = `-- name: GetSubscription :one
WITH module AS (
    SELECT id
    FROM modules
    WHERE name = $2::TEXT
)
SELECT id, key, created_at, topic_id, module_id, deployment_id, name, cursor, state
FROM topic_subscriptions
WHERE name = $1::TEXT
  AND module_id = (SELECT id FROM module)
`

func (q *Queries) GetSubscription(ctx context.Context, column1 string, column2 string) (TopicSubscription, error) {
	row := q.db.QueryRowContext(ctx, getSubscription, column1, column2)
	var i TopicSubscription
	err := row.Scan(
		&i.ID,
		&i.Key,
		&i.CreatedAt,
		&i.TopicID,
		&i.ModuleID,
		&i.DeploymentID,
		&i.Name,
		&i.Cursor,
		&i.State,
	)
	return i, err
}

const getSubscriptionsNeedingUpdate = `-- name: GetSubscriptionsNeedingUpdate :many
WITH runner_count AS (
    SELECT count(r.deployment_id) as runner_count,
           r.deployment_id as deployment
    FROM runners r
    GROUP BY deployment
)
SELECT
    subs.key::subscription_key as key,
    curser.key as cursor,
    topics.key::topic_key as topic,
    subs.name
FROM topic_subscriptions subs
         JOIN runner_count on subs.deployment_id = runner_count.deployment
         LEFT JOIN topics ON subs.topic_id = topics.id
         LEFT JOIN topic_events curser ON subs.cursor = curser.id
WHERE subs.cursor IS DISTINCT FROM topics.head
  AND subs.state = 'idle'
ORDER BY curser.created_at
LIMIT 3
    FOR UPDATE OF subs SKIP LOCKED
`

type GetSubscriptionsNeedingUpdateRow struct {
	Key    model.SubscriptionKey
	Cursor optional.Option[model.TopicEventKey]
	Topic  model.TopicKey
	Name   string
}

// Results may not be ready to be scheduled yet due to event consumption delay
// Sorting ensures that brand new events (that may not be ready for consumption)
// don't prevent older events from being consumed
// We also make sure that the subscription belongs to a deployment that has at least one runner
func (q *Queries) GetSubscriptionsNeedingUpdate(ctx context.Context) ([]GetSubscriptionsNeedingUpdateRow, error) {
	rows, err := q.db.QueryContext(ctx, getSubscriptionsNeedingUpdate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSubscriptionsNeedingUpdateRow
	for rows.Next() {
		var i GetSubscriptionsNeedingUpdateRow
		if err := rows.Scan(
			&i.Key,
			&i.Cursor,
			&i.Topic,
			&i.Name,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTopic = `-- name: GetTopic :one
SELECT id, key, created_at, module_id, name, type, head
FROM topics
WHERE id = $1::BIGINT
`

func (q *Queries) GetTopic(ctx context.Context, dollar_1 int64) (Topic, error) {
	row := q.db.QueryRowContext(ctx, getTopic, dollar_1)
	var i Topic
	err := row.Scan(
		&i.ID,
		&i.Key,
		&i.CreatedAt,
		&i.ModuleID,
		&i.Name,
		&i.Type,
		&i.Head,
	)
	return i, err
}

const getTopicEvent = `-- name: GetTopicEvent :one
SELECT id, created_at, key, topic_id, payload, caller, request_key, trace_context
FROM topic_events
WHERE id = $1::BIGINT
`

func (q *Queries) GetTopicEvent(ctx context.Context, dollar_1 int64) (TopicEvent, error) {
	row := q.db.QueryRowContext(ctx, getTopicEvent, dollar_1)
	var i TopicEvent
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.Key,
		&i.TopicID,
		&i.Payload,
		&i.Caller,
		&i.RequestKey,
		&i.TraceContext,
	)
	return i, err
}

const insertSubscriber = `-- name: InsertSubscriber :exec
INSERT INTO topic_subscribers (
    key,
    topic_subscriptions_id,
    deployment_id,
    sink,
    retry_attempts,
    backoff,
    max_backoff,
    catch_verb
)
VALUES (
           $1::subscriber_key,
           (
               SELECT topic_subscriptions.id as id
               FROM topic_subscriptions
                        INNER JOIN modules ON topic_subscriptions.module_id = modules.id
               WHERE modules.name = $2::TEXT
                 AND topic_subscriptions.name = $3::TEXT
           ),
           (SELECT id FROM deployments WHERE key = $4::deployment_key),
           $5,
           $6,
           $7::interval,
           $8::interval,
           $9
       )
`

type InsertSubscriberParams struct {
	Key              model.SubscriberKey
	Module           string
	SubscriptionName string
	Deployment       model.DeploymentKey
	Sink             schema.RefKey
	RetryAttempts    int32
	Backoff          sqltypes.Duration
	MaxBackoff       sqltypes.Duration
	CatchVerb        optional.Option[schema.RefKey]
}

func (q *Queries) InsertSubscriber(ctx context.Context, arg InsertSubscriberParams) error {
	_, err := q.db.ExecContext(ctx, insertSubscriber,
		arg.Key,
		arg.Module,
		arg.SubscriptionName,
		arg.Deployment,
		arg.Sink,
		arg.RetryAttempts,
		arg.Backoff,
		arg.MaxBackoff,
		arg.CatchVerb,
	)
	return err
}

const publishEventForTopic = `-- name: PublishEventForTopic :exec
INSERT INTO topic_events (
    "key",
    topic_id,
    caller,
    payload,
    request_key,
    trace_context
)
VALUES (
           $1::topic_event_key,
           (
               SELECT topics.id
               FROM topics
                        INNER JOIN modules ON topics.module_id = modules.id
               WHERE modules.name = $2::TEXT
                 AND topics.name = $3::TEXT
           ),
           $4::TEXT,
           $5,
           $6::TEXT,
           $7::jsonb
       )
`

type PublishEventForTopicParams struct {
	Key          model.TopicEventKey
	Module       string
	Topic        string
	Caller       string
	Payload      api.EncryptedAsyncColumn
	RequestKey   string
	TraceContext json.RawMessage
}

func (q *Queries) PublishEventForTopic(ctx context.Context, arg PublishEventForTopicParams) error {
	_, err := q.db.ExecContext(ctx, publishEventForTopic,
		arg.Key,
		arg.Module,
		arg.Topic,
		arg.Caller,
		arg.Payload,
		arg.RequestKey,
		arg.TraceContext,
	)
	return err
}

const setSubscriptionCursor = `-- name: SetSubscriptionCursor :exec
WITH event AS (
    SELECT id, created_at, key, topic_id, payload
    FROM topic_events
    WHERE "key" = $2::topic_event_key
)
UPDATE topic_subscriptions
SET cursor = (SELECT id FROM event)
WHERE key = $1::subscription_key
`

func (q *Queries) SetSubscriptionCursor(ctx context.Context, column1 model.SubscriptionKey, column2 model.TopicEventKey) error {
	_, err := q.db.ExecContext(ctx, setSubscriptionCursor, column1, column2)
	return err
}

const upsertSubscription = `-- name: UpsertSubscription :one
INSERT INTO topic_subscriptions (
    key,
    topic_id,
    module_id,
    deployment_id,
    name)
VALUES (
           $1::subscription_key,
           (
               SELECT topics.id as id
               FROM topics
                        INNER JOIN modules ON topics.module_id = modules.id
               WHERE modules.name = $2::TEXT
                 AND topics.name = $3::TEXT
           ),
           (SELECT id FROM modules WHERE name = $4::TEXT),
           (SELECT id FROM deployments WHERE key = $5::deployment_key),
           $6::TEXT
       )
ON CONFLICT (name, module_id) DO
    UPDATE SET
               topic_id = excluded.topic_id,
               deployment_id = (SELECT id FROM deployments WHERE key = $5::deployment_key)
RETURNING
    id,
    CASE
        WHEN xmax = 0 THEN true
        ELSE false
        END AS inserted
`

type UpsertSubscriptionParams struct {
	Key         model.SubscriptionKey
	TopicModule string
	TopicName   string
	Module      string
	Deployment  model.DeploymentKey
	Name        string
}

type UpsertSubscriptionRow struct {
	ID       int64
	Inserted bool
}

func (q *Queries) UpsertSubscription(ctx context.Context, arg UpsertSubscriptionParams) (UpsertSubscriptionRow, error) {
	row := q.db.QueryRowContext(ctx, upsertSubscription,
		arg.Key,
		arg.TopicModule,
		arg.TopicName,
		arg.Module,
		arg.Deployment,
		arg.Name,
	)
	var i UpsertSubscriptionRow
	err := row.Scan(&i.ID, &i.Inserted)
	return i, err
}

const upsertTopic = `-- name: UpsertTopic :exec
INSERT INTO topics (key, module_id, name, type)
VALUES (
           $1::topic_key,
           (SELECT id FROM modules WHERE name = $2::TEXT LIMIT 1),
           $3::TEXT,
           $4::TEXT
       )
ON CONFLICT (name, module_id) DO
    UPDATE SET
    type = $4::TEXT
RETURNING id
`

type UpsertTopicParams struct {
	Topic     model.TopicKey
	Module    string
	Name      string
	EventType string
}

func (q *Queries) UpsertTopic(ctx context.Context, arg UpsertTopicParams) error {
	_, err := q.db.ExecContext(ctx, upsertTopic,
		arg.Topic,
		arg.Module,
		arg.Name,
		arg.EventType,
	)
	return err
}
