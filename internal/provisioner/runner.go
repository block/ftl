package provisioner

import (
	"context"
	"fmt"
	"reflect"

	"github.com/alecthomas/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/provisioner/state"
)

type Handler struct {
	Executor Executor
	Handles  []state.State
}

type RunnerStage struct {
	Name     string
	Handlers []Handler
}

// Task is an async running provisioner task.
type Task struct {
	runner *Runner

	err     atomic.Value[error]
	outputs atomic.Value[[]state.State]
}

// Runner runs a set of provision handlers on a set of
// input states.
//
// It returns the final set of states after all handlers have been executed.
type Runner struct {
	skipped []state.State // states that did not have executors at the current stage

	State  []state.State
	Stages []RunnerStage
}

func (r *Runner) Run(ctx context.Context) ([]state.State, error) {
	logger := log.FromContext(ctx)

	for _, stage := range r.Stages {
		logger.Debugf("running stage %s", stage.Name)

		if err := r.prepare(ctx, &stage); err != nil {
			return nil, err
		}

		newStates, err := r.execute(ctx, &stage)
		if err != nil {
			return nil, err
		}

		r.State = newStates
	}
	logger.Debugf("runner finished")
	return r.State, nil
}

func (r *Runner) AsyncTask() *Task {
	return &Task{runner: r}
}

func (r *Runner) prepare(ctx context.Context, stage *RunnerStage) error {
	for _, state := range r.State {
		found := false
	handlerLoop:
		for _, handler := range stage.Handlers {
			for _, resource := range handler.Handles {
				if reflect.TypeOf(resource) == reflect.TypeOf(state) {
					if err := handler.Executor.Prepare(ctx, state); err != nil {
						return fmt.Errorf("failed to prepare executor: %w", err)
					}
					found = true
					break handlerLoop
				}
			}
		}
		if !found {
			r.skipped = append(r.skipped, state)
		}
	}
	return nil
}

func (r *Runner) execute(ctx context.Context, stage *RunnerStage) ([]state.State, error) {
	reschan := make(chan []state.State, len(stage.Handlers))
	eg := errgroup.Group{}
	for _, handler := range stage.Handlers {
		eg.Go(func() error {
			outputs, err := handler.Executor.Execute(ctx)
			if err != nil {
				return fmt.Errorf("failed to execute handler %T: %w", handler.Executor, err)
			}
			reschan <- outputs
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("%T execution failed: %w", r, err)
	}

	result := r.skipped
	for range len(stage.Handlers) {
		states := <-reschan
		result = append(result, states...)
	}

	return result, nil
}

func (t *Task) Start(oldCtx context.Context) {
	ctx := context.WithoutCancel(oldCtx)
	logger := log.FromContext(ctx)
	go func() {
		outputs, err := t.runner.Run(ctx)
		if err != nil {
			logger.Errorf(err, "failed to execute provisioner")
			t.err.Store(err)
			return
		}
		t.outputs.Store(outputs)
	}()
}

func (t *Task) Err() error {
	if err := t.err.Load(); err != nil {
		return fmt.Errorf("failed to execute provisioner: %w", err)
	}
	return nil
}

func (t *Task) Outputs() []state.State {
	return t.outputs.Load()
}
