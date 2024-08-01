// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package sql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
	"github.com/google/uuid"
)

type AsyncCallState string

const (
	AsyncCallStatePending   AsyncCallState = "pending"
	AsyncCallStateExecuting AsyncCallState = "executing"
	AsyncCallStateSuccess   AsyncCallState = "success"
	AsyncCallStateError     AsyncCallState = "error"
)

func (e *AsyncCallState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AsyncCallState(s)
	case string:
		*e = AsyncCallState(s)
	default:
		return fmt.Errorf("unsupported scan type for AsyncCallState: %T", src)
	}
	return nil
}

type NullAsyncCallState struct {
	AsyncCallState AsyncCallState
	Valid          bool // Valid is true if AsyncCallState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAsyncCallState) Scan(value interface{}) error {
	if value == nil {
		ns.AsyncCallState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AsyncCallState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAsyncCallState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AsyncCallState), nil
}

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

type CronJobState string

const (
	CronJobStateIdle      CronJobState = "idle"
	CronJobStateExecuting CronJobState = "executing"
)

func (e *CronJobState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = CronJobState(s)
	case string:
		*e = CronJobState(s)
	default:
		return fmt.Errorf("unsupported scan type for CronJobState: %T", src)
	}
	return nil
}

type NullCronJobState struct {
	CronJobState CronJobState
	Valid        bool // Valid is true if CronJobState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullCronJobState) Scan(value interface{}) error {
	if value == nil {
		ns.CronJobState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.CronJobState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullCronJobState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.CronJobState), nil
}

type EventType string

const (
	EventTypeCall              EventType = "call"
	EventTypeLog               EventType = "log"
	EventTypeDeploymentCreated EventType = "deployment_created"
	EventTypeDeploymentUpdated EventType = "deployment_updated"
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

type FsmStatus string

const (
	FsmStatusRunning   FsmStatus = "running"
	FsmStatusCompleted FsmStatus = "completed"
	FsmStatusFailed    FsmStatus = "failed"
)

func (e *FsmStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = FsmStatus(s)
	case string:
		*e = FsmStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for FsmStatus: %T", src)
	}
	return nil
}

type NullFsmStatus struct {
	FsmStatus FsmStatus
	Valid     bool // Valid is true if FsmStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullFsmStatus) Scan(value interface{}) error {
	if value == nil {
		ns.FsmStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.FsmStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullFsmStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.FsmStatus), nil
}

type Origin string

const (
	OriginIngress Origin = "ingress"
	OriginCron    Origin = "cron"
	OriginPubsub  Origin = "pubsub"
)

func (e *Origin) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Origin(s)
	case string:
		*e = Origin(s)
	default:
		return fmt.Errorf("unsupported scan type for Origin: %T", src)
	}
	return nil
}

type NullOrigin struct {
	Origin Origin
	Valid  bool // Valid is true if Origin is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrigin) Scan(value interface{}) error {
	if value == nil {
		ns.Origin, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Origin.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrigin) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Origin), nil
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

type TopicSubscriptionState string

const (
	TopicSubscriptionStateIdle      TopicSubscriptionState = "idle"
	TopicSubscriptionStateExecuting TopicSubscriptionState = "executing"
)

func (e *TopicSubscriptionState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TopicSubscriptionState(s)
	case string:
		*e = TopicSubscriptionState(s)
	default:
		return fmt.Errorf("unsupported scan type for TopicSubscriptionState: %T", src)
	}
	return nil
}

type NullTopicSubscriptionState struct {
	TopicSubscriptionState TopicSubscriptionState
	Valid                  bool // Valid is true if TopicSubscriptionState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTopicSubscriptionState) Scan(value interface{}) error {
	if value == nil {
		ns.TopicSubscriptionState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TopicSubscriptionState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTopicSubscriptionState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TopicSubscriptionState), nil
}

type Artefact struct {
	ID        int64
	CreatedAt time.Time
	Digest    []byte
	Content   []byte
}

type AsyncCall struct {
	ID                int64
	CreatedAt         time.Time
	LeaseID           optional.Option[int64]
	Verb              schema.RefKey
	State             AsyncCallState
	Origin            string
	ScheduledAt       time.Time
	Request           []byte
	Response          []byte
	Error             optional.Option[string]
	RemainingAttempts int32
	Backoff           time.Duration
	MaxBackoff        time.Duration
}

type Controller struct {
	ID       int64
	Key      model.ControllerKey
	Created  time.Time
	LastSeen time.Time
	State    ControllerState
	Endpoint string
}

type CronJob struct {
	ID            int64
	Key           model.CronJobKey
	DeploymentID  int64
	Verb          string
	Schedule      string
	StartTime     time.Time
	NextExecution time.Time
	State         model.CronJobState
	ModuleName    string
}

type Deployment struct {
	ID          int64
	CreatedAt   time.Time
	ModuleID    int64
	Key         model.DeploymentKey
	Schema      *schema.Module
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
	ID           int64
	TimeStamp    time.Time
	DeploymentID int64
	RequestID    optional.Option[int64]
	Type         EventType
	CustomKey1   optional.Option[string]
	CustomKey2   optional.Option[string]
	CustomKey3   optional.Option[string]
	CustomKey4   optional.Option[string]
	Payload      json.RawMessage
}

type FsmInstance struct {
	ID               int64
	CreatedAt        time.Time
	Fsm              schema.RefKey
	Key              string
	Status           FsmStatus
	CurrentState     optional.Option[schema.RefKey]
	DestinationState optional.Option[schema.RefKey]
	AsyncCallID      optional.Option[int64]
	UpdatedAt        time.Time
}

type IngressRoute struct {
	Method       string
	Path         string
	DeploymentID int64
	Module       string
	Verb         string
}

type Lease struct {
	ID             int64
	IdempotencyKey uuid.UUID
	Key            leases.Key
	CreatedAt      time.Time
	ExpiresAt      time.Time
	Metadata       []byte
}

type Module struct {
	ID       int64
	Language string
	Name     string
}

type ModuleConfiguration struct {
	ID        int64
	CreatedAt time.Time
	Module    optional.Option[string]
	Name      string
	Value     []byte
}

type ModuleSecret struct {
	ID        int64
	CreatedAt time.Time
	Module    optional.Option[string]
	Name      string
	Url       string
}

type Request struct {
	ID         int64
	Origin     Origin
	Key        model.RequestKey
	SourceAddr string
}

type Runner struct {
	ID                 int64
	Key                model.RunnerKey
	Created            time.Time
	LastSeen           time.Time
	ReservationTimeout optional.Option[time.Time]
	State              RunnerState
	Endpoint           string
	ModuleName         optional.Option[string]
	DeploymentID       optional.Option[int64]
	Labels             []byte
}

type Topic struct {
	ID        int64
	Key       model.TopicKey
	CreatedAt time.Time
	ModuleID  int64
	Name      string
	Type      string
	Head      optional.Option[int64]
}

type TopicEvent struct {
	ID        int64
	CreatedAt time.Time
	Key       model.TopicEventKey
	TopicID   int64
	Payload   []byte
	Caller    string
}

type TopicSubscriber struct {
	ID                   int64
	Key                  model.SubscriberKey
	CreatedAt            time.Time
	TopicSubscriptionsID int64
	DeploymentID         int64
	Sink                 schema.RefKey
	RetryAttempts        int32
	Backoff              time.Duration
	MaxBackoff           time.Duration
}

type TopicSubscription struct {
	ID           int64
	Key          model.SubscriptionKey
	CreatedAt    time.Time
	TopicID      int64
	ModuleID     int64
	DeploymentID int64
	Name         string
	Cursor       optional.Option[int64]
	State        TopicSubscriptionState
}
