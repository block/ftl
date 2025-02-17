package executor

import (
	"context"
	"fmt"

	"github.com/block/ftl/internal/log"
	"golang.org/x/sync/errgroup"
)

type ProvisionRunner struct {
	stage   Stage
	skipped []State // states that did not have executors at the current stage

	CurrentState []State
	Executors    []Executor
}

func (r *ProvisionRunner) Run(ctx context.Context) ([]State, error) {
	logger := log.FromContext(ctx)

	for {
		if r.stage >= StageDone {
			logger.Debugf("runner finished")
			return r.finalResult()
		}

		logger.Debugf("running stage %s", r.stage)

		executors, err := r.executorMap()
		if err != nil {
			return nil, err
		}

		if err := r.prepare(ctx, executors); err != nil {
			return nil, err
		}

		newStates, err := r.execute(ctx, executors)
		if err != nil {
			return nil, err
		}

		r.CurrentState = newStates
		r.stage++
	}
}

func (r *ProvisionRunner) prepare(ctx context.Context, executors map[ResourceKind]Executor) error {
	for _, state := range r.CurrentState {
		executor, ok := executors[state.Kind()]
		if !ok {
			r.skipped = append(r.skipped, state)
			continue
		}

		if err := executor.Prepare(ctx, state); err != nil {
			return fmt.Errorf("failed to prepare executor: %w", err)
		}
	}
	return nil
}

func (r *ProvisionRunner) execute(ctx context.Context, executors map[ResourceKind]Executor) ([]State, error) {
	reschan := make(chan []State, len(executors))
	eg := errgroup.Group{}
	for _, executor := range executors {
		eg.Go(func() error {
			outputs, err := executor.Execute(ctx)
			if err != nil {
				return fmt.Errorf("failed to execute executor: %w", err)
			}
			reschan <- outputs
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to execute executors: %w", err)
	}

	result := r.skipped
	for range len(executors) {
		states := <-reschan
		result = append(result, states...)
	}

	return result, nil
}

func (r *ProvisionRunner) finalResult() ([]State, error) {
	for _, state := range r.CurrentState {
		if state.Stage() != StageDone {
			// if the provisioner is misconfigured, not all states might be done
			return nil, fmt.Errorf("state %s is not done", state.DebugString())
		}
	}

	return r.CurrentState, nil
}

func (r *ProvisionRunner) executorMap() (map[ResourceKind]Executor, error) {
	result := make(map[ResourceKind]Executor)

	for _, state := range r.CurrentState {
		var defaultExecutor Executor
		found := false
		for _, executor := range r.Executors {
			if executor.Stage() != r.stage {
				continue
			}

			if len(executor.Resources()) == 0 {
				defaultExecutor = executor
				continue
			}

			for _, resource := range executor.Resources() {
				if resource == state.Kind() {
					if found {
						return nil, fmt.Errorf("multiple executors found for resource %s", state.Kind())
					}
					result[state.Kind()] = executor
					found = true
				}
			}
		}

		if !found && defaultExecutor != nil {
			result[state.Kind()] = defaultExecutor
		}
	}

	return result, nil
}
