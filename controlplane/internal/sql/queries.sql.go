// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: queries.sql

package sql

import (
	"context"

	"github.com/TBD54566975/ftl/controlplane/internal/sqltypes"
	"github.com/jackc/pgx/v5/pgtype"
)

const associateArtefactWithDeployment = `-- name: AssociateArtefactWithDeployment :exec
INSERT INTO deployment_artefacts (deployment_id, artefact_id, executable, path)
VALUES ((SELECT id FROM deployments WHERE key = $1), $2, $3, $4)
`

type AssociateArtefactWithDeploymentParams struct {
	Key        sqltypes.Key
	ArtefactID int64
	Executable bool
	Path       string
}

func (q *Queries) AssociateArtefactWithDeployment(ctx context.Context, arg AssociateArtefactWithDeploymentParams) error {
	_, err := q.db.Exec(ctx, associateArtefactWithDeployment,
		arg.Key,
		arg.ArtefactID,
		arg.Executable,
		arg.Path,
	)
	return err
}

const claimRunner = `-- name: ClaimRunner :one
UPDATE runners
SET state               = 'claimed',
    reservation_timeout = $2,
    deployment_id       = COALESCE((SELECT id
                                    FROM deployments d
                                    WHERE d.key = $3
                                    LIMIT 1), -1)
WHERE id = (SELECT id
            FROM runners r
            WHERE r.language = $1
              AND r.state = 'idle'
            LIMIT 1 FOR UPDATE SKIP LOCKED)
RETURNING runners.id, runners.key, runners.last_seen, runners.reservation_timeout, runners.state, runners.language, runners.endpoint, runners.deployment_id
`

// Find an idle runner and claim it for the given deployment.
func (q *Queries) ClaimRunner(ctx context.Context, language string, reservationTimeout pgtype.Timestamptz, deploymentKey sqltypes.Key) (Runner, error) {
	row := q.db.QueryRow(ctx, claimRunner, language, reservationTimeout, deploymentKey)
	var i Runner
	err := row.Scan(
		&i.ID,
		&i.Key,
		&i.LastSeen,
		&i.ReservationTimeout,
		&i.State,
		&i.Language,
		&i.Endpoint,
		&i.DeploymentID,
	)
	return i, err
}

const createArtefact = `-- name: CreateArtefact :one
INSERT INTO artefacts (digest, content)
VALUES ($1, $2)
RETURNING id
`

// Create a new artefact and return the artefact ID.
func (q *Queries) CreateArtefact(ctx context.Context, digest []byte, content []byte) (int64, error) {
	row := q.db.QueryRow(ctx, createArtefact, digest, content)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const createDeployment = `-- name: CreateDeployment :exec
INSERT INTO deployments (module_id, "schema", key)
VALUES ((SELECT id FROM modules WHERE name = $2::TEXT LIMIT 1), $3::BYTEA, $1)
`

func (q *Queries) CreateDeployment(ctx context.Context, key sqltypes.Key, moduleName string, schema []byte) error {
	_, err := q.db.Exec(ctx, createDeployment, key, moduleName, schema)
	return err
}

const deleteStaleRunners = `-- name: DeleteStaleRunners :one
WITH deleted AS (
    DELETE FROM runners
        WHERE last_seen < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM deleted
`

func (q *Queries) DeleteStaleRunners(ctx context.Context, dollar_1 pgtype.Interval) (int64, error) {
	row := q.db.QueryRow(ctx, deleteStaleRunners, dollar_1)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deregisterRunner = `-- name: DeregisterRunner :one
WITH deleted AS (
    DELETE FROM runners WHERE key = $1
        RETURNING 1)
SELECT COUNT(*)
FROM deleted
`

func (q *Queries) DeregisterRunner(ctx context.Context, key sqltypes.Key) (int64, error) {
	row := q.db.QueryRow(ctx, deregisterRunner, key)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const expireRunnerReservations = `-- name: ExpireRunnerReservations :one
WITH rows AS (
    UPDATE runners
        SET state = 'idle',
            deployment_id = NULL,
            reservation_timeout = NULL
        WHERE (state = 'reserved' OR state = 'claimed')
            AND reservation_timeout < (NOW() AT TIME ZONE 'utc')
        RETURNING 1)
SELECT COUNT(*)
FROM rows
`

func (q *Queries) ExpireRunnerReservations(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, expireRunnerReservations)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getArtefactContentRange = `-- name: GetArtefactContentRange :one
SELECT SUBSTRING(a.content FROM $1 FOR $2)::BYTEA AS content
FROM artefacts a
WHERE a.id = $3
`

func (q *Queries) GetArtefactContentRange(ctx context.Context, start int32, count int32, iD int64) ([]byte, error) {
	row := q.db.QueryRow(ctx, getArtefactContentRange, start, count, iD)
	var content []byte
	err := row.Scan(&content)
	return content, err
}

const getArtefactDigests = `-- name: GetArtefactDigests :many
SELECT id, digest
FROM artefacts
WHERE digest = ANY ($1::bytea[])
`

type GetArtefactDigestsRow struct {
	ID     int64
	Digest []byte
}

// Return the digests that exist in the database.
func (q *Queries) GetArtefactDigests(ctx context.Context, digests [][]byte) ([]GetArtefactDigestsRow, error) {
	rows, err := q.db.Query(ctx, getArtefactDigests, digests)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetArtefactDigestsRow
	for rows.Next() {
		var i GetArtefactDigestsRow
		if err := rows.Scan(&i.ID, &i.Digest); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeployment = `-- name: GetDeployment :one
SELECT d.id, d.created_at, d.module_id, d.key, d.schema, d.min_replicas, m.language, m.name AS module_name
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
WHERE d.key = $1
`

type GetDeploymentRow struct {
	ID          int64
	CreatedAt   pgtype.Timestamptz
	ModuleID    int64
	Key         sqltypes.Key
	Schema      []byte
	MinReplicas int32
	Language    string
	ModuleName  string
}

func (q *Queries) GetDeployment(ctx context.Context, key sqltypes.Key) (GetDeploymentRow, error) {
	row := q.db.QueryRow(ctx, getDeployment, key)
	var i GetDeploymentRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.ModuleID,
		&i.Key,
		&i.Schema,
		&i.MinReplicas,
		&i.Language,
		&i.ModuleName,
	)
	return i, err
}

const getDeploymentArtefacts = `-- name: GetDeploymentArtefacts :many
SELECT da.created_at, artefact_id AS id, executable, path, digest, executable
FROM deployment_artefacts da
         INNER JOIN artefacts ON artefacts.id = da.artefact_id
WHERE deployment_id = $1
`

type GetDeploymentArtefactsRow struct {
	CreatedAt    pgtype.Timestamptz
	ID           int64
	Executable   bool
	Path         string
	Digest       []byte
	Executable_2 bool
}

// Get all artefacts matching the given digests.
func (q *Queries) GetDeploymentArtefacts(ctx context.Context, deploymentID int64) ([]GetDeploymentArtefactsRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentArtefacts, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentArtefactsRow
	for rows.Next() {
		var i GetDeploymentArtefactsRow
		if err := rows.Scan(
			&i.CreatedAt,
			&i.ID,
			&i.Executable,
			&i.Path,
			&i.Digest,
			&i.Executable_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentsByID = `-- name: GetDeploymentsByID :many
SELECT id, created_at, module_id, key, schema, min_replicas
FROM deployments
WHERE id = ANY ($1::BIGINT[])
`

func (q *Queries) GetDeploymentsByID(ctx context.Context, ids []int64) ([]Deployment, error) {
	rows, err := q.db.Query(ctx, getDeploymentsByID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Deployment
	for rows.Next() {
		var i Deployment
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.ModuleID,
			&i.Key,
			&i.Schema,
			&i.MinReplicas,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentsNeedingReconciliation = `-- name: GetDeploymentsNeedingReconciliation :many
SELECT d.key               AS key,
       m.name              AS module_name,
       m.language          AS language,
       COUNT(r.id)         AS assigned_runners_count,
       SUM(d.min_replicas) AS required_runners_count
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
         LEFT JOIN runners r ON d.id = r.deployment_id AND r.state = 'assigned'
GROUP BY d.key, m.name, m.language
HAVING COUNT(r.id) != SUM(d.min_replicas)
`

type GetDeploymentsNeedingReconciliationRow struct {
	Key                  sqltypes.Key
	ModuleName           string
	Language             string
	AssignedRunnersCount int64
	RequiredRunnersCount int64
}

// Get deployments that have a mismatch between the number of assigned and required replicas.
func (q *Queries) GetDeploymentsNeedingReconciliation(ctx context.Context) ([]GetDeploymentsNeedingReconciliationRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentsNeedingReconciliation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentsNeedingReconciliationRow
	for rows.Next() {
		var i GetDeploymentsNeedingReconciliationRow
		if err := rows.Scan(
			&i.Key,
			&i.ModuleName,
			&i.Language,
			&i.AssignedRunnersCount,
			&i.RequiredRunnersCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentsWithArtefacts = `-- name: GetDeploymentsWithArtefacts :many
SELECT d.id, d.created_at, d.key, m.name
FROM deployments d
         INNER JOIN modules m ON d.module_id = m.id
WHERE EXISTS (SELECT 1
              FROM deployment_artefacts da
                       INNER JOIN artefacts a ON da.artefact_id = a.id
              WHERE a.digest = ANY ($1::bytea[])
                AND da.deployment_id = d.id
              HAVING COUNT(*) = $2 -- Number of unique digests provided
)
`

type GetDeploymentsWithArtefactsRow struct {
	ID        int64
	CreatedAt pgtype.Timestamptz
	Key       sqltypes.Key
	Name      string
}

// Get all deployments that have artefacts matching the given digests.
func (q *Queries) GetDeploymentsWithArtefacts(ctx context.Context, digests [][]byte, count interface{}) ([]GetDeploymentsWithArtefactsRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentsWithArtefacts, digests, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentsWithArtefactsRow
	for rows.Next() {
		var i GetDeploymentsWithArtefactsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.Key,
			&i.Name,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIdleRunnerCountsByLanguage = `-- name: GetIdleRunnerCountsByLanguage :many
SELECT language, COUNT(*) AS count
FROM runners
WHERE state = 'idle'
GROUP BY language
ORDER BY language
`

type GetIdleRunnerCountsByLanguageRow struct {
	Language string
	Count    int64
}

func (q *Queries) GetIdleRunnerCountsByLanguage(ctx context.Context) ([]GetIdleRunnerCountsByLanguageRow, error) {
	rows, err := q.db.Query(ctx, getIdleRunnerCountsByLanguage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetIdleRunnerCountsByLanguageRow
	for rows.Next() {
		var i GetIdleRunnerCountsByLanguageRow
		if err := rows.Scan(&i.Language, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIdleRunnersForLanguage = `-- name: GetIdleRunnersForLanguage :many
SELECT id, key, last_seen, reservation_timeout, state, language, endpoint, deployment_id
FROM runners
WHERE language = $1
  AND state = 'idle'
LIMIT $2
`

func (q *Queries) GetIdleRunnersForLanguage(ctx context.Context, language string, limit int32) ([]Runner, error) {
	rows, err := q.db.Query(ctx, getIdleRunnersForLanguage, language, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Runner
	for rows.Next() {
		var i Runner
		if err := rows.Scan(
			&i.ID,
			&i.Key,
			&i.LastSeen,
			&i.ReservationTimeout,
			&i.State,
			&i.Language,
			&i.Endpoint,
			&i.DeploymentID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getModulesByID = `-- name: GetModulesByID :many
SELECT id, language, name
FROM modules
WHERE id = ANY ($1::BIGINT[])
`

func (q *Queries) GetModulesByID(ctx context.Context, ids []int64) ([]Module, error) {
	rows, err := q.db.Query(ctx, getModulesByID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Module
	for rows.Next() {
		var i Module
		if err := rows.Scan(&i.ID, &i.Language, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoutingTable = `-- name: GetRoutingTable :many
SELECT endpoint
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
         INNER JOIN modules m on d.module_id = m.id
WHERE state = 'assigned'
  AND m.name = $1
`

func (q *Queries) GetRoutingTable(ctx context.Context, name string) ([]string, error) {
	rows, err := q.db.Query(ctx, getRoutingTable, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var endpoint string
		if err := rows.Scan(&endpoint); err != nil {
			return nil, err
		}
		items = append(items, endpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRunnerState = `-- name: GetRunnerState :one
SELECT state
FROM runners
WHERE key = $1
`

func (q *Queries) GetRunnerState(ctx context.Context, key sqltypes.Key) (RunnerState, error) {
	row := q.db.QueryRow(ctx, getRunnerState, key)
	var state RunnerState
	err := row.Scan(&state)
	return state, err
}

const insertDeploymentLogEntry = `-- name: InsertDeploymentLogEntry :exec
INSERT INTO deployment_logs (deployment_id, time_stamp, level, scope, message, error)
VALUES ((SELECT id FROM deployments WHERE key = $1 LIMIT 1)::UUID, $2, $3, $4, $5, $6)
`

type InsertDeploymentLogEntryParams struct {
	Key       sqltypes.Key
	TimeStamp pgtype.Timestamptz
	Level     int32
	Scope     string
	Message   string
	Error     pgtype.Text
}

func (q *Queries) InsertDeploymentLogEntry(ctx context.Context, arg InsertDeploymentLogEntryParams) error {
	_, err := q.db.Exec(ctx, insertDeploymentLogEntry,
		arg.Key,
		arg.TimeStamp,
		arg.Level,
		arg.Scope,
		arg.Message,
		arg.Error,
	)
	return err
}

const insertMetricEntry = `-- name: InsertMetricEntry :exec
INSERT INTO metrics (runner_key, start_time, end_time, source_module, source_verb, dest_module, dest_verb, name, type,
                     value)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

type InsertMetricEntryParams struct {
	RunnerKey    sqltypes.Key
	StartTime    pgtype.Timestamptz
	EndTime      pgtype.Timestamptz
	SourceModule string
	SourceVerb   string
	DestModule   string
	DestVerb     string
	Name         string
	Type         MetricType
	Value        []byte
}

func (q *Queries) InsertMetricEntry(ctx context.Context, arg InsertMetricEntryParams) error {
	_, err := q.db.Exec(ctx, insertMetricEntry,
		arg.RunnerKey,
		arg.StartTime,
		arg.EndTime,
		arg.SourceModule,
		arg.SourceVerb,
		arg.DestModule,
		arg.DestVerb,
		arg.Name,
		arg.Type,
		arg.Value,
	)
	return err
}

const setDeploymentDesiredReplicas = `-- name: SetDeploymentDesiredReplicas :exec
UPDATE deployments
SET min_replicas = $2
WHERE key = $1
RETURNING 1
`

func (q *Queries) SetDeploymentDesiredReplicas(ctx context.Context, key sqltypes.Key, minReplicas int32) error {
	_, err := q.db.Exec(ctx, setDeploymentDesiredReplicas, key, minReplicas)
	return err
}

const upsertModule = `-- name: UpsertModule :one
INSERT INTO modules (language, name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id
`

func (q *Queries) UpsertModule(ctx context.Context, language string, name string) (int64, error) {
	row := q.db.QueryRow(ctx, upsertModule, language, name)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const upsertRunner = `-- name: UpsertRunner :one
WITH deployment_rel AS (
    SELECT CASE
               WHEN $5::UUID IS NULL THEN NULL
               ELSE COALESCE((SELECT id
                              FROM deployments d
                              WHERE d.key = $5
                              LIMIT 1), -1) END AS id)
INSERT
INTO runners (key, language, endpoint, state, deployment_id, last_seen)
VALUES ($1, $2, $3, $4, (SELECT id FROM deployment_rel), NOW() AT TIME ZONE 'utc')
ON CONFLICT (key) DO UPDATE SET language      = $2,
                                endpoint      = $3,
                                state         = $4,
                                deployment_id = (SELECT id FROM deployment_rel),
                                last_seen     = NOW() AT TIME ZONE 'utc'
RETURNING deployment_id
`

type UpsertRunnerParams struct {
	Key           sqltypes.Key
	Language      string
	Endpoint      string
	State         RunnerState
	DeploymentKey pgtype.UUID
}

// Upsert a runner and return the deployment ID that it is assigned to, if any.
// If the deployment key is null, then deployment_rel.id will be null,
// otherwise we try to retrieve the deployments.id using the key. If
// there is no corresponding deployment, then the deployment ID is -1
// and the parent statement will fail due to a foreign key constraint.
func (q *Queries) UpsertRunner(ctx context.Context, arg UpsertRunnerParams) (pgtype.Int8, error) {
	row := q.db.QueryRow(ctx, upsertRunner,
		arg.Key,
		arg.Language,
		arg.Endpoint,
		arg.State,
		arg.DeploymentKey,
	)
	var deployment_id pgtype.Int8
	err := row.Scan(&deployment_id)
	return deployment_id, err
}
