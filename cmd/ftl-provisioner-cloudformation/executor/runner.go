package executor

import (
	"context"
	"fmt"
	"reflect"

	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/internal/log"
)

type Handler struct {
	Executor Executor
	Handles  []State
}

type RunnerStage struct {
	Name      string
	Executors []Handler
}

type ProvisionRunner struct {
	skipped []State // states that did not have executors at the current stage

	CurrentState []State
	Stages       []RunnerStage
}

func (r *ProvisionRunner) Run(ctx context.Context) ([]State, error) {
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

		r.CurrentState = newStates
	}
	logger.Debugf("runner finished")
	return r.CurrentState, nil
}

func (r *ProvisionRunner) prepare(ctx context.Context, stage *RunnerStage) error {
	for _, state := range r.CurrentState {
		found := false
	executorLoop:
		for _, executor := range stage.Executors {
			for _, resource := range executor.Handles {
				if reflect.TypeOf(resource) == reflect.TypeOf(state) {
					if err := executor.Executor.Prepare(ctx, state); err != nil {
						return fmt.Errorf("failed to prepare executor: %w", err)
					}
					found = true
					break executorLoop
				}
			}
		}
		if !found {
			r.skipped = append(r.skipped, state)
		}
	}
	return nil
}

func (r *ProvisionRunner) execute(ctx context.Context, stage *RunnerStage) ([]State, error) {
	reschan := make(chan []State, len(stage.Executors))
	eg := errgroup.Group{}
	for _, executor := range stage.Executors {
		eg.Go(func() error {
			outputs, err := executor.Executor.Execute(ctx)
			if err != nil {
				return fmt.Errorf("failed to execute executor %T: %w", executor.Executor, err)
			}
			reschan <- outputs
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("%T execution failed: %w", r, err)
	}

	result := r.skipped
	for range len(stage.Executors) {
		states := <-reschan
		result = append(result, states...)
	}

	return result, nil
}
