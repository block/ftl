//protobuf:package xyz.block.ftl.cron.v1
//protobuf:option go_package="github.com/block/ftl/backend/protos/xyz/block/ftl/cron/v1;cronpb"
package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	cronpb "github.com/block/ftl/backend/protos/xyz/block/ftl/cron/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/statemachine"
	"google.golang.org/protobuf/proto"
)

// CronState is the state of scheduled cron jobs
//
//protobuf:export
type CronState struct {
	// Map of job key to last execution time
	LastExecutions map[string]time.Time `protobuf:"1"`
	// Map of job key to next scheduled time
	NextExecutions map[string]time.Time `protobuf:"2"`
}

func (v *CronState) Marshal() ([]byte, error) {
	stateProto := v.ToProto()
	bytes, err := proto.Marshal(stateProto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema state: %w", err)
	}
	return bytes, nil
}

func (v *CronState) Unmarshal(data []byte) error {
	stateProto := &cronpb.CronState{}
	if err := proto.Unmarshal(data, stateProto); err != nil {
		return fmt.Errorf("failed to unmarshal cron state: %w", err)
	}
	out, err := CronStateFromProto(stateProto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cron state: %w", err)
	}
	*v = *out
	return nil
}

// CronEvent represents changes to the cron state
type CronEvent struct {
	// Job that was executed
	JobKey string
	// When the job was executed
	ExecutedAt time.Time
	// Next scheduled execution
	NextExecution time.Time
}

func (e CronEvent) MarshalBinary() ([]byte, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cron event: %w", err)
	}
	return bytes, nil
}

func (e *CronEvent) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, e); err != nil {
		return fmt.Errorf("failed to unmarshal cron event: %w", err)
	}
	return nil
}

// Handle applies the event to the view
func (e CronEvent) Handle(state CronState) (CronState, error) {
	if state.LastExecutions == nil {
		state.LastExecutions = make(map[string]time.Time)
	}
	if state.NextExecutions == nil {
		state.NextExecutions = make(map[string]time.Time)
	}

	state.LastExecutions[e.JobKey] = e.ExecutedAt
	state.NextExecutions[e.JobKey] = e.NextExecution
	return state, nil
}

// cronStateMachine implements a state machine for tracking cron job executions
type cronStateMachine struct {
	state      CronState
	notifier   *channels.Notifier
	mu         sync.Mutex
	runningCtx context.Context
}

var _ statemachine.Snapshotting[struct{}, CronState, CronEvent] = &cronStateMachine{}
var _ statemachine.Listenable[struct{}, CronState, CronEvent] = &cronStateMachine{}

func newStateMachine(ctx context.Context) *cronStateMachine {
	return &cronStateMachine{
		state: CronState{
			LastExecutions: make(map[string]time.Time),
			NextExecutions: make(map[string]time.Time),
		},
		notifier:   channels.NewNotifier(ctx),
		runningCtx: ctx,
	}
}

func (s *cronStateMachine) Lookup(key struct{}) (CronState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return reflect.DeepCopy(s.state), nil
}

func (s *cronStateMachine) Publish(event CronEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := log.FromContext(s.runningCtx)

	var err error
	s.state, err = event.Handle(s.state)
	if err != nil {
		logger.Errorf(err, "failed to apply event")
		return nil
	}
	s.notifier.Notify(s.runningCtx)
	return nil
}

func (s *cronStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return s.notifier.Subscribe(), nil
}

func (s *cronStateMachine) Close() error {
	return nil
}

func (s *cronStateMachine) Recover(snapshot io.Reader) error {
	snapshotBytes, err := io.ReadAll(snapshot)
	if err != nil {
		return fmt.Errorf("failed to read snapshot: %w", err)
	}
	if err := s.state.Unmarshal(snapshotBytes); err != nil {
		return fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return nil
}

func (s *cronStateMachine) Save(w io.Writer) error {
	snapshotBytes, err := s.state.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}
	_, err = w.Write(snapshotBytes)
	if err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}
	return nil
}
