-- migrate:up
CREATE
    EXTENSION IF NOT EXISTS pgcrypto;

-- Function for deployment notifications.
CREATE OR REPLACE FUNCTION notify_deployment_event() RETURNS TRIGGER AS
$$
DECLARE
    payload JSONB;
BEGIN
    IF TG_OP = 'DELETE'
    THEN
        payload = jsonb_build_object(
                'table', TG_TABLE_NAME,
                'action', TG_OP,
                'old', old.key
            );
    ELSE
        payload = jsonb_build_object(
                'table', TG_TABLE_NAME,
                'action', TG_OP,
                'new', new.key
            );
    END IF;
    PERFORM pg_notify('notify_events', payload::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE modules
(
    id       BIGINT         NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    language VARCHAR        NOT NULL,
    name     VARCHAR UNIQUE NOT NULL
);

-- Proto-encoded module schema.
CREATE DOMAIN module_schema_pb AS BYTEA;

CREATE DOMAIN runner_key AS varchar;
CREATE DOMAIN controller_key AS varchar;
CREATE DOMAIN deployment_key AS varchar;

CREATE TABLE deployments
(
    id           BIGINT         NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    module_id    BIGINT         NOT NULL REFERENCES modules (id) ON DELETE CASCADE,
    -- Unique key for this deployment in the form <module-name>-<random>.
    "key"       deployment_key UNIQUE NOT NULL,
    "schema"     module_schema_pb  NOT NULL,
    -- Labels are used to match deployments to runners.
    "labels"     JSONB          NOT NULL DEFAULT '{}',
    min_replicas INT            NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX deployments_key_idx ON deployments (key);
CREATE INDEX deployments_module_id_idx ON deployments (module_id);
-- Only allow one deployment per module.
CREATE UNIQUE INDEX deployments_unique_idx ON deployments (module_id)
    WHERE min_replicas > 0;

CREATE TRIGGER deployments_notify_event
    AFTER INSERT OR UPDATE OR DELETE
    ON deployments
    FOR EACH ROW
EXECUTE PROCEDURE notify_deployment_event();

CREATE TABLE artefacts
(
    id         BIGINT       NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    -- SHA256 digest of the content.
    digest     BYTEA UNIQUE NOT NULL,
    content    BYTEA        NOT NULL
);

CREATE UNIQUE INDEX artefacts_digest_idx ON artefacts (digest);

CREATE TABLE deployment_artefacts
(
    artefact_id   BIGINT      NOT NULL REFERENCES artefacts (id) ON DELETE CASCADE,
    deployment_id BIGINT      NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    executable    BOOLEAN     NOT NULL,
    -- Path relative to the module root.
    path          VARCHAR     NOT NULL
);

CREATE INDEX deployment_artefacts_deployment_id_idx ON deployment_artefacts (deployment_id);

CREATE TYPE runner_state AS ENUM (
    -- The Runner is available to run deployments.
    'idle',
    -- The Runner is reserved but has not yet deployed.
    'reserved',
    -- The Runner has been assigned a deployment.
    'assigned',
    -- The Runner is dead.
    'dead'
    );

-- Runners are processes that are available to run modules.
CREATE TABLE runners
(
    id                  BIGINT       NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    -- Unique identifier for this runner, generated at startup.
    key                 runner_key UNIQUE  NOT NULL,
    created             TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    last_seen           TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    -- If the runner is reserved, this is the time at which the reservation expires.
    reservation_timeout TIMESTAMPTZ,
    state               runner_state NOT NULL DEFAULT 'idle',
    endpoint            VARCHAR      NOT NULL,
    -- Some denormalisation for performance. Without this we need to do a two table join.
    module_name         VARCHAR,
    deployment_id       BIGINT       REFERENCES deployments (id) ON DELETE SET NULL,
    labels              JSONB        NOT NULL DEFAULT '{}'
);

-- Automatically update module_name when deployment_id is set or unset.
CREATE OR REPLACE FUNCTION runners_update_module_name() RETURNS TRIGGER AS
$$
BEGIN
    IF NEW.deployment_id IS NULL
    THEN
        NEW.module_name = NULL;
    ELSE
        SELECT m.name
        INTO NEW.module_name
        FROM modules m
                 INNER JOIN deployments d on m.id = d.module_id
        WHERE d.id = NEW.deployment_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER runners_update_module_name
    BEFORE INSERT OR UPDATE
    ON runners
    FOR EACH ROW
EXECUTE PROCEDURE runners_update_module_name();

-- Set a default reservation_timeout when a runner is reserved.
CREATE OR REPLACE FUNCTION runners_set_reservation_timeout() RETURNS TRIGGER AS
$$
BEGIN
    IF OLD.state != 'reserved' AND NEW.state = 'reserved' AND NEW.reservation_timeout IS NULL
    THEN
        NEW.reservation_timeout = NOW() AT TIME ZONE 'utc' + INTERVAL '2 minutes';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER runners_set_reservation_timeout
    BEFORE INSERT OR UPDATE
    ON runners
    FOR EACH ROW
EXECUTE PROCEDURE runners_set_reservation_timeout();

CREATE UNIQUE INDEX runners_key ON runners (key);
CREATE UNIQUE INDEX runners_endpoint_not_dead_idx ON runners (endpoint) WHERE state <> 'dead';
CREATE INDEX runners_module_name_idx ON runners (module_name);
CREATE INDEX runners_state_idx ON runners (state);
CREATE INDEX runners_deployment_id_idx ON runners (deployment_id);
CREATE INDEX runners_labels_idx ON runners USING GIN (labels);

CREATE TABLE ingress_routes
(
    method        VARCHAR NOT NULL,
    path          VARCHAR NOT NULL,
    -- The deployment that should handle this route.
    deployment_id BIGINT  NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    -- Duplicated here to avoid having to join from this to deployments then modules.
    module        VARCHAR NOT NULL,
    verb          VARCHAR NOT NULL
);

CREATE INDEX ingress_routes_method_path_idx ON ingress_routes (method, path);

CREATE TYPE origin AS ENUM (
    'ingress',
    -- Not supported yet.
    'cron',
    'pubsub'
    );

-- Requests originating from outside modules, either from external sources or from
-- events within the system.
CREATE TABLE requests
(
    id          BIGINT         NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    origin      origin         NOT NULL,
    -- Will be in the form <origin>-<description>-<hash>:
    --
    -- ingress: ingress-<method>-<path>-<hash> (eg. ingress-GET-foo-bar-<hash>)
    -- cron: cron-<name>-<hash>                (eg. cron-poll-news-sources-<hash>)
    -- pubsub: pubsub-<subscription>-<hash>    (eg. pubsub-articles-<hash>)
    name        VARCHAR UNIQUE NOT NULL,
    source_addr VARCHAR        NOT NULL
);

CREATE INDEX requests_origin_idx ON requests (origin);
CREATE UNIQUE INDEX ingress_requests_key_idx ON requests (name);

CREATE TYPE controller_state AS ENUM (
    'live',
    'dead'
    );

CREATE TABLE controller
(
    id        BIGINT           NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    key       controller_key UNIQUE      NOT NULL,
    created   TIMESTAMPTZ      NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    last_seen TIMESTAMPTZ      NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    state     controller_state NOT NULL DEFAULT 'live',
    endpoint  VARCHAR          NOT NULL
);

CREATE UNIQUE INDEX controller_endpoint_not_dead_idx ON controller (endpoint) WHERE state <> 'dead';

CREATE TYPE event_type AS ENUM (
    'call',
    'log',
    'deployment_created',
    'deployment_updated'
    );

CREATE TABLE events
(
    id            BIGINT      NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    time_stamp    TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

    deployment_id BIGINT      NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    request_id    BIGINT      NULL REFERENCES requests (id) ON DELETE CASCADE,

    type          event_type  NOT NULL,

    -- Type-specific keys used to index events for searching.
    custom_key_1  VARCHAR     NULL,
    custom_key_2  VARCHAR     NULL,
    custom_key_3  VARCHAR     NULL,
    custom_key_4  VARCHAR     NULL,

    payload       JSONB       NOT NULL
);

CREATE INDEX events_timestamp_idx ON events (time_stamp);
CREATE INDEX events_deployment_id_idx ON events (deployment_id);
CREATE INDEX events_request_id_idx ON events (request_id);
CREATE INDEX events_type_idx ON events (type);
CREATE INDEX events_custom_key_1_idx ON events (custom_key_1);
CREATE INDEX events_custom_key_2_idx ON events (custom_key_2);
CREATE INDEX events_custom_key_3_idx ON events (custom_key_3);
CREATE INDEX events_custom_key_4_idx ON events (custom_key_4);

-- migrate:down