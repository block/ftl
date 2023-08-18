// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package sql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/internal/sqltypes"
	"github.com/alecthomas/types"
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

type EventType string

const (
	EventTypeCall       EventType = "call"
	EventTypeLog        EventType = "log"
	EventTypeDeployment EventType = "deployment"
)

func (e *EventType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EventType(s)
	case string:
		*e = EventType(s)
	default:
		return fmt.Errorf("unsupported scan type for EventType: %T", src)
	}
	return nil
}

type NullEventType struct {
	EventType EventType
	Valid     bool // Valid is true if EventType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEventType) Scan(value interface{}) error {
	if value == nil {
		ns.EventType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EventType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEventType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EventType), nil
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
	CreatedAt time.Time
	Digest    []byte
	Content   []byte
}

type Controller struct {
	ID       int64
	Key      model.ControllerKey
	Created  time.Time
	LastSeen time.Time
	State    ControllerState
	Endpoint string
}

type Deployment struct {
	ID          int64
	CreatedAt   time.Time
	ModuleID    int64
	Name        model.DeploymentName
	Schema      []byte
	Labels      []byte
	MinReplicas int32
}

type DeploymentArtefact struct {
	ArtefactID   int64
	DeploymentID int64
	CreatedAt    time.Time
	Executable   bool
	Path         string
}

type Event struct {
	TimeStamp    time.Time
	DeploymentID int64
	RequestID    types.Option[int64]
	Type         EventType
	CustomKey1   types.Option[string]
	CustomKey2   types.Option[string]
	CustomKey3   types.Option[string]
	CustomKey4   types.Option[string]
	Payload      json.RawMessage
}

type IngressRequest struct {
	ID         int64
	Key        model.IngressRequestKey
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
	Key                model.RunnerKey
	Created            time.Time
	LastSeen           time.Time
	ReservationTimeout sqltypes.NullTime
	State              RunnerState
	Endpoint           string
	DeploymentID       types.Option[int64]
	Labels             []byte
}
