package state

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/key"
)

type Runner struct {
	Key        key.Runner
	Create     time.Time
	LastSeen   time.Time
	Endpoint   string
	Module     string
	Deployment key.Deployment
}

func (r *RunnerState) Runner(s key.Runner) optional.Option[Runner] {
	result, ok := r.runners[s]
	if ok {
		return optional.Ptr(result)
	}
	return optional.None[Runner]()
}

func (r *RunnerState) Runners() []Runner {
	var ret []Runner
	for _, v := range r.runners {
		ret = append(ret, *v)
	}
	return ret
}

func (r *RunnerState) RunnersForDeployment(deployment key.Deployment) []Runner {
	var ret []Runner
	for _, v := range r.runnersByDeployment[deployment] {
		ret = append(ret, *v)
	}
	return ret
}

var _ RunnerEvent = (*RunnerRegisteredEvent)(nil)
var _ eventstream.VerboseMessage = (*RunnerRegisteredEvent)(nil)
var _ RunnerEvent = (*RunnerDeletedEvent)(nil)

type RunnerRegisteredEvent struct {
	Key        key.Runner
	Time       time.Time
	Endpoint   string
	Module     string
	Deployment key.Deployment
}

func (r *RunnerRegisteredEvent) VerboseMessage() {
	// Stops this message being logged every second
}

func (r *RunnerRegisteredEvent) Handle(t RunnerState) (RunnerState, error) {
	if existing := t.runners[r.Key]; existing != nil {
		existing.LastSeen = r.Time
		return t, nil
	}
	n := Runner{
		Key:        r.Key,
		Create:     r.Time,
		LastSeen:   r.Time,
		Endpoint:   r.Endpoint,
		Module:     r.Module,
		Deployment: r.Deployment,
	}
	t.runners[r.Key] = &n
	t.runnersByDeployment[r.Deployment] = append(t.runnersByDeployment[r.Deployment], &n)
	return t, nil
}

type RunnerDeletedEvent struct {
	Key key.Runner
}

func (r *RunnerDeletedEvent) Handle(t RunnerState) (RunnerState, error) {
	existing := t.runners[r.Key]
	if existing != nil {
		delete(t.runners, r.Key)
	}
	return t, nil
}
