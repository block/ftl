// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package sql

import (
	"database/sql/driver"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/internal/sqltypes"
	"github.com/jackc/pgx/v5/pgtype"
)

type ControllerState string

const (
	ControllerStateLive ControllerState = "live"
	ControllerStateDead ControllerState = "dead"
)

func (e *ControllerState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ControllerState(s)
	case string:
		*e = ControllerState(s)
	default:
		return fmt.Errorf("unsupported scan type for ControllerState: %T", src)
	}
	return nil
}

type NullControllerState struct {
	ControllerState ControllerState
	Valid           bool // Valid is true if ControllerState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullControllerState) Scan(value interface{}) error {
	if value == nil {
		ns.ControllerState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ControllerState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullControllerState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ControllerState), nil
}

type RunnerState string

const (
	RunnerStateIdle     RunnerState = "idle"
	RunnerStateReserved RunnerState = "reserved"
	RunnerStateAssigned RunnerState = "assigned"
	RunnerStateDead     RunnerState = "dead"
)

func (e *RunnerState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RunnerState(s)
	case string:
		*e = RunnerState(s)
	default:
		return fmt.Errorf("unsupported scan type for RunnerState: %T", src)
	}
	return nil
}

type NullRunnerState struct {
	RunnerState RunnerState
	Valid       bool // Valid is true if RunnerState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRunnerState) Scan(value interface{}) error {
	if value == nil {
		ns.RunnerState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RunnerState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRunnerState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RunnerState), nil
}

type Artefact struct {
	ID        int64
	CreatedAt pgtype.Timestamptz
	Digest    []byte
	Content   []byte
}

type Call struct {
	ID           int64
	RequestID    int64
	RunnerID     int64
	ControllerID int64
	Time         pgtype.Timestamptz
	DestModule   string
	DestVerb     string
	SourceModule string
	SourceVerb   string
	DurationMs   int64
	Request      []byte
	Response     []byte
	Error        pgtype.Text
}

type Controller struct {
	ID       int64
	Key      sqltypes.Key
	Created  pgtype.Timestamptz
	LastSeen pgtype.Timestamptz
	State    ControllerState
	Endpoint string
}

type Deployment struct {
	ID          int64
	CreatedAt   pgtype.Timestamptz
	ModuleID    int64
	Key         sqltypes.Key
	Schema      []byte
	MinReplicas int32
}

type DeploymentArtefact struct {
	ArtefactID   int64
	DeploymentID int64
	CreatedAt    pgtype.Timestamptz
	Executable   bool
	Path         string
}

type DeploymentLog struct {
	ID           int64
	DeploymentID int64
	RunnerID     int64
	TimeStamp    pgtype.Timestamptz
	Level        int32
	Attributes   []byte
	Message      string
	Error        pgtype.Text
}

type IngressRequest struct {
	ID         int64
	Key        sqltypes.Key
	SourceAddr string
}

type IngressRoute struct {
	Method       string
	Path         string
	DeploymentID int64
	Module       string
	Verb         string
}

type Module struct {
	ID       int64
	Language string
	Name     string
}

type Runner struct {
	ID                 int64
	Key                sqltypes.Key
	Created            pgtype.Timestamptz
	LastSeen           pgtype.Timestamptz
	ReservationTimeout pgtype.Timestamptz
	State              RunnerState
	Languages          sqltypes.Languages
	Endpoint           string
	DeploymentID       pgtype.Int8
}
