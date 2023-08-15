-- name: UpsertModule :one
INSERT INTO modules (language, name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id;

-- name: GetDeploymentsByID :many
SELECT *
FROM deployments
WHERE id = ANY (@ids::BIGINT[]);

-- name: GetModulesByID :many
SELECT *
FROM modules
WHERE id = ANY (@ids::BIGINT[]);

-- name: CreateDeployment :exec
INSERT INTO deployments (module_id, "schema", key)
VALUES ((SELECT id FROM modules WHERE name = @module_name::TEXT LIMIT 1), @schema::BYTEA, $1);

-- name: GetArtefactDigests :many
-- Return the digests that exist in the database.
SELECT id, digest
FROM artefacts
WHERE digest = ANY (@digests::bytea[]);

-- name: GetDeploymentArtefacts :many
-- Get all artefacts matching the given digests.
SELECT da.created_at, artefact_id AS id, executable, path, digest, executable
FROM deployment_artefacts da
         INNER JOIN artefacts ON artefacts.id = da.artefact_id
WHERE deployment_id = $1;

-- name: CreateArtefact :one
-- Create a new artefact and return the artefact ID.
INSERT INTO artefacts (digest, content)
VALUES ($1, $2)
RETURNING id;

-- name: AssociateArtefactWithDeployment :exec
INSERT INTO deployment_artefacts (deployment_id, artefact_id, executable, path)
VALUES ((SELECT id FROM deployments WHERE key = $1), $2, $3, $4);

-- name: ReplaceDeployment :one
WITH update_container AS (
    UPDATE deployments AS d
        SET min_replicas = update_deployments.min_replicas
        FROM (VALUES (sqlc.arg('old_deployment')::UUID, 0),
                     (sqlc.arg('new_deployment')::UUID, sqlc.arg('min_replicas')::INT))
            AS update_deployments(key, min_replicas)
        WHERE d.key = update_deployments.key
        RETURNING 1)
SELECT COUNT(*)
FROM update_container;

-- name: GetDeployment :one
SELECT d.*, m.language, m.name AS module_name
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
WHERE d.key = $1;

-- name: GetDeploymentsWithArtefacts :many
-- Get all deployments that have artefacts matching the given digests.
SELECT d.id, d.created_at, d.key, m.name
FROM deployments d
         INNER JOIN modules m ON d.module_id = m.id
WHERE EXISTS (SELECT 1
              FROM deployment_artefacts da
                       INNER JOIN artefacts a ON da.artefact_id = a.id
              WHERE a.digest = ANY (@digests::bytea[])
                AND da.deployment_id = d.id
              HAVING COUNT(*) = @count -- Number of unique digests provided
);

-- name: GetArtefactContentRange :one
SELECT SUBSTRING(a.content FROM @start FOR @count)::BYTEA AS content
FROM artefacts a
WHERE a.id = @id;

-- name: UpsertRunner :one
-- Upsert a runner and return the deployment ID that it is assigned to, if any.
WITH deployment_rel AS (
-- If the deployment key is null, then deployment_rel.id will be null,
-- otherwise we try to retrieve the deployments.id using the key. If
-- there is no corresponding deployment, then the deployment ID is -1
-- and the parent statement will fail due to a foreign key constraint.
    SELECT CASE
               WHEN sqlc.narg('deployment_key')::UUID IS NULL
                   THEN NULL
               ELSE COALESCE((SELECT id
                              FROM deployments d
                              WHERE d.key = sqlc.narg('deployment_key')
                              LIMIT 1), -1) END AS id)
INSERT
INTO runners (key, endpoint, state, labels, deployment_id, last_seen)
VALUES ($1, $2, $3, $4, (SELECT id FROM deployment_rel), NOW() AT TIME ZONE 'utc')
ON CONFLICT (key) DO UPDATE SET endpoint      = $2,
                                state         = $3,
                                labels        = $4,
                                deployment_id = (SELECT id FROM deployment_rel),
                                last_seen     = NOW() AT TIME ZONE 'utc'
RETURNING deployment_id;

-- name: KillStaleRunners :one
WITH matches AS (
    UPDATE runners
        SET state = 'dead'
        WHERE state <> 'dead' AND last_seen < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: DeregisterRunner :one
WITH matches AS (
    UPDATE runners
        SET state = 'dead'
        WHERE key = $1
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetActiveRunners :many
SELECT DISTINCT ON (r.key) r.key                                  AS runner_key,
                           r.endpoint,
                           r.state,
                           r.labels,
                           r.last_seen,
                           COALESCE(CASE
                                        WHEN r.deployment_id IS NOT NULL
                                            THEN d.key END, NULL) AS deployment_key
FROM runners r
         LEFT JOIN deployments d on d.id = r.deployment_id OR r.deployment_id IS NULL
WHERE sqlc.arg('all')::bool = true
   OR r.state <> 'dead'
ORDER BY r.key;

-- name: GetDeployments :many
SELECT d.*, m.name AS module_name, m.language
FROM deployments d
         INNER JOIN modules m on d.module_id = m.id
WHERE sqlc.arg('all')::bool = true
   OR min_replicas > 0
ORDER BY d.key;

-- name: GetIdleRunners :many
SELECT *
FROM runners
WHERE labels @> sqlc.arg('labels')::jsonb
  AND state = 'idle'
LIMIT sqlc.arg('limit');

-- name: SetDeploymentDesiredReplicas :exec
UPDATE deployments
SET min_replicas = $2
WHERE key = $1
RETURNING 1;

-- name: GetExistingDeploymentForModule :one
SELECT d.*
FROM deployments d
         INNER JOIN modules m on d.module_id = m.id
WHERE m.name = $1
  AND min_replicas > 0
LIMIT 1;

-- name: GetDeploymentsNeedingReconciliation :many
-- Get deployments that have a mismatch between the number of assigned and required replicas.
SELECT d.key                  AS key,
       m.name                 AS module_name,
       m.language             AS language,
       COUNT(r.id)            AS assigned_runners_count,
       d.min_replicas::BIGINT AS required_runners_count
FROM deployments d
         LEFT JOIN runners r ON d.id = r.deployment_id AND r.state <> 'dead'
         JOIN modules m ON d.module_id = m.id
GROUP BY d.key, d.min_replicas, m.name, m.language
HAVING COUNT(r.id) <> d.min_replicas;


-- name: ReserveRunner :one
-- Find an idle runner and reserve it for the given deployment.
UPDATE runners
SET state               = 'reserved',
    reservation_timeout = @reservation_timeout,
    -- If a deployment is not found, then the deployment ID is -1
    -- and the update will fail due to a FK constraint.
    deployment_id       = COALESCE((SELECT id
                                    FROM deployments d
                                    WHERE d.key = sqlc.arg('deployment_key')
                                    LIMIT 1), -1)
WHERE id = (SELECT id
            FROM runners r
            WHERE r.state = 'idle'
              AND r.labels @> sqlc.arg('labels')::jsonb
            LIMIT 1 FOR UPDATE SKIP LOCKED)
RETURNING runners.*;

-- name: GetRunnerState :one
SELECT state
FROM runners
WHERE key = $1;

-- name: GetRunner :one
SELECT DISTINCT ON (r.key) r.key                                  AS runner_key,
                           r.endpoint,
                           r.state,
                           r.labels,
                           r.last_seen,
                           COALESCE(CASE
                                        WHEN r.deployment_id IS NOT NULL
                                            THEN d.key END, NULL) AS deployment_key
FROM runners r
         LEFT JOIN deployments d on d.id = r.deployment_id OR r.deployment_id IS NULL
WHERE r.key = $1;

-- name: GetRoutingTable :many
SELECT endpoint, r.key AS runner_key, d.key deployment_key
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
         INNER JOIN modules m on d.module_id = m.id
WHERE state = 'assigned'
  AND m.name = $1;

-- name: GetRunnersForDeployment :many
SELECT r.*
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
WHERE state = 'assigned'
  AND d.key = $1;

-- name: ExpireRunnerReservations :one
WITH rows AS (
    UPDATE runners
        SET state = 'idle',
            deployment_id = NULL,
            reservation_timeout = NULL
        WHERE state = 'reserved'
            AND reservation_timeout < (NOW() AT TIME ZONE 'utc')
        RETURNING 1)
SELECT COUNT(*)
FROM rows;

-- name: InsertLogEvent :exec
INSERT INTO events (deployment_id, request_id, time_stamp, custom_key_1, type, payload)
VALUES ((SELECT id FROM deployments d WHERE d.key = sqlc.arg('deployment_key') LIMIT 1),
        (CASE
             WHEN sqlc.narg('request_key')::UUID IS NULL THEN NULL
             ELSE (SELECT id FROM ingress_requests ir WHERE ir.key = sqlc.narg('request_key')::UUID LIMIT 1)
            END),
        sqlc.arg('time_stamp')::TIMESTAMPTZ,
        sqlc.arg('level')::INT,
        'log',
        jsonb_build_object(
                'message', sqlc.arg('message')::TEXT,
                'attributes', sqlc.arg('attributes')::JSONB,
                'error', sqlc.narg('error')::TEXT
            ));

-- name: GetDeploymentLogs :many
SELECT ir.key                              AS request_key,
       ir.key                              AS deployment_key,
       e.time_stamp                        AS time_stamp,
       e.custom_key_1                      AS level,
       (e.payload ->> 'message')::TEXT     AS message,
       (e.payload ->> 'attributes')::JSONB AS attributes,
       (e.payload ->> 'error')::TEXT       AS error
FROM events e
         INNER JOIN ingress_requests ir ON e.request_id = ir.id
         INNER JOIN deployments d ON e.deployment_id = d.id
WHERE e.type = 'log'
  AND d.key = sqlc.arg('deployment_key')
  AND ir.key = sqlc.arg('request_key')
  AND e.time_stamp >= sqlc.arg('after_timestamp')
  AND e.time_stamp <= sqlc.arg('before_timestamp');

-- name: InsertCallEvent :exec
INSERT INTO events (deployment_id, request_id, time_stamp, type,
                    custom_key_1, custom_key_2, custom_key_3, custom_key_4, payload)
VALUES ((SELECT id FROM deployments WHERE deployments.key = sqlc.arg('deployment_key')::UUID),
        (CASE
             WHEN sqlc.narg('request_key')::UUID IS NULL THEN NULL
             ELSE (SELECT id FROM ingress_requests ir WHERE ir.key = sqlc.arg('request_key')::UUID)
            END),
        sqlc.arg('time_stamp')::TIMESTAMPTZ,
        'call',
        sqlc.arg('source_module')::TEXT,
        sqlc.arg('source_verb')::TEXT,
        sqlc.arg('dest_module')::TEXT,
        sqlc.arg('dest_verb')::TEXT,
        jsonb_build_object(
                'duration_ms', sqlc.arg('duration_ms')::BIGINT,
                'request', sqlc.arg('request')::JSONB,
                'response', sqlc.arg('response')::JSONB,
                'error', sqlc.narg('error')::TEXT
            ));

-- name: GetCalls :many
SELECT d.key          AS deployment_key,
       ir.key         AS request_key,
       e.custom_key_1 AS source_module,
       e.custom_key_2 AS source_verb,
       e.custom_key_3 AS dest_module,
       e.custom_key_4 AS dest_verb,
       e.payload      AS payload,
       e.time_stamp   AS time_stamp
FROM events e
         INNER JOIN deployments d ON e.deployment_id = d.id
         INNER JOIN ingress_requests ir ON e.request_id = ir.id
WHERE CASE
          WHEN sqlc.narg('request_key')::UUID IS NOT NULL
              THEN ir.key = sqlc.narg('request_key')::UUID
          WHEN sqlc.narg('dest_module')::TEXT[] IS NOT NULL
              THEN e.custom_key_3 = ANY (sqlc.narg('dest_module')::TEXT[])
          ELSE
              TRUE
    END
  AND type = 'call';

-- name: CreateIngressRequest :exec
INSERT INTO ingress_requests (key, source_addr)
VALUES ($1, $2);

-- name: UpsertController :one
INSERT INTO controller (key, endpoint)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET state     = 'live',
                                endpoint  = $2,
                                last_seen = NOW() AT TIME ZONE 'utc'
RETURNING id;

-- name: KillStaleControllers :one
-- Mark any controller entries that haven't been updated recently as dead.
WITH matches AS (
    UPDATE controller
        SET state = 'dead'
        WHERE state <> 'dead' AND last_seen < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetControllers :many
SELECT c.*
FROM controller c
WHERE sqlc.arg('all')::bool = true
   OR c.state <> 'dead'
ORDER BY c.key;

-- name: CreateIngressRoute :exec
INSERT INTO ingress_routes (deployment_id, module, verb, method, path)
VALUES ((SELECT id FROM deployments WHERE key = $1 LIMIT 1), $2, $3, $4, $5);

-- name: GetIngressRoutes :many
-- Get the runner endpoints corresponding to the given ingress route.
SELECT r.key AS runner_key, endpoint, ir.module, ir.verb
FROM ingress_routes ir
         INNER JOIN runners r ON ir.deployment_id = r.deployment_id
WHERE r.state = 'assigned'
  AND ir.method = $1
  AND ir.path = $2;

-- name: GetAllIngressRoutes :many
SELECT d.key AS deployment_key, ir.module, ir.verb, ir.method, ir.path
FROM ingress_routes ir
         INNER JOIN deployments d ON ir.deployment_id = d.id
WHERE sqlc.arg('all')::bool = true
   OR d.min_replicas > 0;


-- name: InsertEvent :exec
INSERT INTO events (deployment_id, request_id, type,
                    custom_key_1, custom_key_2, custom_key_3, custom_key_4,
                    payload)
VALUES ($1, $2, $3, $4, $4, $5, $6, $7);

-- name: GetEvents :many
SELECT d.key        AS deployment_key,
       ir.key       AS request_key,
       e.time_stamp AS time_stamp,
       e.type       AS type,
       e.payload    AS payload
FROM events e
         INNER JOIN deployments d on e.deployment_id = d.id
         INNER JOIN ingress_requests ir on e.request_id = ir.id
WHERE time_stamp >= sqlc.arg('after_timestamp')
  AND time_stamp <= COALESCE(sqlc.narg('before_timestamp'), NOW() AT TIME ZONE 'utc')
  AND d.key = sqlc.arg('deployment_key')
  AND ir.key = sqlc.arg('request_key')
  AND e.type = ANY (sqlc.arg('types')::TEXT[]);